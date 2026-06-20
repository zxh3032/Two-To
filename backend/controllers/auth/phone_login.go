package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/proto"
)

// PhoneLogin 处理 POST /api/v1/auth/phone-login，手机号验证码登录和注册入口共用该接口。
func (h *Handler) PhoneLogin(ctx *gin.Context) {
	var req proto.PhoneLoginRequest
	if !h.bind(ctx, "phone_login", &req) {
		return
	}
	data, err := h.service.PhoneLogin(ctx.Request.Context(), &req, meta(ctx))
	h.write(ctx, "phone_login", data, err)
}
