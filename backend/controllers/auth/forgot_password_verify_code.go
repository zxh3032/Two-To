package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// ForgotPasswordVerifyCode 处理 POST /api/v1/auth/forgot-password/verify-code，验证码通过后返回重置密码 token。
func (h *Handler) ForgotPasswordVerifyCode(ctx *gin.Context) {
	var req proto.ForgotPasswordVerifyCodeRequest
	if !h.bind(ctx, "forgot_password_verify_code", &req) {
		return
	}
	data, err := h.service.ForgotPasswordVerifyCode(ctx.Request.Context(), &req)
	h.write(ctx, "forgot_password_verify_code", data, err)
}
