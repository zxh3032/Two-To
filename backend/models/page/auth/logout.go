package auth

import (
	"strings"

	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// Logout 撤销 refresh token 对应的当前设备会话，并删除 session cache。
func (s *Service) Logout(ctx *servlet.Context, req *proto.LogoutRequest, _ *proto.LogoutResponse) error {
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
