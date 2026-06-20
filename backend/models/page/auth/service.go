package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/email"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/sms"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	identityTypeEmailText = "email"
	identityTypePhoneText = "phone"
)

// Service 编排登录、注册、验证码、找回密码等账号入口流程。
type Service struct {
	cfg          config.Config
	log          *zap.Logger
	store        *data.Store
	cacheStore   cache.Store
	tokenManager *authlib.TokenManager
	emailSender  email.Sender
	smsSender    sms.Sender
}

// NewService 创建 auth page service，并组装数据库、cache、token、邮件和短信发送能力。
func NewService(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Service {
	return &Service{
		cfg:          cfg,
		log:          log,
		store:        data.NewStore(db),
		cacheStore:   cacheStore,
		tokenManager: tokenManager,
		emailSender:  email.NewSender(cfg, log),
		smsSender:    sms.NewSender(cfg, log),
	}
}

// Captcha 生成图片验证码，把答案写入短期 cache，返回给前端验证码 ID 和图片。
func (s *Service) Captcha(ctx context.Context) (*proto.CaptchaResponse, error) {
	id, code, image, err := verification.NewCaptcha()
	if err != nil {
		return nil, internalError(err)
	}
	// 图片验证码只保存小写答案，校验时忽略大小写；TTL 到期后自动失效。
	key := s.cacheStore.Key("captcha", id)
	if err := s.cacheStore.Set(ctx, key, strings.ToLower(code), time.Duration(s.cfg.Verification.CaptchaTTLSeconds)*time.Second); err != nil {
		return nil, internalError(err)
	}
	return &proto.CaptchaResponse{
		CaptchaId:   id,
		ImageBase64: image,
		ExpiresIn:   s.cfg.Verification.CaptchaTTLSeconds,
	}, nil
}

// SendCode 发送短信或邮箱验证码。流程包含目标标准化、图片验证码校验、频控、发送和审计。
func (s *Service) SendCode(ctx context.Context, req *proto.SendCodeRequest, meta pagectx.RequestMeta) (*proto.SendCodeResponse, error) {
	// 先把邮箱/手机号标准化，保证 cache key、唯一索引和审计记录使用同一种格式。
	codeType, target, err := s.normalizeCodeTarget(req.GetCodeType(), req.GetTarget())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	scene := strings.TrimSpace(req.GetScene())
	if !validScene(scene) {
		return nil, apperror.BadRequest(response.CodeBadRequest, "验证码场景不正确")
	}
	if err := s.verifyCaptcha(ctx, req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
		return nil, err
	}
	// 发送验证码前做按目标、按 IP 的多维频控，避免短信/邮件接口被刷。
	if err := s.checkSendRateLimit(ctx, scene, target, security.ClientIPHash(meta.IP)); err != nil {
		return nil, err
	}
	// 邮箱注册场景需要提前判断是否已注册，避免用户走完整流程后才失败。
	if codeType == verification.CodeTypeEmail && scene == verification.SceneRegister && s.store.DBReady() {
		identity, err := s.store.FindIdentity(ctx, dao.IdentityTypeEmail, target)
		if err != nil {
			return nil, internalError(err)
		}
		if identity != nil {
			return nil, apperror.Conflict(response.CodeConflict, "该邮箱已注册，请直接登录或找回密码")
		}
	}

	// 明文验证码只写入 cache；MySQL 审计表只存 hash，避免验证码泄露风险。
	code := verification.NewDigitCode(6)
	codeKey := s.cacheStore.Key("verify-code", scene, target)
	if err := s.cacheStore.Set(ctx, codeKey, code, time.Duration(s.cfg.Verification.CodeTTLSeconds)*time.Second); err != nil {
		return nil, internalError(err)
	}

	provider := ""
	// 邮件和短信走不同 sender，但对 page 层统一成“发送 6 位验证码”的业务动作。
	if codeType == verification.CodeTypeEmail {
		provider = s.emailSender.Name()
		err = s.emailSender.SendCode(ctx, target, code, scene)
	} else {
		provider, err = s.smsSender.SendCode(ctx, target, code, scene)
	}
	if err != nil {
		// 发送失败时删除 cache 中的验证码，避免用户收到失败提示却仍能用该验证码通过校验。
		_ = s.cacheStore.Del(ctx, codeKey)
		return nil, internalError(err)
	}

	s.auditVerificationCode(ctx, codeType, scene, target, code, provider, meta)
	return &proto.SendCodeResponse{
		CooldownSeconds: s.cfg.Verification.CodeResendCooldownSeconds,
		ExpiresIn:       s.cfg.Verification.CodeTTLSeconds,
	}, nil
}

// EmailLogin 使用邮箱密码登录。连续失败超过阈值后要求图片验证码，并统一返回模糊错误。
func (s *Service) EmailLogin(ctx context.Context, req *proto.EmailLoginRequest, meta pagectx.RequestMeta) (*proto.LoginResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	emailValue, err := security.NormalizeEmail(req.GetEmail())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	failKey := s.loginFailKey(emailValue, meta.IP)
	failCount := s.currentFailCount(ctx, failKey)
	// 失败次数达到阈值后启用图片验证码；验证码缺失时通过 data 告诉前端展示验证码。
	if failCount >= s.cfg.Verification.LoginCaptchaFailThreshold {
		if err := s.verifyCaptcha(ctx, req.GetCaptchaId(), req.GetCaptchaCode()); err != nil {
			return nil, apperror.WithData(http.StatusBadRequest, response.CodeBadRequest, "需要完成图片验证码后再登录", map[string]bool{"captchaRequired": true})
		}
	}

	// 先查身份表再查用户主表，邮箱不直接耦合在 users 表上。
	identity, err := s.store.FindIdentity(ctx, dao.IdentityTypeEmail, emailValue)
	if err != nil {
		return nil, internalError(err)
	}
	if identity == nil {
		// 不暴露邮箱是否存在，避免账号枚举；同时记录失败计数。
		s.recordLoginFailure(ctx, failKey, dao.IdentityTypeEmail, emailValue, meta)
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "账号或密码不正确")
	}
	user, err := s.store.FindUser(ctx, identity.UserID)
	if err != nil {
		return nil, internalError(err)
	}
	if user == nil || user.Status != dao.UserStatusNormal || !authlib.CheckPassword(user.PasswordHash, req.GetPassword()) {
		s.recordLoginFailure(ctx, failKey, dao.IdentityTypeEmail, emailValue, meta)
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "账号或密码不正确")
	}
	// 登录成功后清理失败计数，并签发新的设备会话。
	_ = s.cacheStore.Del(ctx, failKey)
	s.logSecurity(ctx, user.ID, "email_login", map[string]string{"email": security.MaskEmail(emailValue)}, meta)
	return s.issueLogin(ctx, user, meta)
}

