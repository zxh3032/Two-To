package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// EmailLogin 处理 POST /api/v1/auth/email-login，使用邮箱和密码换取登录态。
func EmailLogin(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.EmailLoginRequest, proto.EmailLoginResponse](rt.Action("auth", "email_login"), service.EmailLogin)
}
