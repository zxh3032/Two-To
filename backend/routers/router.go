package routers

import (
	"github.com/gin-gonic/gin"
	accountController "github.com/zxh3032/two-to/backend/controllers/account"
	authController "github.com/zxh3032/two-to/backend/controllers/auth"
	"github.com/zxh3032/two-to/backend/controllers/health"
	"github.com/zxh3032/two-to/backend/controllers/ping"
	"github.com/zxh3032/two-to/backend/controllers/slogan"
	authLibrary "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/middlewares"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewRouter 注册后端所有 HTTP 路由和通用中间件。
func NewRouter(cfg config.Config, log *zap.Logger, db *gorm.DB, cacheStore cache.Store) *gin.Engine {
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
	{
		apiV1.GET("/ping", ping.Ping(cfg, log))
		apiV1.GET("/slogan", slogan.Get(log))

		tokenManager := authLibrary.NewTokenManager(cfg.Token)

		// auth 分组是未登录也可以访问的账号入口，覆盖验证码、登录、注册、找回密码和 token 刷新。
		authGroup := apiV1.Group("/auth")
		{
			authHandler := authController.NewHandler(cfg, log, db, cacheStore, tokenManager)
			// GET /api/v1/auth/captcha 生成图片验证码，供发送短信/邮件验证码和登录风控使用。
			authGroup.GET("/captcha", authHandler.Captcha)
			// POST /api/v1/auth/send-code 发送短信或邮箱 6 位验证码，发送前必须校验图片验证码。
			authGroup.POST("/send-code", authHandler.SendCode)
			// POST /api/v1/auth/email-login 使用邮箱和密码登录，失败达到阈值后要求图片验证码。
			authGroup.POST("/email-login", authHandler.EmailLogin)
			// POST /api/v1/auth/phone-login 使用 +86 手机号验证码登录；新手机号会进入资料完善流程。
			authGroup.POST("/phone-login", authHandler.PhoneLogin)
			// POST /api/v1/auth/email-register/verify 校验邮箱注册验证码和密码，只签发资料完善 token，不直接创建账号。
			authGroup.POST("/email-register/verify", authHandler.EmailRegisterVerify)
			// POST /api/v1/auth/complete-profile 提交昵称和画像资料，完成账号创建并直接登录。
			authGroup.POST("/complete-profile", authHandler.CompleteProfile)
			// POST /api/v1/auth/refresh 使用 refresh token 轮换并签发新的 access token。
			authGroup.POST("/refresh", authHandler.Refresh)
			// POST /api/v1/auth/logout 退出当前 refresh token 对应的设备会话。
			authGroup.POST("/logout", authHandler.Logout)
			// POST /api/v1/auth/forgot-password/verify-code 校验找回密码验证码并签发重置密码 token。
			authGroup.POST("/forgot-password/verify-code", authHandler.ForgotPasswordVerifyCode)
			// POST /api/v1/auth/forgot-password/reset 使用重置密码 token 设置新密码并撤销历史会话。
			authGroup.POST("/forgot-password/reset", authHandler.ForgotPasswordReset)
		}

		// account 分组必须先通过 access token 鉴权，接口只允许操作当前登录用户自己的账号数据。
		accountGroup := apiV1.Group("/account", middlewares.Auth(tokenManager, cacheStore, db, log))
		{
			accountHandler := accountController.NewHandler(cfg, log, db, cacheStore, tokenManager)
			// GET /api/v1/account/me 返回当前用户、画像和已绑定联系方式。
			accountGroup.GET("/me", accountHandler.Me)
			// PATCH /api/v1/account/profile 更新昵称和基础画像资料。
			accountGroup.PATCH("/profile", accountHandler.UpdateProfile)
			// PUT /api/v1/account/security/email 绑定或更换邮箱。
			accountGroup.PUT("/security/email", accountHandler.UpdateEmail)
			// PUT /api/v1/account/security/phone 绑定或更换中国大陆手机号。
			accountGroup.PUT("/security/phone", accountHandler.UpdatePhone)
			// DELETE /api/v1/account/security/email 解绑邮箱，解绑前校验当前密码并保证至少保留一种联系方式。
			accountGroup.DELETE("/security/email", accountHandler.UnbindEmail)
			// DELETE /api/v1/account/security/phone 解绑手机号，解绑前校验当前密码并保证至少保留一种联系方式。
			accountGroup.DELETE("/security/phone", accountHandler.UnbindPhone)
			// PUT /api/v1/account/security/password 修改密码，成功后撤销其他设备会话。
			accountGroup.PUT("/security/password", accountHandler.UpdatePassword)
			// GET /api/v1/account/sessions 查询当前账号所有设备会话。
			accountGroup.GET("/sessions", accountHandler.Sessions)
			// DELETE /api/v1/account/sessions/:sessionId 踢出指定非当前设备。
			accountGroup.DELETE("/sessions/:sessionId", accountHandler.RevokeSession)
			// POST /api/v1/account/sessions/revoke-all 退出全部设备，适合用户发现账号异常时使用。
			accountGroup.POST("/sessions/revoke-all", accountHandler.RevokeAllSessions)
		}
	}

	return router
}