// PhoneLogin 使用手机号验证码登录注册一体。已绑定手机号直接登录，新手机号返回资料完善 token。
func (s *Service) PhoneLogin(ctx context.Context, req *proto.PhoneLoginRequest, meta pagectx.RequestMeta) (*proto.PhoneLoginResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	phone, err := security.NormalizeMainlandPhone(req.GetPhone())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := s.verifyCode(ctx, verification.CodeTypeSMS, verification.SceneLogin, phone, req.GetSmsCode()); err != nil {
		return nil, err
	}
	// 手机号验证码已经证明联系方式可用；若身份不存在，就先进入资料完善而不是直接创建空资料账号。
	identity, err := s.store.FindIdentity(ctx, dao.IdentityTypePhone, phone)
	if err != nil {
		return nil, internalError(err)
	}
	if identity == nil {
		token, err := s.createProfileSetupToken(ctx, verification.SetupPayload{Kind: identityTypePhoneText, Phone: phone})
		if err != nil {
			return nil, err
		}
		return &proto.PhoneLoginResponse{RequiresProfileSetup: true, ProfileSetupToken: token}, nil
	}
	user, err := s.store.FindUser(ctx, identity.UserID)
	if err != nil {
		return nil, internalError(err)
	}
	if user == nil || user.Status != dao.UserStatusNormal {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "账号状态不可用")
	}
	authResp, err := s.issueLogin(ctx, user, meta)
	if err != nil {
		return nil, err
	}
	s.logSecurity(ctx, user.ID, "phone_login", map[string]string{"phone": security.MaskPhone(phone)}, meta)
	return &proto.PhoneLoginResponse{Auth: authResp}, nil
}

