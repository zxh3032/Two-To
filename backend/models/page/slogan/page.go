package slogan

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/proto"
	"go.uber.org/zap"
)

const projectSlogan = "两两相逢，奔赴朝夕。"

// Page 承载项目 slogan 接口的一次业务编排，后续可在这里扩展多语言或运营配置。
type Page struct {
	log *zap.Logger
}

// New 创建 slogan page，注入日志以便记录前后端联调链路。
func New(log *zap.Logger) *Page {
	return &Page{log: log}
}

// Handle 返回 Two-To 当前项目口号，并透传 request id 方便排查链路。
func (p *Page) Handle(ctx context.Context, req *proto.SloganRequest) (*proto.SloganResponse, error) {
	if req == nil {
		return nil, errors.New("slogan request is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	p.log.Info("处理项目 slogan 请求", zap.String("requestId", req.GetRequestId()))
	return &proto.SloganResponse{
		Slogan:    projectSlogan,
		RequestId: req.GetRequestId(),
	}, nil
}
