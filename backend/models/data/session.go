package data

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm"
)

// CreateSession 创建登录设备会话，refresh token 只保存摘要。
func (s *Store) CreateSession(ctx context.Context, session *dao.UserSession) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(session).Error
}

// FindSessionByID 按 session ID 查询设备会话，鉴权 cache 未命中时会回源使用。
func (s *Store) FindSessionByID(ctx context.Context, sessionID uint64) (*dao.UserSession, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var session dao.UserSession
	err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, err
}

// FindSessionByRefreshHash 查询仍然有效的 refresh token 会话，用于刷新 token 和退出登录。
func (s *Store) FindSessionByRefreshHash(ctx context.Context, refreshHash string) (*dao.UserSession, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var session dao.UserSession
	err := s.db.WithContext(ctx).
		Where("refresh_token_hash = ? AND status = ?", refreshHash, dao.SessionStatusNormal).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, err
}

// UpdateSessionRefreshToken 轮换 refresh token 摘要，并刷新会话活跃时间和过期时间。
func (s *Store) UpdateSessionRefreshToken(ctx context.Context, sessionID uint64, refreshHash string, lastActiveTime int64, expireTime int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&dao.UserSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"refresh_token_hash": refreshHash,
			"last_active_time":   lastActiveTime,
			"expire_time":        expireTime,
			"update_time":        lastActiveTime,
		}).Error
}

// RevokeSession 撤销单个设备会话。userID 大于 0 时会限制只能操作当前用户自己的设备。
func (s *Store) RevokeSession(ctx context.Context, sessionID uint64, userID uint64, status int32, reason string, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	query := s.db.WithContext(ctx).Model(&dao.UserSession{}).Where("id = ? AND status = ?", sessionID, dao.SessionStatusNormal)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	return query.Updates(map[string]interface{}{
		"status":        status,
		"revoke_time":   now,
		"revoke_reason": reason,
		"update_time":   now,
	}).Error
}

// RevokeUserSessions 批量撤销用户会话，可通过 exceptSessionID 保留当前设备。
func (s *Store) RevokeUserSessions(ctx context.Context, userID uint64, exceptSessionID uint64, status int32, reason string, now int64) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	query := s.db.WithContext(ctx).Model(&dao.UserSession{}).
		Where("user_id = ? AND status = ?", userID, dao.SessionStatusNormal)
	if exceptSessionID > 0 {
		query = query.Where("id <> ?", exceptSessionID)
	}
	return query.Updates(map[string]interface{}{
		"status":        status,
		"revoke_time":   now,
		"revoke_reason": reason,
		"update_time":   now,
	}).Error
}

// ListSessions 查询账号设备列表，供账号安全页展示当前设备和历史设备。
func (s *Store) ListSessions(ctx context.Context, userID uint64) ([]dao.UserSession, error) {
	if err := s.ensureDB(); err != nil {
		return nil, err
	}
	var sessions []dao.UserSession
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_active_time DESC, id DESC").
		Find(&sessions).Error
	return sessions, err
}
