package data

import (
	"context"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm"
)

// CreateVerificationCode 写入验证码审计记录。验证码明文只存在 cache，MySQL 只保存摘要。
func (s *Store) CreateVerificationCode(ctx context.Context, code *dao.VerificationCode) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(code).Error
}

// MarkVerificationCodeUsed 将验证码审计记录标记为已使用，和 cache 删除一起表达“成功使用后立即失效”。
func (s *Store) MarkVerificationCodeUsed(ctx context.Context, codeType int32, scene string, target string, codeHash string, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&dao.VerificationCode{}).
		Where("code_type = ? AND scene = ? AND target = ? AND code_hash = ? AND status = ?", codeType, scene, target, codeHash, dao.VerificationStatusUnused).
		Updates(map[string]interface{}{
			"status":      dao.VerificationStatusUsed,
			"use_time":    now,
			"update_time": now,
		}).Error
}

// IncrementVerificationCodeAttempt 记录验证码错误次数，MySQL 仅作为审计留痕。
func (s *Store) IncrementVerificationCodeAttempt(ctx context.Context, codeType int32, scene string, target string, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&dao.VerificationCode{}).
		Where("code_type = ? AND scene = ? AND target = ? AND status = ? AND expire_time > ?", codeType, scene, target, dao.VerificationStatusUnused, now).
		Updates(map[string]interface{}{
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"update_time":   now,
		}).Error
}

// RevokeVerificationCode 将当前仍可用的验证码作废，适用于错误次数过多等风控场景。
func (s *Store) RevokeVerificationCode(ctx context.Context, codeType int32, scene string, target string, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&dao.VerificationCode{}).
		Where("code_type = ? AND scene = ? AND target = ? AND status = ? AND expire_time > ?", codeType, scene, target, dao.VerificationStatusUnused, now).
		Updates(map[string]interface{}{
			"status":      dao.VerificationStatusRevoked,
			"update_time": now,
		}).Error
}
