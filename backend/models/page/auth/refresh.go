package auth

import (
	"strconv"
	"strings"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
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
	refreshHash := s.tokenManager.HashRefreshToken(strings.TrimSpace(req.GetRefreshToken()))
	session, err := s.store.FindSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return internalError(err)
	}
	now := data.Now()
	if session == nil || session.ExpireTime <= now || session.Status != dao.SessionStatusNormal {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效，请重新登录")
	}
	// 每次刷新都轮换 refresh token，降低旧 token 泄露后的可复用窗口。
	refreshToken, newRefreshHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return internalError(err)
	}
	expireTime := now + int64(s.tokenManager.RefreshTTL().Seconds())
	if err := s.store.UpdateSessionRefreshToken(ctx, session.ID, newRefreshHash, now, expireTime); err != nil {
		return internalError(err)
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
