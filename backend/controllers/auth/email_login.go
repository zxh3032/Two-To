package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// EmailLogin 处理 POST /api/v1/auth/email-login，使用邮箱和密码换取登录态。
func (h *Handler) EmailLogin(ctx *gin.Context) {
	var req proto.EmailLoginRequest
	if !h.bind(ctx, "email_login", &req) {
		return
	}
	data, err := h.service.EmailLogin(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "email_login", data, err)
}
