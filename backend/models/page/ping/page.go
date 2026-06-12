package ping

import (
	"context"
	"errors"
	"time"

	"github.com/zxh3032/two-to/backend/library/config"
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
func (p *Page) Handle(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	if req == nil {
		return nil, errors.New("ping request is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	p.log.Info("处理 ping 请求", zap.String("requestId", req.GetRequestId()))
	return &proto.PingResponse{
		Message:     "pong",
		Service:     p.cfg.AppName,
		Environment: p.cfg.Env,
		RequestId:   req.GetRequestId(),
		Timestamp:   time.Now().Unix(),
	}, nil
}
