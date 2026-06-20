package account

import (
	"context"
	"strconv"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
)

// RevokeSession 退出指定设备，当前设备需要走 logout 保持语义清晰。
func (s *Service) RevokeSession(ctx context.Context, userID uint64, currentSessionID uint64, sessionID uint64, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	if sessionID == currentSessionID {
		return apperror.BadRequest(response.CodeBadRequest, "当前设备请使用退出登录")
	}
	now := data.Now()
	if err := s.store.RevokeSession(ctx, sessionID, userID, dao.SessionStatusRevoked, "user_revoke", now); err != nil {
		return internalError(err)
	}
	_ = authlib.DeleteCachedSession(ctx, s.cacheStore, sessionID)
	s.logSecurity(ctx, userID, "revoke_session", map[string]string{"sessionId": strconv.FormatUint(sessionID, 10)}, meta)
	return nil
}