// EmailRegisterVerify 校验邮箱注册验证码和密码规则，成功后签发资料完善 token。
func (s *Service) EmailRegisterVerify(ctx context.Context, req *proto.EmailRegisterVerifyRequest, meta pagectx.RequestMeta) (*proto.ProfileSetupResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	emailValue, err := security.NormalizeEmail(req.GetEmail())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := security.ValidatePassword(req.GetPassword(), emailValue); err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	// 邮箱注册分两步：这里仅做验证和密码哈希，不创建用户，避免缺失画像数据的账号落库。
	identity, err := s.store.FindIdentity(ctx, dao.IdentityTypeEmail, emailValue)
	if err != nil {
		return nil, internalError(err)
	}
	if identity != nil {
		return nil, apperror.Conflict(response.CodeConflict, "该邮箱已注册，请直接登录或找回密码")
	}
	if err := s.verifyCode(ctx, verification.CodeTypeEmail, verification.SceneRegister, emailValue, req.GetEmailCode()); err != nil {
		return nil, err
	}
	passwordHash, err := authlib.HashPassword(req.GetPassword())
	if err != nil {
		return nil, internalError(err)
	}
	token, err := s.createProfileSetupToken(ctx, verification.SetupPayload{Kind: identityTypeEmailText, Email: emailValue, PasswordHash: passwordHash})
	if err != nil {
		return nil, err
	}
	s.log.Debug("邮箱注册验证码校验完成，进入资料完善", zap.String("email", security.MaskEmail(emailValue)), zap.String("ipHash", security.ClientIPHash(meta.IP)))
	return &proto.ProfileSetupResponse{ProfileSetupToken: token}, nil
}

// CompleteProfile 使用资料完善 token 创建正式账号，并在创建成功后直接登录。
func (s *Service) CompleteProfile(ctx context.Context, req *proto.CompleteProfileRequest, meta pagectx.RequestMeta) (*proto.LoginResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	nickname, err := validateNickname(req.GetNickname())
	if err != nil {
		return nil, err
	}
	if err := validateProfileFields(req.GetPetStage(), req.GetInterestedPetTypes()); err != nil {
		return nil, err
	}
	setupKey := s.cacheStore.Key("profile-setup", req.GetProfileSetupToken())
	payload, err := verification.LoadJSON[verification.SetupPayload](ctx, s.cacheStore, setupKey)
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, "资料完善流程已过期，请重新开始")
	}
	// token payload 里只保存通过验证的联系方式和密码哈希，避免前端篡改联系方式类型。
	var identityType int32
	var identityValue string
	if payload.Kind == identityTypeEmailText {
		identityType = dao.IdentityTypeEmail
		identityValue = payload.Email
	} else if payload.Kind == identityTypePhoneText {
		identityType = dao.IdentityTypePhone
		identityValue = payload.Phone
	} else {
		return nil, apperror.BadRequest(response.CodeBadRequest, "资料完善流程不正确")
	}
	existing, err := s.store.FindIdentity(ctx, identityType, identityValue)
	if err != nil {
		return nil, internalError(err)
	}
	if existing != nil {
		return nil, apperror.Conflict(response.CodeConflict, "该联系方式已被注册，请直接登录")
	}
	now := data.Now()
	// users、user_auth_identities、user_profiles 必须在同一事务创建，任何一步失败都回滚。
	user := &dao.User{
		Nickname:           nickname,
		PasswordHash:       payload.PasswordHash,
		PasswordUpdateTime: passwordUpdateTime(payload.PasswordHash, now),
		BaseModel:          dao.NewBaseModel(dao.UserStatusNormal, now),
	}
	identity := &dao.UserAuthIdentity{
		IdentityType:  identityType,
		IdentityValue: data.StringPtr(identityValue),
		VerifyTime:    now,
		BaseModel:     dao.NewBaseModel(dao.IdentityStatusNormal, now),
	}
	profile := buildProfile(0, req.GetPetStage(), req.GetInterestedPetTypes(), req.GetPetExperience(), req.GetDailyCompanyTime(), req.GetLivingSituation(), req.GetPetConstraints(), now)
	if err := s.store.CreateUserWithIdentityAndProfile(ctx, user, identity, profile); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperror.Conflict(response.CodeConflict, "该联系方式已被注册，请直接登录")
		}
		return nil, internalError(err)
	}
	// 创建成功后立刻删除资料完善 token，避免同一 token 重复创建账号。
	_ = s.cacheStore.Del(ctx, setupKey)
	s.logSecurity(ctx, user.ID, "complete_profile", map[string]string{"identityType": payload.Kind}, meta)
	return s.issueLogin(ctx, user, meta)
}

