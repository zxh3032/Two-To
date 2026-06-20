package account

import (
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/proto"
)

// Sessions 返回当前账号的登录设备列表，并标记当前 access token 对应的会话。
func (s *Service) Sessions(ctx *servlet.Context, _ *proto.SessionsRequest, resp *proto.SessionsResponse) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	sessions, err := s.store.ListSessions(ctx, ctx.UserID)
	if err != nil {
		return internalError(err)
	}
	resp.CurrentSessionId = int64(ctx.SessionID)
	for _, session := range sessions {
		resp.Sessions = append(resp.Sessions, &proto.UserSession{
			Id:             int64(session.ID),
			DeviceName:     session.DeviceName,
			Status:         session.Status,
			LastActiveTime: session.LastActiveTime,
			ExpireTime:     session.ExpireTime,
			Current:        session.ID == ctx.SessionID,
		})
	}
	return nil
}
