package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/proto"
)

// UnbindEmail 处理 DELETE /api/v1/account/security/email，解绑邮箱前需要校验当前密码。
func (h *Handler) UnbindEmail(ctx *gin.Context) {
	var req proto.CurrentPasswordRequest
	if !h.bind(ctx, "unbind_email", &req) {
		return
	}
	h.write(ctx, "unbind_email", nil, h.service.UnbindEmail(ctx.Request.Context(), authlib.GetGinUserID(ctx), req.GetCurrentPassword(), meta(ctx)))
}
