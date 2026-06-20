package health

import (
	"net/http"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Page 承载健康检查接口的一次执行。
type Page struct {
	cfg config.Config
	db  *gorm.DB
	log *zap.Logger
}

// New 创建健康检查 page。
func New(cfg config.Config, db *gorm.DB, log *zap.Logger) *Page {
	return &Page{cfg: cfg, db: db, log: log}
}

// Handle 返回服务健康状态，数据库未配置时不阻断本地开发启动。
func (p *Page) Handle(ctx *servlet.Context, _ *proto.HealthRequest, resp *proto.HealthResponse) error {
	resp.Status = "ok"
	resp.Service = p.cfg.AppName
	resp.Environment = p.cfg.Env
	resp.Database = "not_configured"

	if p.db == nil {
		return nil
	}

	sqlDB, err := p.db.DB()
	if err != nil {
		p.log.Error("健康检查读取数据库连接失败", zap.Error(err))
		return p.databaseUnavailable(resp, err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		p.log.Error("健康检查数据库 ping 失败", zap.Error(err))
		return p.databaseUnavailable(resp, err)
	}
	resp.Database = "ok"
	return nil
}

func (p *Page) databaseUnavailable(resp *proto.HealthResponse, cause error) error {
	resp.Status = "degraded"
	resp.Database = "error"
	return &apperror.Error{
		HTTPStatus: http.StatusServiceUnavailable,
		Code:       response.CodeInternalError,
		Message:    "database unavailable",
		Data:       resp,
		Cause:      cause,
	}
}
