package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"go.uber.org/zap"
)

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
		attempts, err := s.recordVerificationFailure(ctx, codeType, scene, target)
		if err != nil {
			return err
		}
		if attempts >= s.maxVerificationAttempts() {
			return apperror.BadRequest(response.CodeBadRequest, "验证码错误次数过多，请重新获取")
		}
		return apperror.BadRequest(response.CodeBadRequest, "验证码不正确")
	}
	// 验证码成功使用后立即失效，符合“同一验证码只能成功使用一次”的规则。
	_ = s.cacheStore.Del(ctx, key, s.verificationAttemptKey(scene, target))
	if s.store.DBReady() {
		_ = s.store.MarkVerificationCodeUsed(ctx, codeType, scene, target, s.tokenManager.HashCode(code), data.Now())
	}
	return nil
}

func (s *Service) verificationAttemptKey(scene string, target string) string {
	return s.cacheStore.Key("verify-code-attempts", scene, target)
}

func (s *Service) maxVerificationAttempts() int64 {
	if s.cfg.Verification.CodeMaxAttempts <= 0 {
		return 5
	}
	return s.cfg.Verification.CodeMaxAttempts
}

func (s *Service) recordVerificationFailure(ctx context.Context, codeType int32, scene string, target string) (int64, error) {
	attemptKey := s.verificationAttemptKey(scene, target)
	attempts, err := s.cacheStore.IncrWithTTL(ctx, attemptKey, time.Duration(s.cfg.Verification.CodeTTLSeconds)*time.Second)
	if err != nil {
		return 0, internalError(err)
	}
	now := data.Now()
	if s.store.DBReady() {
		_ = s.store.IncrementVerificationCodeAttempt(ctx, codeType, scene, target, now)
	}
	if attempts >= s.maxVerificationAttempts() {
		_ = s.cacheStore.Del(ctx, s.cacheStore.Key("verify-code", scene, target), attemptKey)
		if s.store.DBReady() {
			_ = s.store.RevokeVerificationCode(ctx, codeType, scene, target, now)
		}
	}
	return attempts, nil
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
func (s *Service) auditVerificationCode(ctx *servlet.Context, codeType int32, scene string, target string, code string, provider string) {
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
		IPHash:        security.ClientIPHash(ctx.IP),
		UserAgentHash: security.HashPlain(ctx.UserAgent),
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

func (s *Service) loginLockKey(identity string, ip string) string {
	return s.cacheStore.Key("login-lock", identityTypeEmailText, identity, security.ClientIPHash(ip))
}

func (s *Service) maxLoginFailuresBeforeLock() int64 {
	if s.cfg.Verification.LoginLockFailThreshold <= 0 {
		return 5
	}
	return s.cfg.Verification.LoginLockFailThreshold
}

func (s *Service) loginLockTTL() time.Duration {
	if s.cfg.Verification.LoginLockTTLSeconds <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(s.cfg.Verification.LoginLockTTLSeconds) * time.Second
}

func (s *Service) checkLoginLock(ctx context.Context, key string) error {
	_, err := s.cacheStore.Get(ctx, key)
	if errors.Is(err, cache.ErrMiss) {
		return nil
	}
	if err != nil {
		return internalError(err)
	}
	ttl, err := s.cacheStore.TTL(ctx, key)
	if err != nil {
		return apperror.RateLimited(response.CodeRateLimited, "登录失败次数过多，请稍后再试")
	}
	minutes := int64(ttl.Minutes())
	if minutes <= 0 {
		minutes = 1
	}
	return apperror.RateLimited(response.CodeRateLimited, "登录失败次数过多，请 "+strconv.FormatInt(minutes, 10)+" 分钟后再试")
}

// currentFailCount 读取登录失败次数。cache 不存在时表示还未触发风控。
func (s *Service) currentFailCount(ctx context.Context, key string) (int64, error) {
	raw, err := s.cacheStore.Get(ctx, key)
	if errors.Is(err, cache.ErrMiss) {
		return 0, nil
	}
	if err != nil {
		return 0, internalError(err)
	}
	count, _ := strconv.ParseInt(raw, 10, 64)
	return count, nil
}

// recordLoginFailure 同步记录 Redis 失败计数和 MySQL 审计记录。
func (s *Service) recordLoginFailure(ctx *servlet.Context, failKey string, lockKey string, identityType int32, identityValue string) (int64, error) {
	count, err := s.cacheStore.IncrWithTTL(ctx, failKey, 15*time.Minute)
	if err != nil {
		s.log.Error("登录失败计数写入 cache 失败", zap.Error(err))
		return 0, internalError(err)
	}
	lockUntilTime := int64(0)
	if count >= s.maxLoginFailuresBeforeLock() {
		lockUntilTime = data.Now() + int64(s.loginLockTTL().Seconds())
		if err := s.cacheStore.Set(ctx, lockKey, strconv.FormatInt(lockUntilTime, 10), s.loginLockTTL()); err != nil {
			return 0, internalError(err)
		}
	}
	if s.store.DBReady() {
		now := data.Now()
		_ = s.store.UpsertLoginAttempt(ctx, &dao.LoginAttempt{
			IdentityType:  identityType,
			IdentityValue: identityValue,
			IPHash:        security.ClientIPHash(ctx.IP),
			FailCount:     int32(count),
			LockUntilTime: lockUntilTime,
			LastFailTime:  now,
			BaseModel:     dao.NewBaseModel(dao.LoginAttemptStatusNormal, now),
		})
	}
	return count, nil
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
