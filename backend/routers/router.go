package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/controllers/health"
	"github.com/zxh3032/two-to/backend/controllers/ping"
	"github.com/zxh3032/two-to/backend/controllers/slogan"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/middlewares"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewRouter 注册后端所有 HTTP 路由和通用中间件。
func NewRouter(cfg config.Config, log *zap.Logger, db *gorm.DB) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middlewares.RequestID(),
		middlewares.Recovery(log),
		middlewares.AccessLog(log),
		middlewares.CORS(),
	)

	router.GET("/health", health.Check(cfg, db, log))

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/ping", ping.Ping(cfg, log))
	apiV1.GET("/slogan", slogan.Get(log))

	return router
}
