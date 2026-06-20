package data

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FindProfile 查询用户画像；画像缺失不视为系统错误，返回 nil 由 page 层兜底为空画像。
func (s *Store) FindProfile(ctx context.Context, userID uint64) (*dao.UserProfile, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var profile dao.UserProfile
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &profile, err
}

// UpsertProfile 新增或覆盖用户画像，适用于资料完善后再次编辑基础资料。
func (s *Store) UpsertProfile(ctx context.Context, profile *dao.UserProfile) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"pet_stage", "interested_pet_types", "pet_experience", "daily_company_time", "living_situation", "pet_constraints", "status", "update_time"}),
	}).Create(profile).Error
}

// UpdateUserProfileFields 同步更新 users.nickname 和 user_profiles，调用方通常已在外层事务中执行。
func (s *Store) UpdateUserProfileFields(ctx context.Context, user *dao.User, profile *dao.UserProfile) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Model(&dao.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"nickname":    user.Nickname,
			"update_time": user.UpdateTime,
		}).Error; err != nil {
		return err
	}
	return s.UpsertProfile(ctx, profile)
}
