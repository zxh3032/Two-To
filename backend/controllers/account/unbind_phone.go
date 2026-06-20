package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// UnbindPhone 处理 DELETE /api/v1/account/security/phone，解绑手机号前需要校验当前密码。
func (h *Handler) UnbindPhone(ctx *gin.Context) {
	var req proto.CurrentPasswordRequest
	if !h.bind(ctx, "unbind_phone", &req) {
		return
	}
	h.write(ctx, "unbind_phone", nil, h.service.UnbindPhone(ctx.Request.Context(), authlib.GetGinUserID(ctx), req.GetCurrentPassword(), meta(ctx)))
}
