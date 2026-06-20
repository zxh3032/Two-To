package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
)

// Sessions 处理 GET /api/v1/account/sessions，列出当前账号全部登录设备。
func (h *Handler) Sessions(ctx *gin.Context) {
	data, err := h.service.Sessions(ctx.Request.Context(), authlib.GetGinUserID(ctx), authlib.GetGinSessionID(ctx))
	h.write(ctx, "sessions", data, err)
}
