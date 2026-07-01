package account

import (
	"context"
	"strings"
	"time"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/models/data"
)

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
		attempts, err := s.recordVerificationFailure(ctx, codeType, scene, target)
		if err != nil {
			return err
		}
		if attempts >= s.maxVerificationAttempts() {
			return apperror.BadRequest(response.CodeBadRequest, "验证码错误次数过多，请重新获取")
		}
		return apperror.BadRequest(response.CodeBadRequest, "验证码不正确")
	}
	// 验证码成功使用后立即失效，防止同一验证码被重复绑定。
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
