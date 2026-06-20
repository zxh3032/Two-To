package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/servlet"
	pageAuth "github.com/zxh3032/two-to/backend/models/page/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// Captcha 处理 GET /api/v1/auth/captcha，返回图片验证码 ID、base64 图片和有效期。
func Captcha(rt servlet.Runtime) gin.HandlerFunc {
	service := pageAuth.NewService(rt.Config, rt.Log, rt.DB, rt.Cache, rt.TokenManager)
	return servlet.Empty[proto.CaptchaRequest, proto.CaptchaResponse](rt.Action("auth", "captcha"), service.Captcha)
}
