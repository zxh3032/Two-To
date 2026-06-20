package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// ForgotPasswordReset 处理 POST /api/v1/auth/forgot-password/reset，使用重置 token 设置新密码。
func ForgotPasswordReset(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.ForgotPasswordResetRequest, proto.ForgotPasswordResetResponse](rt.Action("auth", "forgot_password_reset"), service.ForgotPasswordReset)
}
