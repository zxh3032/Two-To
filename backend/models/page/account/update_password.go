package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
)

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
