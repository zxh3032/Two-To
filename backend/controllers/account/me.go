package account

import (
	"github.com/gin-gonic/gin"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
)

// Me 处理 GET /api/v1/account/me，返回当前用户、画像和联系方式。
func (h *Handler) Me(ctx *gin.Context) {
	data, err := h.service.Me(ctx.Request.Context(), authlib.GetGinUserID(ctx))
	h.write(ctx, "me", data, err)
}
