package account

import (
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/models/data"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 编排账号资料、安全设置和设备管理。
type Service struct {
	cfg          config.Config
	log          *zap.Logger
	store        *data.Store
	cacheStore   cache.Store
	tokenManager *authlib.TokenManager
}

// NewService 创建账号中心 page service，集中注入 DB、缓存和 token 管理器。
func NewService(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Service {
	return &Service{
		cfg:          cfg,
		log:          log,
		store:        data.NewStore(db),
		cacheStore:   cacheStore,
		tokenManager: tokenManager,
	}
}
