package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
)

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
