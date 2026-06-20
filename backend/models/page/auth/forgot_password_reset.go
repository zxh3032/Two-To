package auth

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
)

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
