package data

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm"
)

// FindIdentity 按身份类型和值查询正常绑定记录；不存在时返回 nil，便于 page 层区分未注册和数据库错误。
func (s *Store) FindIdentity(ctx context.Context, identityType int32, identityValue string) (*dao.UserAuthIdentity, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var item dao.UserAuthIdentity
	err := s.db.WithContext(ctx).
		Where("identity_type = ? AND identity_value = ? AND status = ?", identityType, identityValue, dao.IdentityStatusNormal).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

// ListIdentities 查询用户当前可用的邮箱/手机号绑定，已解绑记录不会返回给前端。
func (s *Store) ListIdentities(ctx context.Context, userID uint64) ([]dao.UserAuthIdentity, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var items []dao.UserAuthIdentity
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, dao.IdentityStatusNormal).
		Order("identity_type ASC").
		Find(&items).Error
	return items, err
}

// CountActiveIdentities 统计账号可用联系方式数量，用于解绑前保证至少保留一种联系方式。
func (s *Store) CountActiveIdentities(ctx context.Context, userID uint64) (int64, error) {
	if err := s.ensureDB(); err != nil {
		return 0, err
	}
	var count int64
	err := s.db.WithContext(ctx).Model(&dao.UserAuthIdentity{}).
		Where("user_id = ? AND status = ?", userID, dao.IdentityStatusNormal).
		Count(&count).Error
	return count, err
}

// UpsertIdentity 绑定或重新绑定邮箱/手机号；如果目标联系方式属于其他用户，返回重复键错误。
func (s *Store) UpsertIdentity(ctx context.Context, identity *dao.UserAuthIdentity) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	existing, err := s.FindIdentity(ctx, identity.IdentityType, deref(identity.IdentityValue))
	if err != nil {
		return err
	}
	if existing != nil && existing.UserID != identity.UserID {
		return gorm.ErrDuplicatedKey
	}
	if existing != nil {
		return s.db.WithContext(ctx).Model(&dao.UserAuthIdentity{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"user_id":     identity.UserID,
				"verify_time": identity.VerifyTime,
				"status":      dao.IdentityStatusNormal,
				"update_time": identity.UpdateTime,
			}).Error
	}
	return s.db.WithContext(ctx).Create(identity).Error
}

// UnbindIdentity 逻辑解绑邮箱或手机号：清空 identity_value 并把状态置为已解绑，不物理删除记录。
func (s *Store) UnbindIdentity(ctx context.Context, userID uint64, identityType int32, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&dao.UserAuthIdentity{}).
		Where("user_id = ? AND identity_type = ? AND status = ?", userID, identityType, dao.IdentityStatusNormal).
		Updates(map[string]interface{}{
			"identity_value": nil,
			"status":         dao.IdentityStatusUnbound,
			"update_time":    now,
		}).Error
}

// deref 安全读取字符串指针，nil 时返回空字符串。
func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
