package data

import (
	"context"

	"github.com/zxh3032/two-to/backend/models/dao"
)

// LogSecurity 写入账号安全事件，detail 中只允许放脱敏后的摘要信息。
func (s *Store) LogSecurity(ctx context.Context, log *dao.SecurityLog) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(log).Error
}
