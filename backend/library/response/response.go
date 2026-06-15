package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/requestctx"
)

const (
	// CodeOK 表示请求处理成功。
	CodeOK = 0
	// CodeBadRequest 表示参数错误或业务规则不满足。
	CodeBadRequest = 40000
	// CodeUnauthorized 表示未登录或登录态失效。
	CodeUnauthorized = 40100
	// CodeForbidden 表示当前登录用户无权访问。
	CodeForbidden = 40300
	// CodeNotFound 表示请求资源不存在。
	CodeNotFound = 40400
	// CodeConflict 表示唯一键或业务状态冲突。
	CodeConflict = 40900
	// CodeRateLimited 表示请求触发频控。
	CodeRateLimited = 42900
	// CodeInternalError 表示服务端内部错误。
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

// ErrorData 输出带结构化 data 的错误响应，适合提示前端需要补验证码等可恢复状态。
func ErrorData(ctx *gin.Context, httpStatus int, code int, message string, data interface{}) {
	ctx.AbortWithStatusJSON(httpStatus, Body{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestctx.GetGinRequestID(ctx),
	})
}
