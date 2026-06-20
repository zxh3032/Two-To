package ping

import (
	"errors"
	"time"

	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

// Page 承载 ping 接口的一次业务编排，后续复杂接口也按 page 层承接业务流程。
type Page struct {
	cfg config.Config
	log *zap.Logger
}

// New 创建 ping page，注入配置和日志以便业务层记录关键上下文。
func New(cfg config.Config, log *zap.Logger) *Page {
	return &Page{
		cfg: cfg,
		log: log,
	}
}

// Handle 处理 ping 请求，返回服务基础信息和本次请求的链路追踪 ID。
func (p *Page) Handle(ctx *servlet.Context, req *proto.PingRequest, resp *proto.PingResponse) error {
	if req == nil {
		return errors.New("ping request is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	p.log.Info("处理 ping 请求", zap.String("requestId", ctx.RequestID))
	resp.Message = "pong"
	resp.Service = p.cfg.AppName
	resp.Environment = p.cfg.Env
	resp.RequestId = ctx.RequestID
	resp.Timestamp = time.Now().Unix()
	return nil
}