// Refresh 使用 refresh token 查找会话，轮换 refresh token 并签发新的 access token。
func (s *Service) Refresh(ctx context.Context, req *proto.RefreshRequest, meta pagectx.RequestMeta) (*proto.TokenPair, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	// refresh token 明文只在客户端保存，服务端用摘要查找会话。
	refreshHash := s.tokenManager.HashRefreshToken(strings.TrimSpace(req.GetRefreshToken()))
	session, err := s.store.FindSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return nil, internalError(err)
	}
	now := data.Now()
	if session == nil || session.ExpireTime <= now || session.Status != dao.SessionStatusNormal {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效，请重新登录")
	}
	// 每次刷新都轮换 refresh token，降低旧 token 泄露后的可复用窗口。
	refreshToken, newRefreshHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, internalError(err)
	}
	expireTime := now + int64(s.tokenManager.RefreshTTL().Seconds())
	if err := s.store.UpdateSessionRefreshToken(ctx, session.ID, newRefreshHash, now, expireTime); err != nil {
		return nil, internalError(err)
	}
	if err := authlib.SetCachedSession(ctx, s.cacheStore, session.ID, authlib.CachedSession{UserID: session.UserID, RefreshTokenHash: newRefreshHash, ExpireTime: expireTime}); err != nil {
		return nil, internalError(err)
	}
	accessToken, accessTTL, err := s.tokenManager.GenerateAccessToken(session.UserID, session.ID)
	if err != nil {
		return nil, internalError(err)
	}
	s.logSecurity(ctx, session.UserID, "refresh_token", map[string]string{"sessionId": strconv.FormatUint(session.ID, 10)}, meta)
	return &proto.TokenPair{
		AccessToken:           accessToken,
		AccessTokenExpiresIn:  accessTTL,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresIn: int64(s.tokenManager.RefreshTTL().Seconds()),
	}, nil
}

// Logout 撤销 refresh token 对应的当前设备会话，并删除 session cache。
func (s *Service) Logout(ctx context.Context, req *proto.LogoutRequest) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	refreshHash := s.tokenManager.HashRefreshToken(strings.TrimSpace(req.GetRefreshToken()))
	session, err := s.store.FindSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return internalError(err)
	}
	if session == nil {
		return nil
	}
	now := data.Now()
	if err := s.store.RevokeSession(ctx, session.ID, 0, dao.SessionStatusLogout, "logout", now); err != nil {
		return internalError(err)
	}
	_ = authlib.DeleteCachedSession(ctx, s.cacheStore, session.ID)
	return nil
}

