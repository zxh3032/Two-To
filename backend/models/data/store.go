package data

import (
	"context"
	"errors"

	"gorm.io/gorm"
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
