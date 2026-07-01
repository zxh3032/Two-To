package auth

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// Refresh 使用 refresh token 查找会话，轮换 refresh token 并签发新的 access token。
func (s *Service) Refresh(ctx *servlet.Context, req *proto.RefreshRequest, resp *proto.RefreshResponse) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	// refresh token 明文只在客户端保存，服务端用摘要查找会话。
	rawRefreshToken := strings.TrimSpace(req.GetRefreshToken())
	if rawRefreshToken == "" {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效，请重新登录")
	}
	refreshHash := s.tokenManager.HashRefreshToken(rawRefreshToken)
	if err := s.rejectReusedRefreshToken(ctx, refreshHash); err != nil {
		return err
	}
	session, err := s.store.FindSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return internalError(err)
	}
	now := data.Now()
	if session == nil || session.ExpireTime <= now || session.Status != dao.SessionStatusNormal {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效，请重新登录")
	}
	lockKey := s.cacheStore.Key("refresh-lock", strconv.FormatUint(session.ID, 10))
	locked, err := s.cacheStore.SetNX(ctx, lockKey, "1", 15*time.Second)
	if err != nil {
		return internalError(err)
	}
	if !locked {
		return apperror.RateLimited(response.CodeRateLimited, "登录态正在刷新，请稍后重试")
	}
	defer func() {
		_ = s.cacheStore.Del(ctx, lockKey)
	}()
	// 每次刷新都轮换 refresh token，降低旧 token 泄露后的可复用窗口。
	refreshToken, newRefreshHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return internalError(err)
	}
	expireTime := now + int64(s.tokenManager.RefreshTTL().Seconds())
	updated, err := s.store.UpdateSessionRefreshTokenIfMatch(ctx, session.ID, refreshHash, newRefreshHash, now, expireTime)
	if err != nil {
		return internalError(err)
	}
	if !updated {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已刷新，请使用新的登录态")
	}
	if err := s.recordRotatedRefreshToken(ctx, refreshHash, session.ID, session.ExpireTime); err != nil {
		return err
	}
	if err := authlib.SetCachedSession(ctx, s.cacheStore, session.ID, authlib.CachedSession{UserID: session.UserID, RefreshTokenHash: newRefreshHash, ExpireTime: expireTime}); err != nil {
		return internalError(err)
	}
	accessToken, accessTTL, err := s.tokenManager.GenerateAccessToken(session.UserID, session.ID)
	if err != nil {
		return internalError(err)
	}
	s.logSecurity(ctx, session.UserID, "refresh_token", map[string]string{"sessionId": strconv.FormatUint(session.ID, 10)})
	resp.AccessToken = accessToken
	resp.AccessTokenExpiresIn = accessTTL
	resp.RefreshToken = refreshToken
	resp.RefreshTokenExpiresIn = int64(s.tokenManager.RefreshTTL().Seconds())
	return nil
}

func (s *Service) refreshReplayKey(refreshHash string) string {
	return s.cacheStore.Key("refresh-used", refreshHash)
}

func (s *Service) recordRotatedRefreshToken(ctx *servlet.Context, oldRefreshHash string, sessionID uint64, oldExpireTime int64) error {
	ttl := time.Until(time.Unix(oldExpireTime, 0))
	if ttl <= 0 {
		return nil
	}
	key := s.refreshReplayKey(oldRefreshHash)
	if err := s.cacheStore.Set(ctx, key, strconv.FormatUint(sessionID, 10), ttl); err != nil {
		return internalError(err)
	}
	return nil
}

func (s *Service) rejectReusedRefreshToken(ctx *servlet.Context, refreshHash string) error {
	rawSessionID, err := s.cacheStore.Get(ctx, s.refreshReplayKey(refreshHash))
	if errors.Is(err, cache.ErrMiss) {
		return nil
	}
	if err != nil {
		return internalError(err)
	}
	sessionID, parseErr := strconv.ParseUint(rawSessionID, 10, 64)
	if parseErr == nil && sessionID > 0 {
		now := data.Now()
		session, findErr := s.store.FindSessionByID(ctx, sessionID)
		if findErr != nil {
			return internalError(findErr)
		}
		if session != nil {
			_ = s.store.RevokeSession(ctx, session.ID, 0, dao.SessionStatusRevoked, "refresh_token_reused", now)
			_ = authlib.DeleteCachedSession(ctx, s.cacheStore, session.ID)
			s.logSecurity(ctx, session.UserID, "refresh_token_reused", map[string]string{"sessionId": strconv.FormatUint(session.ID, 10)})
		}
	}
	return apperror.Unauthorized(response.CodeUnauthorized, "登录态存在异常，请重新登录")
}