// ForgotPasswordVerifyCode 校验找回密码验证码，账号存在时签发短期 reset token。
func (s *Service) ForgotPasswordVerifyCode(ctx context.Context, req *proto.ForgotPasswordVerifyCodeRequest) (*proto.ForgotPasswordVerifyCodeResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	identityType, target, err := s.normalizeIdentity(req.GetIdentityType(), req.GetIdentityValue())
	if err != nil {
		return nil, apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	codeType := verification.CodeTypeEmail
	if identityType == dao.IdentityTypePhone {
		codeType = verification.CodeTypeSMS
	}
	if err := s.verifyCode(ctx, codeType, verification.SceneForgotPassword, target, req.GetCode()); err != nil {
		return nil, err
	}
	identity, err := s.store.FindIdentity(ctx, identityType, target)
	if err != nil {
		return nil, internalError(err)
	}
	if identity == nil {
		// 找回密码不暴露账号是否存在，避免邮箱/手机号被枚举。
		return &proto.ForgotPasswordVerifyCodeResponse{}, nil
	}
	token, err := security.RandomToken(32)
	if err != nil {
		return nil, internalError(err)
	}
	key := s.cacheStore.Key("password-reset", token)
	if err := verification.StoreJSON(ctx, s.cacheStore, key, verification.ResetPayload{UserID: identity.UserID}, time.Duration(s.cfg.Verification.PasswordResetTTLSeconds)*time.Second); err != nil {
		return nil, internalError(err)
	}
	return &proto.ForgotPasswordVerifyCodeResponse{ResetToken: token}, nil
}

// ForgotPasswordReset 使用 reset token 设置新密码，并撤销该用户所有历史会话。
func (s *Service) ForgotPasswordReset(ctx context.Context, req *proto.ForgotPasswordResetRequest, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	key := s.cacheStore.Key("password-reset", req.GetResetToken())
	payload, err := verification.LoadJSON[verification.ResetPayload](ctx, s.cacheStore, key)
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, "重置密码流程已过期，请重新开始")
	}
	user, err := s.store.FindUser(ctx, payload.UserID)
	if err != nil {
		return internalError(err)
	}
	if user == nil {
		return apperror.BadRequest(response.CodeBadRequest, "重置密码流程无效")
	}
	if err := security.ValidatePassword(req.GetNewPassword()); err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	passwordHash, err := authlib.HashPassword(req.GetNewPassword())
	if err != nil {
		return internalError(err)
	}
	now := data.Now()
	// 重置密码属于高风险操作，成功后全部 refresh token 失效，用户需要重新登录。
	sessions, _ := s.store.ListSessions(ctx, user.ID)
	if err := s.store.UpdatePassword(ctx, user.ID, passwordHash, now); err != nil {
		return internalError(err)
	}
	if err := s.store.RevokeUserSessions(ctx, user.ID, 0, dao.SessionStatusRevoked, "password_reset", now); err != nil {
		return internalError(err)
	}
	for _, session := range sessions {
		_ = authlib.DeleteCachedSession(ctx, s.cacheStore, session.ID)
	}
	_ = s.cacheStore.Del(ctx, key)
	s.logSecurity(ctx, user.ID, "password_reset", map[string]string{}, meta)
	return nil
}

// createProfileSetupToken 生成资料完善短期 token，把已验证的联系方式和密码哈希放入 cache。
func (s *Service) createProfileSetupToken(ctx context.Context, payload verification.SetupPayload) (string, error) {
	token, err := security.RandomToken(32)
	if err != nil {
		return "", internalError(err)
	}
	key := s.cacheStore.Key("profile-setup", token)
	if err := verification.StoreJSON(ctx, s.cacheStore, key, payload, time.Duration(s.cfg.Verification.ProfileSetupTTLSeconds)*time.Second); err != nil {
		return "", internalError(err)
	}
	return token, nil
}

// issueLogin 创建设备会话、写入 session cache，并返回 access token 与 refresh token。
func (s *Service) issueLogin(ctx context.Context, user *dao.User, meta pagectx.RequestMeta) (*proto.LoginResponse, error) {
	now := data.Now()
	refreshToken, refreshHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, internalError(err)
	}
	expireTime := now + int64(s.tokenManager.RefreshTTL().Seconds())
	// MySQL 作为会话事实源，用于账号安全页展示设备和审计会话状态变化。
	session := &dao.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		DeviceName:       deviceName(meta.UserAgent),
		UserAgentHash:    security.HashPlain(meta.UserAgent),
		IPHash:           security.ClientIPHash(meta.IP),
		LastActiveTime:   now,
		ExpireTime:       expireTime,
		BaseModel:        dao.NewBaseModel(dao.SessionStatusNormal, now),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, internalError(err)
	}
	// Redis/cache 作为鉴权热路径，避免每个 access token 请求都回源 MySQL。
	if err := authlib.SetCachedSession(ctx, s.cacheStore, session.ID, authlib.CachedSession{UserID: user.ID, RefreshTokenHash: refreshHash, ExpireTime: expireTime}); err != nil {
		return nil, internalError(err)
	}
	accessToken, accessTTL, err := s.tokenManager.GenerateAccessToken(user.ID, session.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return &proto.LoginResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresIn:  accessTTL,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresIn: int64(s.tokenManager.RefreshTTL().Seconds()),
		User:                  toUserInfo(user),
	}, nil
}

