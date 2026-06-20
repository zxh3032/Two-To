package auth

import (
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/email"
	"github.com/zxh3032/two-to/backend/library/sms"
	"github.com/zxh3032/two-to/backend/models/data"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	identityTypeEmailText = "email"
	identityTypePhoneText = "phone"
)

// Service 编排登录、注册、验证码、找回密码等账号入口流程。
type Service struct {
	cfg          config.Config
	log          *zap.Logger
	store        *data.Store
	cacheStore   cache.Store
	tokenManager *authlib.TokenManager
	emailSender  email.Sender
	smsSender    sms.Sender
}

// NewService 创建 auth page service，并组装数据库、cache、token、邮件和短信发送能力。
func NewService(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store, tokenManager *authlib.TokenManager) *Service {
	return &Service{
		cfg:          cfg,
		log:          log,
		store:        data.NewStore(db),
		cacheStore:   cacheStore,
		tokenManager: tokenManager,
		emailSender:  email.NewSender(cfg, log),
		smsSender:    sms.NewSender(cfg, log),
	}
}
