package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// ForgotPasswordVerifyCode 处理 POST /api/v1/auth/forgot-password/verify-code，验证码通过后返回重置密码 token。
func ForgotPasswordVerifyCode(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.JSON[proto.ForgotPasswordVerifyCodeRequest, proto.ForgotPasswordVerifyCodeResponse](rt.Action("auth", "forgot_password_verify_code"), service.ForgotPasswordVerifyCode)
}