// normalizeCodeTarget 将验证码发送目标统一转换为系统内部格式，手机号当前只支持中国大陆 +86。
func (s *Service) normalizeCodeTarget(codeTypeText string, target string) (int32, string, error) {
	switch strings.ToLower(strings.TrimSpace(codeTypeText)) {
	case "sms":
		value, err := security.NormalizeMainlandPhone(target)
		return verification.CodeTypeSMS, value, err
	case "email":
		value, err := security.NormalizeEmail(target)
		return verification.CodeTypeEmail, value, err
	default:
		return 0, "", errors.New("验证码类型不正确")
	}
}

// normalizeIdentity 将找回密码入参中的账号类型和值转换成身份表枚举和值。
func (s *Service) normalizeIdentity(identityTypeText string, value string) (int32, string, error) {
	switch strings.ToLower(strings.TrimSpace(identityTypeText)) {
	case identityTypeEmailText:
		normalized, err := security.NormalizeEmail(value)
		return dao.IdentityTypeEmail, normalized, err
	case identityTypePhoneText:
		normalized, err := security.NormalizeMainlandPhone(value)
		return dao.IdentityTypePhone, normalized, err
	default:
		return 0, "", errors.New("账号类型不正确")
	}
}

// verifyCaptcha 校验图片验证码；成功后立即删除，保证同一验证码只能使用一次。
func (s *Service) verifyCaptcha(ctx context.Context, captchaID string, captchaCode string) error {
	captchaID = strings.TrimSpace(captchaID)
	captchaCode = strings.TrimSpace(captchaCode)
	if captchaID == "" || captchaCode == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请输入图片验证码")
	}
	key := s.cacheStore.Key("captcha", captchaID)
	expected, err := s.cacheStore.Get(ctx, key)
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, "图片验证码已过期，请刷新后重试")
	}
	if !strings.EqualFold(expected, captchaCode) {
		return apperror.BadRequest(response.CodeBadRequest, "图片验证码不正确")
	}
	// 图片验证码用于防刷，不允许重复提交。
	_ = s.cacheStore.Del(ctx, key)
	return nil
}

// verifyCode 校验短信/邮箱验证码；成功后删除 cache，并把 MySQL 审计记录标记为已使用。
func (s *Service) verifyCode(ctx context.Context, codeType int32, scene string, target string, code string) error {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return apperror.BadRequest(response.CodeBadRequest, "请输入 6 位验证码")
	}
	key := s.cacheStore.Key("verify-code", scene, target)
	expected, err := s.cacheStore.Get(ctx, key)
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, "验证码已过期，请重新获取")
	}
	if expected != code {
		return apperror.BadRequest(response.CodeBadRequest, "验证码不正确")
	}
	// 验证码成功使用后立即失效，符合“同一验证码只能成功使用一次”的规则。
	_ = s.cacheStore.Del(ctx, key)
	if s.store.DBReady() {
		_ = s.store.MarkVerificationCodeUsed(ctx, codeType, scene, target, s.tokenManager.HashCode(code), data.Now())
	}
	return nil
}

