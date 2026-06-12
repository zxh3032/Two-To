package database

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// OpenMySQL 使用 GORM 初始化 MySQL 连接。DSN 为空时跳过连接，便于本地先跑通 API 骨架。
func OpenMySQL(ctx context.Context, dsn string, log *zap.Logger) (*gorm.DB, error) {
	if dsn == "" {
		log.Warn("未配置 MySQL DSN，跳过数据库初始化")
		return nil, nil
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Info("MySQL 数据库初始化完成")
	return db, nil
}
