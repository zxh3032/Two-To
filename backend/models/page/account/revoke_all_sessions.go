package account

import (
	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// RevokeAllSessions 校验当前密码后退出全部设备。
func (s *Service) RevokeAllSessions(ctx *servlet.Context, req *proto.RevokeAllSessionsRequest, _ *proto.RevokeAllSessionsResponse) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	user, err := s.store.FindUser(ctx, ctx.UserID)
	if err != nil {
		return internalError(err)
	}
	if user == nil {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if user.PasswordHash == "" || !authlib.CheckPassword(user.PasswordHash, req.GetCurrentPassword()) {
		return apperror.BadRequest(response.CodeBadRequest, "当前密码不正确")
	}
	sessions, _ := s.store.ListSessions(ctx, ctx.UserID)
	now := data.Now()
	if err := s.store.RevokeUserSessions(ctx, ctx.UserID, 0, dao.SessionStatusRevoked, "revoke_all", now); err != nil {
		return internalError(err)
	}
	for _, session := range sessions {
		_ = authlib.DeleteCachedSession(ctx, s.cacheStore, session.ID)
	}
	s.logSecurity(ctx, ctx.UserID, "revoke_all_sessions", map[string]string{})
	return nil
}
