package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/database"
	"github.com/zxh3032/two-to/backend/library/logger"
	"github.com/zxh3032/two-to/backend/routers"
	"go.uber.org/zap"
)

// main 负责装配配置、日志、数据库、路由和 HTTP 服务生命周期。
func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.LogLevel, cfg.Env)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()
	if err := cfg.ValidateStartup(); err != nil {
		log.Fatal("启动配置校验失败", zap.Error(err))
	}

	initCtx, cancelInit := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelInit()

	db, err := database.OpenMySQL(initCtx, cfg.MySQLDSN, log)
	if err != nil {
		log.Fatal("初始化 MySQL 数据库失败", zap.Error(err))
	}

	cacheStore, err := cache.Open(initCtx, cfg.Cache, log)
	if err != nil {
		log.Fatal("初始化 cache 失败", zap.Error(err))
	}
	defer func() {
		if err := cacheStore.Close(); err != nil {
			log.Error("关闭 cache 失败", zap.Error(err))
		}
	}()

	router := routers.NewRouter(cfg, log, db, cacheStore)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("Two-To API 服务启动",
			zap.String("addr", cfg.HTTPAddr),
			zap.String("env", cfg.Env),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP 服务启动失败", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("收到停止信号，开始优雅关闭 HTTP 服务")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP 服务关闭失败", zap.Error(err))
	}

	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			log.Error("读取数据库底层连接失败", zap.Error(err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Error("关闭数据库连接失败", zap.Error(err))
		}
	}
}
