package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// EmailRegisterVerify 处理 POST /api/v1/auth/email-register/verify，只校验邮箱注册条件并签发资料完善 token。
func (h *Handler) EmailRegisterVerify(ctx *gin.Context) {
	var req proto.EmailRegisterVerifyRequest
	if !h.bind(ctx, "email_register_verify", &req) {
		return
	}
	data, err := h.service.EmailRegisterVerify(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "email_register_verify", data, err)
}
