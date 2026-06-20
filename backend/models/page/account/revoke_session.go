package account

import (
	"strconv"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// RevokeSession 退出指定设备，当前设备需要走 logout 保持语义清晰。
func (s *Service) RevokeSession(ctx *servlet.Context, req *proto.RevokeSessionRequest, _ *proto.RevokeSessionResponse) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	if req.GetSessionId() == ctx.SessionID {
		return apperror.BadRequest(response.CodeBadRequest, "当前设备请使用退出登录")
	}
	now := data.Now()
	if err := s.store.RevokeSession(ctx, req.GetSessionId(), ctx.UserID, dao.SessionStatusRevoked, "user_revoke", now); err != nil {
		return internalError(err)
	}
	_ = authlib.DeleteCachedSession(ctx, s.cacheStore, req.GetSessionId())
	s.logSecurity(ctx, ctx.UserID, "revoke_session", map[string]string{"sessionId": strconv.FormatUint(req.GetSessionId(), 10)})
	return nil
}
