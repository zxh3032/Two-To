package health

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/servlet"
	healthpage "github.com/zxh3032/two-to/backend/models/page/health"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Check 返回服务健康状态，数据库未配置时不阻断本地开发启动。
func Check(cfg config.Config, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	page := healthpage.New(cfg, db, log)
	return servlet.Empty[proto.HealthRequest, proto.HealthResponse](servlet.Action{Module: "health", Name: "check", Log: log}, page.Handle)
}
