package data

import (
	"context"

	"github.com/zxh3032/two-to/backend/models/dao"
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
