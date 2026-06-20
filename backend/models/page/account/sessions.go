package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/proto"
)

// Sessions 返回当前账号的登录设备列表，并标记当前 access token 对应的会话。
func (s *Service) Sessions(ctx context.Context, userID uint64, currentSessionID uint64) (*proto.SessionsResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	sessions, err := s.store.ListSessions(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	resp := &proto.SessionsResponse{CurrentSessionId: int64(currentSessionID)}
	for _, session := range sessions {
		resp.Sessions = append(resp.Sessions, &proto.UserSession{
			Id:             int64(session.ID),
			DeviceName:     session.DeviceName,
			Status:         session.Status,
			LastActiveTime: session.LastActiveTime,
			ExpireTime:     session.ExpireTime,
			Current:        session.ID == currentSessionID,
		})
	}
	return resp, nil
}
