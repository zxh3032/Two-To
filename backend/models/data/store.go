package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/zxh3032/two-to/backend/models/dao"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrDBNotConfigured = errors.New("数据库未配置")

// Store 封装账号体系需要的数据库读写，page 层不直接触碰 GORM 细节。
type Store struct {
	db *gorm.DB
}

// NewStore 创建账号数据访问对象。db 允许为空，调用方可通过 DBReady 判断账号能力是否可用。
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// DBReady 判断 MySQL 是否已经初始化；本地只看页面时允许未配置数据库。
func (s *Store) DBReady() bool {
	return s != nil && s.db != nil
}

// ensureDB 是所有数据库写读方法的统一前置保护，避免 nil db 导致 panic。
func (s *Store) ensureDB() error {
	if !s.DBReady() {
		return ErrDBNotConfigured
	}
	return nil
}

// WithTx 在同一个 MySQL 事务中执行账号写操作，适合用户、身份、画像这类必须同时成功的流程。
func (s *Store) WithTx(ctx context.Context, fn func(tx *Store) error) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Store{db: tx})
	})
}

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

// LogSecurity 写入账号安全事件，detail 中只允许放脱敏后的摘要信息。
func (s *Store) LogSecurity(ctx context.Context, log *dao.SecurityLog) error {
	if err := s.ensureDB(); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(log).Error
}

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

// JSONStrings 把字符串数组转成 JSON 字符串，供 GORM 写入 MySQL JSON 字段。
func JSONStrings(values []string) string {
	raw, _ := json.Marshal(values)
	return string(raw)
}

// ParseJSONStringArray 把 MySQL JSON 字段解析为字符串数组，解析失败时按空值处理。
func ParseJSONStringArray(raw string) []string {
	if raw == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return values
}

// Now 返回 Unix 秒级时间戳，统一匹配表字段的 create_time/update_time 约定。
func Now() int64 {
	return time.Now().Unix()
}

// StringPtr 生成字符串指针，便于区分身份值空字符串和 NULL。
func StringPtr(value string) *string {
	return &value
}

// deref 安全读取字符串指针，nil 时返回空字符串。
func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