// checkSendRateLimit 叠加目标小时/天级限制和 IP 小时级限制，保护短信与邮箱发送资源。
func (s *Service) checkSendRateLimit(ctx context.Context, scene string, target string, ipHash string) error {
	hourTTL := time.Hour
	dayTTL := 24 * time.Hour
	targetHour, err := s.cacheStore.IncrWithTTL(ctx, s.cacheStore.Key("rate", "code", "hour", scene, target), hourTTL)
	if err != nil {
		return internalError(err)
	}
	targetDay, err := s.cacheStore.IncrWithTTL(ctx, s.cacheStore.Key("rate", "code", "day", scene, target, time.Now().Format("20060102")), dayTTL)
	if err != nil {
		return internalError(err)
	}
	ipHour, err := s.cacheStore.IncrWithTTL(ctx, s.cacheStore.Key("rate", "ip", "hour", scene, ipHash), hourTTL)
	if err != nil {
		return internalError(err)
	}
	if targetHour > s.cfg.Verification.CodeTargetHourlyLimit || targetDay > s.cfg.Verification.CodeTargetDailyLimit || ipHour > s.cfg.Verification.CodeIPHourlyLimit {
		return apperror.RateLimited(response.CodeRateLimited, "验证码发送过于频繁，请稍后再试")
	}
	return nil
}

// auditVerificationCode 写入验证码发送审计。cache 是校验事实源，MySQL 是安全审计来源。
func (s *Service) auditVerificationCode(ctx context.Context, codeType int32, scene string, target string, code string, provider string, meta pagectx.RequestMeta) {
	if !s.store.DBReady() {
		return
	}
	now := data.Now()
	if err := s.store.CreateVerificationCode(ctx, &dao.VerificationCode{
		CodeType:      codeType,
		Scene:         scene,
		Target:        target,
		CodeHash:      s.tokenManager.HashCode(code),
		ExpireTime:    now + s.cfg.Verification.CodeTTLSeconds,
		IPHash:        security.ClientIPHash(meta.IP),
		UserAgentHash: security.HashPlain(meta.UserAgent),
		Provider:      provider,
		BaseModel:     dao.NewBaseModel(dao.VerificationStatusUnused, now),
	}); err != nil {
		s.log.Error("验证码审计记录写入失败", zap.Error(err))
	}
}

// loginFailKey 生成邮箱密码登录失败计数 key，按账号和 IP 共同计数。
func (s *Service) loginFailKey(identity string, ip string) string {
	return s.cacheStore.Key("login-fail", identityTypeEmailText, identity, security.ClientIPHash(ip))
}

// currentFailCount 读取登录失败次数。cache 不存在时表示还未触发风控。
func (s *Service) currentFailCount(ctx context.Context, key string) int64 {
	raw, err := s.cacheStore.Get(ctx, key)
	if err != nil {
		return 0
	}
	count, _ := strconv.ParseInt(raw, 10, 64)
	return count
}

// recordLoginFailure 同步记录 Redis 失败计数和 MySQL 审计记录。
func (s *Service) recordLoginFailure(ctx context.Context, key string, identityType int32, identityValue string, meta pagectx.RequestMeta) {
	count, err := s.cacheStore.IncrWithTTL(ctx, key, 15*time.Minute)
	if err != nil {
		s.log.Error("登录失败计数写入 cache 失败", zap.Error(err))
	}
	if s.store.DBReady() {
		now := data.Now()
		_ = s.store.UpsertLoginAttempt(ctx, &dao.LoginAttempt{
			IdentityType:  identityType,
			IdentityValue: identityValue,
			IPHash:        security.ClientIPHash(meta.IP),
			FailCount:     int32(count),
			LastFailTime:  now,
			BaseModel:     dao.NewBaseModel(dao.LoginAttemptStatusNormal, now),
		})
	}
}

// requireDB 检查账号接口是否具备 MySQL。没有数据库时只允许健康检查和验证码等非账号事实写入能力。
func (s *Service) requireDB() error {
	if s.store.DBReady() {
		return nil
	}
	return apperror.New(http.StatusServiceUnavailable, response.CodeInternalError, "账号服务需要配置 MySQL 后才能使用")
}

