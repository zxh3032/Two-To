package auth

import "github.com/gin-gonic/gin"

// Captcha 处理 GET /api/v1/auth/captcha，返回图片验证码 ID、base64 图片和有效期。
func (h *Handler) Captcha(ctx *gin.Context) {
	data, err := h.service.Captcha(ctx.Request.Context())
	h.write(ctx, "captcha", data, err)
}
