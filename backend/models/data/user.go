package data

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm"
)

// FindUser 按用户 ID 查询 users 主表；不存在时返回 nil。
func (s *Store) FindUser(ctx context.Context, userID uint64) (*dao.User, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var user dao.User
	err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// CreateUserWithIdentityAndProfile 在一个事务内创建用户主表、登录身份和画像，保证注册流程不会产生半成品账号。
func (s *Store) CreateUserWithIdentityAndProfile(ctx context.Context, user *dao.User, identity *dao.UserAuthIdentity, profile *dao.UserProfile) error {
	return s.WithTx(ctx, func(tx *Store) error {
		if err := tx.db.Create(user).Error; err != nil {
			return err
		}
		identity.UserID = user.ID
		profile.UserID = user.ID
		if err := tx.db.Create(identity).Error; err != nil {
			return err
		}
		return tx.db.Create(profile).Error
	})
}

// UpdatePassword 更新用户密码摘要和密码更新时间。调用方负责先校验当前密码或重置 token。
func (s *Store) UpdatePassword(ctx context.Context, userID uint64, passwordHash string, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&dao.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password_hash":        passwordHash,
			"password_update_time": now,
			"update_time":          now,
		}).Error
}