// logSecurity 记录安全事件，detail 必须使用脱敏后的摘要，不能写入密码、验证码或 token。
func (s *Service) logSecurity(ctx context.Context, userID uint64, eventType string, detail map[string]string, meta pagectx.RequestMeta) {
	if !s.store.DBReady() {
		return
	}
	raw, _ := json.Marshal(detail)
	now := data.Now()
	if err := s.store.LogSecurity(ctx, &dao.SecurityLog{
		UserID:        userID,
		EventType:     eventType,
		Detail:        string(raw),
		IPHash:        security.ClientIPHash(meta.IP),
		UserAgentHash: security.HashPlain(meta.UserAgent),
		BaseModel:     dao.NewBaseModel(dao.SecurityLogStatusNormal, now),
	}); err != nil {
		s.log.Error("安全日志写入失败", zap.Error(err))
	}
}

// validScene 校验验证码场景，避免前端传入任意 scene 污染频控 key 和审计数据。
func validScene(scene string) bool {
	switch scene {
	case verification.SceneLogin, verification.SceneRegister, verification.SceneForgotPassword, verification.SceneBindPhone, verification.SceneBindEmail:
		return true
	default:
		return false
	}
}

// validateNickname 校验昵称长度。昵称允许重复，系统内部以用户 ID 区分用户。
func validateNickname(nickname string) (string, error) {
	value := strings.TrimSpace(nickname)
	count := utf8.RuneCountInString(value)
	if count < 2 || count > 20 {
		return "", apperror.BadRequest(response.CodeBadRequest, "昵称需为 2-20 个字符")
	}
	return value, nil
}

// validateProfileFields 校验资料完善页必填项：养宠阶段和至少一种感兴趣宠物。
func validateProfileFields(petStage string, interestedPetTypes []string) error {
	if strings.TrimSpace(petStage) == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请选择当前养宠阶段")
	}
	if len(interestedPetTypes) == 0 {
		return apperror.BadRequest(response.CodeBadRequest, "请至少选择一种感兴趣的宠物")
	}
	return nil
}

// buildProfile 把前端资料完善/资料编辑入参转换为 user_profiles DAO。
func buildProfile(userID uint64, petStage string, interestedPetTypes []string, petExperience string, dailyCompanyTime string, livingSituation string, petConstraints []string, now int64) *dao.UserProfile {
	return &dao.UserProfile{
		UserID:             userID,
		PetStage:           strings.TrimSpace(petStage),
		InterestedPetTypes: data.JSONStrings(interestedPetTypes),
		PetExperience:      strings.TrimSpace(petExperience),
		DailyCompanyTime:   strings.TrimSpace(dailyCompanyTime),
		LivingSituation:    strings.TrimSpace(livingSituation),
		PetConstraints:     data.JSONStrings(petConstraints),
		BaseModel:          dao.NewBaseModel(dao.ProfileStatusNormal, now),
	}
}

// passwordUpdateTime 根据是否存在密码哈希决定密码更新时间。手机号验证码注册时密码可以为空。
func passwordUpdateTime(passwordHash string, now int64) int64 {
	if passwordHash == "" {
		return 0
	}
	return now
}

// deviceName 从 User-Agent 粗略提取浏览器和系统，用于账号安全页展示登录设备。
func deviceName(userAgent string) string {
	lower := strings.ToLower(userAgent)
	browser := "浏览器"
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "chrome/"):
		browser = "Chrome"
	case strings.Contains(lower, "safari/"):
		browser = "Safari"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	}
	osName := "未知设备"
	switch {
	case strings.Contains(lower, "mac os"):
		osName = "macOS"
	case strings.Contains(lower, "windows"):
		osName = "Windows"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		osName = "iOS"
	case strings.Contains(lower, "android"):
		osName = "Android"
	}
	return fmt.Sprintf("%s · %s", browser, osName)
}

// toUserInfo 将 users DAO 转成接口返回结构，避免把 password_hash 等字段暴露给前端。
func toUserInfo(user *dao.User) *proto.UserInfo {
	if user == nil {
		return nil
	}
	return &proto.UserInfo{
		Id:       int64(user.ID),
		Nickname: user.Nickname,
		Status:   user.Status,
	}
}

// internalError 把底层错误包装成统一业务错误，controller 会负责转换成 HTTP 响应。
func internalError(err error) *apperror.Error {
	return apperror.Wrap(http.StatusInternalServerError, response.CodeInternalError, "服务暂时不可用", err)
}
