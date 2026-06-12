package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/requestctx"
)

const (
	CodeOK            = 0
	CodeInternalError = 50000
)

// Body 是后端统一响应结构，requestId 用于联调和线上日志排查。
type Body struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// Success 输出成功响应，业务数据统一放在 data 字段中。
func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Body{
		Code:      CodeOK,
		Message:   "success",
		Data:      data,
		RequestID: requestctx.GetGinRequestID(ctx),
	})
}

// Error 输出失败响应，HTTP 状态码和业务错误码都需要调用方明确传入。
func Error(ctx *gin.Context, httpStatus int, code int, message string) {
	ctx.AbortWithStatusJSON(httpStatus, Body{
		Code:      code,
		Message:   message,
		RequestID: requestctx.GetGinRequestID(ctx),
	})
}
