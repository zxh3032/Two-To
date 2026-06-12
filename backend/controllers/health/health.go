package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Data 描述服务健康检查结果，database 字段用于判断数据库是否已配置并可访问。
type Data struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Database    string `json:"database"`
}

// Check 返回服务健康状态，数据库未配置时不阻断本地开发启动。
func Check(cfg config.Config, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result := Data{
			Status:      "ok",
			Service:     cfg.AppName,
			Environment: cfg.Env,
			Database:    "not_configured",
		}

		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				log.Error("健康检查读取数据库连接失败", zap.Error(err))
				result.Status = "degraded"
				result.Database = "error"
				ctx.JSON(http.StatusServiceUnavailable, response.Body{
					Code:    response.CodeInternalError,
					Message: "database unavailable",
					Data:    result,
				})
				return
			}
			if err := sqlDB.PingContext(ctx.Request.Context()); err != nil {
				log.Error("健康检查数据库 ping 失败", zap.Error(err))
				result.Status = "degraded"
				result.Database = "error"
				ctx.JSON(http.StatusServiceUnavailable, response.Body{
					Code:    response.CodeInternalError,
					Message: "database unavailable",
					Data:    result,
				})
				return
			}
			result.Database = "ok"
		}

		response.Success(ctx, result)
	}
}
