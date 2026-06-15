package account

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// Service 编排账号资料、安全设置和设备管理。
type Service struct {
	cfg          config.Config
	log          *zap.Logger
	store        *data.Store
	cacheStore   cache.Store
	tokenManager *authlib.TokenManager
}

// NewService 创建账号中心 page service，集中注入 DB、缓存和 token 管理器。
func NewService(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Service {
	return &Service{
		cfg:          cfg,
		log:          log,
		store:        data.NewStore(db),
		cacheStore:   cacheStore,
		tokenManager: tokenManager,
	}
}

// Me 查询当前登录用户的账号资料、扩展画像和已绑定联系方式。
func (s *Service) Me(ctx context.Context, userID uint64) (*proto.AccountMeResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	if user == nil {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	profile, err := s.store.FindProfile(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	identities, err := s.store.ListIdentities(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	return &proto.AccountMeResponse{
		User:       toUserInfo(user),
		Profile:    toProfile(profile),
		Identities: toIdentities(identities),
	}, nil
}

// UpdateProfile 更新昵称和养宠画像，供注册后补充资料以及个人资料页复用。
func (s *Service) UpdateProfile(ctx context.Context, userID uint64, req *proto.UpdateProfileRequest, meta pagectx.RequestMeta) (*proto.AccountMeResponse, error) {
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
	now := data.Now()
	// 昵称在 users 表，画像在 user_profiles 表，使用同一个事务保证资料更新一致。
	err = s.store.WithTx(ctx, func(tx *data.Store) error {
		user, err := tx.FindUser(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return gorm.ErrRecordNotFound
		}
		user.Nickname = nickname
		user.UpdateTime = now
		if err := tx.UpdateUserProfileFields(ctx, user, buildProfile(userID, req.GetPetStage(), req.GetInterestedPetTypes(), req.GetPetExperience(), req.GetDailyCompanyTime(), req.GetLivingSituation(), req.GetPetConstraints(), now)); err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if err != nil {
		return nil, internalError(err)
	}
	s.logSecurity(ctx, userID, "update_profile", map[string]string{}, meta)
	return s.Me(ctx, userID)
}

// UpdateEmail 绑定或换绑邮箱，验证码校验成功后才更新身份表。
func (s *Service) UpdateEmail(ctx context.Context, userID uint64, req *proto.UpdateEmailRequest, meta pagectx.RequestMeta) error {
	emailValue, err := security.NormalizeEmail(req.GetEmail())
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := s.verifyCode(ctx, verification.CodeTypeEmail, verification.SceneBindEmail, emailValue, req.GetEmailCode()); err != nil {
		return err
	}
	return s.updateIdentity(ctx, userID, dao.IdentityTypeEmail, emailValue, meta)
}

// UpdatePhone 绑定或换绑中国大陆手机号，当前版本只接受 +86 手机号。
func (s *Service) UpdatePhone(ctx context.Context, userID uint64, req *proto.UpdatePhoneRequest, meta pagectx.RequestMeta) error {
	phone, err := security.NormalizeMainlandPhone(req.GetPhone())
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := s.verifyCode(ctx, verification.CodeTypeSMS, verification.SceneBindPhone, phone, req.GetSmsCode()); err != nil {
		return err
	}
	return s.updateIdentity(ctx, userID, dao.IdentityTypePhone, phone, meta)
}

// UnbindEmail 解绑邮箱，底层会校验当前密码并确保账号仍保留至少一种联系方式。
func (s *Service) UnbindEmail(ctx context.Context, userID uint64, currentPassword string, meta pagectx.RequestMeta) error {
	return s.unbindIdentity(ctx, userID, dao.IdentityTypeEmail, currentPassword, meta)
}

// UnbindPhone 解绑手机号，底层会校验当前密码并确保账号仍保留至少一种联系方式。
func (s *Service) UnbindPhone(ctx context.Context, userID uint64, currentPassword string, meta pagectx.RequestMeta) error {
	return s.unbindIdentity(ctx, userID, dao.IdentityTypePhone, currentPassword, meta)
}

// UpdatePassword 修改登录密码，并吊销除当前设备外的其他登录会话。
func (s *Service) UpdatePassword(ctx context.Context, userID uint64, currentSessionID uint64, req *proto.UpdatePasswordRequest, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return internalError(err)
	}
	if user == nil {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if user.PasswordHash != "" && !authlib.CheckPassword(user.PasswordHash, req.GetCurrentPassword()) {
		return apperror.BadRequest(response.CodeBadRequest, "当前密码不正确")
	}
	if err := security.ValidatePassword(req.GetNewPassword()); err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	passwordHash, err := authlib.HashPassword(req.GetNewPassword())
	if err != nil {
		return internalError(err)
	}
	now := data.Now()
	// 先读取会话列表，用于 DB 状态变更后同步清理 Redis 中的会话缓存。
	sessions, _ := s.store.ListSessions(ctx, userID)
	if err := s.store.UpdatePassword(ctx, userID, passwordHash, now); err != nil {
		return internalError(err)
	}
	// 改密后保留当前会话，避免用户在当前设备上被立即踢出。
	if err := s.store.RevokeUserSessions(ctx, userID, currentSessionID, dao.SessionStatusRevoked, "password_changed", now); err != nil {
		return internalError(err)
	}
	for _, session := range sessions {
		if session.ID != currentSessionID {
			_ = authlib.DeleteCachedSession(ctx, s.cacheStore, session.ID)
		}
	}
	s.logSecurity(ctx, userID, "update_password", map[string]string{}, meta)
	return nil
}

// Sessions 返回当前账号的登录设备列表，并标记当前 access token 对应的会话。
func (s *Service) Sessions(ctx context.Context, userID uint64, currentSessionID uint64) (*proto.SessionsResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	sessions, err := s.store.ListSessions(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	resp := &proto.SessionsResponse{CurrentSessionId: int64(currentSessionID)}
	for _, session := range sessions {
		resp.Sessions = append(resp.Sessions, &proto.UserSession{
			Id:             int64(session.ID),
			DeviceName:     session.DeviceName,
			Status:         session.Status,
			LastActiveTime: session.LastActiveTime,
			ExpireTime:     session.ExpireTime,
			Current:        session.ID == currentSessionID,
		})
	}
	return resp, nil
}

// RevokeSession 退出指定设备，当前设备需要走 logout 保持语义清晰。
func (s *Service) RevokeSession(ctx context.Context, userID uint64, currentSessionID uint64, sessionID uint64, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	if sessionID == currentSessionID {
		return apperror.BadRequest(response.CodeBadRequest, "当前设备请使用退出登录")
	}
	now := data.Now()
	if err := s.store.RevokeSession(ctx, sessionID, userID, dao.SessionStatusRevoked, "user_revoke", now); err != nil {
		return internalError(err)
	}
	_ = authlib.DeleteCachedSession(ctx, s.cacheStore, sessionID)
	s.logSecurity(ctx, userID, "revoke_session", map[string]string{"sessionId": strconv.FormatUint(sessionID, 10)}, meta)
	return nil
}

// RevokeAllSessions 校验当前密码后退出全部设备。
func (s *Service) RevokeAllSessions(ctx context.Context, userID uint64, currentPassword string, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return internalError(err)
	}
	if user == nil {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if user.PasswordHash == "" || !authlib.CheckPassword(user.PasswordHash, currentPassword) {
		return apperror.BadRequest(response.CodeBadRequest, "当前密码不正确")
	}
	sessions, _ := s.store.ListSessions(ctx, userID)
	now := data.Now()
	if err := s.store.RevokeUserSessions(ctx, userID, 0, dao.SessionStatusRevoked, "revoke_all", now); err != nil {
		return internalError(err)
	}
	for _, session := range sessions {
		_ = authlib.DeleteCachedSession(ctx, s.cacheStore, session.ID)
	}
	s.logSecurity(ctx, userID, "revoke_all_sessions", map[string]string{}, meta)
	return nil
}

// updateIdentity 绑定或换绑联系方式；同类型旧联系方式会先解绑，再写入新联系方式。
func (s *Service) updateIdentity(ctx context.Context, userID uint64, identityType int32, identityValue string, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	now := data.Now()
	err := s.store.WithTx(ctx, func(tx *data.Store) error {
		// 一个邮箱或手机号只能绑定到一个账号，避免登录入口产生歧义。
		existing, err := tx.FindIdentity(ctx, identityType, identityValue)
		if err != nil {
			return err
		}
		if existing != nil && existing.UserID != userID {
			return gorm.ErrDuplicatedKey
		}
		// 先解绑同类型旧身份，再 upsert 新身份，保证账号侧同类型联系方式只有一个有效值。
		if err := tx.UnbindIdentity(ctx, userID, identityType, now); err != nil {
			return err
		}
		return tx.UpsertIdentity(ctx, &dao.UserAuthIdentity{
			UserID:        userID,
			IdentityType:  identityType,
			IdentityValue: data.StringPtr(identityValue),
			VerifyTime:    now,
			BaseModel:     dao.NewBaseModel(dao.IdentityStatusNormal, now),
		})
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperror.Conflict(response.CodeConflict, "该联系方式已被其他账号绑定")
	}
	if err != nil {
		return internalError(err)
	}
	eventType := "bind_phone"
	masked := security.MaskPhone(identityValue)
	if identityType == dao.IdentityTypeEmail {
		eventType = "bind_email"
		masked = security.MaskEmail(identityValue)
	}
	s.logSecurity(ctx, userID, eventType, map[string]string{"value": masked}, meta)
	return nil
}

// unbindIdentity 解绑联系方式，必须通过当前密码校验并保留至少一种登录联系方式。
func (s *Service) unbindIdentity(ctx context.Context, userID uint64, identityType int32, currentPassword string, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return internalError(err)
	}
	if user == nil {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if user.PasswordHash == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请先设置密码后再解绑联系方式")
	}
	if !authlib.CheckPassword(user.PasswordHash, currentPassword) {
		return apperror.BadRequest(response.CodeBadRequest, "当前密码不正确")
	}
	// 邮箱和手机号至少保留一个，否则账号会失去可找回、可登录的联系方式。
	count, err := s.store.CountActiveIdentities(ctx, userID)
	if err != nil {
		return internalError(err)
	}
	if count <= 1 {
		return apperror.BadRequest(response.CodeBadRequest, "账号至少需要保留一种联系方式")
	}
	now := data.Now()
	if err := s.store.UnbindIdentity(ctx, userID, identityType, now); err != nil {
		return internalError(err)
	}
	eventType := "unbind_phone"
	if identityType == dao.IdentityTypeEmail {
		eventType = "unbind_email"
	}
	s.logSecurity(ctx, userID, eventType, map[string]string{}, meta)
	return nil
}

// verifyCode 校验绑定场景验证码，并在成功后立即删除缓存中的验证码。
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
	// 验证码成功使用后立即失效，防止同一验证码被重复绑定。
	_ = s.cacheStore.Del(ctx, key)
	if s.store.DBReady() {
		_ = s.store.MarkVerificationCodeUsed(ctx, codeType, scene, target, s.tokenManager.HashCode(code), data.Now())
	}
	return nil
}

// requireDB 统一拦截未配置 MySQL 的账号中心请求。
func (s *Service) requireDB() error {
	if s.store.DBReady() {
		return nil
	}
	return apperror.New(http.StatusServiceUnavailable, response.CodeInternalError, "账号服务需要配置 MySQL 后才能使用")
}

// logSecurity 写入账号安全日志；日志只保存脱敏后的详情和哈希后的客户端信息。
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

// validateNickname 校验昵称长度，昵称允许重复，系统内部通过 user_id 区分用户。
func validateNickname(nickname string) (string, error) {
	value := strings.TrimSpace(nickname)
	count := utf8.RuneCountInString(value)
	if count < 2 || count > 20 {
		return "", apperror.BadRequest(response.CodeBadRequest, "昵称需为 2-20 个字符")
	}
	return value, nil
}

// validateProfileFields 校验后续匹配依赖的必填画像字段。
func validateProfileFields(petStage string, interestedPetTypes []string) error {
	if strings.TrimSpace(petStage) == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请选择当前养宠阶段")
	}
	if len(interestedPetTypes) == 0 {
		return apperror.BadRequest(response.CodeBadRequest, "请至少选择一种感兴趣的宠物")
	}
	return nil
}

// buildProfile 将前端提交的画像字段转换为 user_profiles 表结构。
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

// toUserInfo 将用户主表模型转换为接口响应结构。
func toUserInfo(user *dao.User) *proto.UserInfo {
	if user == nil {
		return nil
	}
	return &proto.UserInfo{Id: int64(user.ID), Nickname: user.Nickname, Status: user.Status}
}

// toProfile 将画像模型转换为接口响应结构，兼容用户尚未填写画像的情况。
func toProfile(profile *dao.UserProfile) *proto.UserProfile {
	if profile == nil {
		return &proto.UserProfile{}
	}
	return &proto.UserProfile{
		PetStage:           profile.PetStage,
		InterestedPetTypes: data.ParseJSONStringArray(profile.InterestedPetTypes),
		PetExperience:      profile.PetExperience,
		DailyCompanyTime:   profile.DailyCompanyTime,
		LivingSituation:    profile.LivingSituation,
		PetConstraints:     data.ParseJSONStringArray(profile.PetConstraints),
	}
}

// toIdentities 将绑定身份转换为响应结构，同时补充脱敏展示值。
func toIdentities(items []dao.UserAuthIdentity) []*proto.UserIdentity {
	result := make([]*proto.UserIdentity, 0, len(items))
	for _, item := range items {
		value := ""
		if item.IdentityValue != nil {
			value = *item.IdentityValue
		}
		identityType := "phone"
		if item.IdentityType == dao.IdentityTypeEmail {
			identityType = "email"
		}
		result = append(result, &proto.UserIdentity{
			Id:            int64(item.ID),
			IdentityType:  identityType,
			IdentityValue: value,
			MaskedValue:   security.MaskIdentity(item.IdentityType, value),
			VerifyTime:    item.VerifyTime,
		})
	}
	return result
}

// internalError 将底层错误包装为统一的接口错误结构。
func internalError(err error) *apperror.Error {
	return apperror.New(http.StatusInternalServerError, response.CodeInternalError, err.Error())
}
