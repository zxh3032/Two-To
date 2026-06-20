package servlet

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
)

// Handler 是统一的 page 执行函数签名：上下文、请求、响应显式传入。
type Handler[Req any, Resp any] func(*Context, *Req, *Resp) error

// Decoder 把 Gin 请求中的 body、path、query 等数据写入 request。
type Decoder[Req any] func(*gin.Context, *Req) error

// JSON 适用于需要 JSON body 的接口。
func JSON[Req any, Resp any](action Action, handle Handler[Req, Resp]) gin.HandlerFunc {
	return Exec(action, BindJSON[Req], handle)
}

// Empty 适用于无请求体的接口，仍然会创建对应 request。
func Empty[Req any, Resp any](action Action, handle Handler[Req, Resp]) gin.HandlerFunc {
	return Exec(action, Noop[Req], handle)
}

// Path 适用于只从 path 或其他非 body 来源解析请求参数的接口。
func Path[Req any, Resp any](action Action, decode Decoder[Req], handle Handler[Req, Resp]) gin.HandlerFunc {
	return Exec(action, decode, handle)
}

// Exec 执行统一 HTTP 入口协议：创建 request/response、解析请求、调用 page、输出响应。
func Exec[Req any, Resp any](action Action, decode Decoder[Req], handle Handler[Req, Resp]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := new(Req)
		responseBody := new(Resp)

		if err := decode(ctx, request); err != nil {
			response.WriteResult(ctx, action.Log, action.Module, action.Name, responseBody, err)
			return
		}

		err := handle(newContext(ctx), request, responseBody)
		response.WriteResult(ctx, action.Log, action.Module, action.Name, responseBody, err)
	}
}

// BindJSON 将 JSON 请求体绑定到 request，绑定失败时返回统一参数错误。
func BindJSON[Req any](ctx *gin.Context, request *Req) error {
	if err := ctx.ShouldBindJSON(request); err != nil {
		return apperror.Wrap(http.StatusBadRequest, response.CodeBadRequest, "请求参数格式不正确", err)
	}
	return nil
}

// Noop 保留 request 创建动作，但不从 HTTP 请求中解析额外参数。
func Noop[Req any](_ *gin.Context, _ *Req) error {
	return nil
}

// Uint64Path 解析正整数 path 参数并写入 request。
func Uint64Path[Req any](name string, message string, set func(*Req, uint64)) Decoder[Req] {
	return func(ctx *gin.Context, request *Req) error {
		value, err := strconv.ParseUint(ctx.Param(name), 10, 64)
		if err != nil || value == 0 {
			if err == nil {
				err = fmt.Errorf("%s is zero", name)
			}
			return apperror.Wrap(http.StatusBadRequest, response.CodeBadRequest, message, err)
		}
		set(request, value)
		return nil
	}
}
