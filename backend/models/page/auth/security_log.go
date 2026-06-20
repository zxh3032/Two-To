package auth

import (
	"context"
	"encoding/json"

	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"go.uber.org/zap"
)

// logSecurity 记录安全事件，detail 必须使用脱敏后的摘要，不能写入密码、验证码或 token。
func (s *Service) logSecurity(ctx context.Context, userID uint64, eventType string, detail map[string]string, meta pagectx.RequestMeta) {
	if !s.store.DBReady() {
		return
	}
	raw, _ := json.Marshal(detail)
	now := data.Now()
	if err := s.store.LogSecurity(ctx, &dao.SecurityLog{
		UserID:        userID,
		EventType:     eventType,
		Detail:        string(raw),
		IPHash:        security.ClientIPHash(meta.IP),
		UserAgentHash: security.HashPlain(meta.UserAgent),
		BaseModel:     dao.NewBaseModel(dao.SecurityLogStatusNormal, now),
	}); err != nil {
		s.log.Error("安全日志写入失败", zap.Error(err))
	}
}
