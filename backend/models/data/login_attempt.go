package data

import (
	"context"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm/clause"
)

// UpsertLoginAttempt 记录登录失败次数和锁定信息，Redis 负责实时风控，MySQL 负责审计留痕。
func (s *Store) UpsertLoginAttempt(ctx context.Context, item *dao.LoginAttempt) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "identity_type"}, {Name: "identity_value"}, {Name: "ip_hash"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"fail_count":      item.FailCount,
			"lock_until_time": item.LockUntilTime,
			"last_fail_time":  item.LastFailTime,
			"status":          item.Status,
			"update_time":     item.UpdateTime,
		}),
	}).Create(item).Error
}
