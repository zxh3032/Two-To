package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// ForgotPasswordReset 处理 POST /api/v1/auth/forgot-password/reset，使用重置 token 设置新密码。
func (h *Handler) ForgotPasswordReset(ctx *gin.Context) {
	var req proto.ForgotPasswordResetRequest
	if !h.bind(ctx, "forgot_password_reset", &req) {
		return
	}
	h.write(ctx, "forgot_password_reset", nil, h.service.ForgotPasswordReset(ctx.Request.Context(), &req, meta(ctx)))
}
