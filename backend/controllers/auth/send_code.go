package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// SendCode 处理 POST /api/v1/auth/send-code，发送短信或邮箱验证码。
func (h *Handler) SendCode(ctx *gin.Context) {
	var req proto.SendCodeRequest
	if !h.bind(ctx, "send_code", &req) {
		return
	}
	data, err := h.service.SendCode(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "send_code", data, err)
}
