package servlet

import (
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Runtime 聚合 controller 创建 page service 时需要的基础依赖。
type Runtime struct {
	Config       config.Config
	Log          *zap.Logger
	DB           *gorm.DB
	Cache        cache.Store
	TokenManager *authlib.TokenManager
}

// Action 返回一次 HTTP 接口执行所需的模块和动作标识，用于响应日志。
func (r Runtime) Action(module string, name string) Action {
	return Action{Module: module, Name: name, Log: r.Log}
}

// Action 描述一个 HTTP 接口动作。
type Action struct {
	Module string
	Name   string
	Log    *zap.Logger
}
