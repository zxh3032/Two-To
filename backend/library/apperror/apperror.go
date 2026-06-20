package apperror

import "net/http"

// Error 是 page 层返回给 controller 的业务错误，包含 HTTP 状态码和业务码。
type Error struct {
	HTTPStatus int
	Code       int
	Message    string
	Data       interface{}
	Cause      error
}

// Error 实现 error 接口，便于 controller 按业务错误统一分支处理。
func (e *Error) Error() string {
	return e.Message
}

// Unwrap 返回底层原始错误，响应层只对日志使用它，不能直接返回给前端。
func (e *Error) Unwrap() error {
	return e.Cause
}

// New 创建不带额外 data 的业务错误。
func New(httpStatus int, code int, message string) *Error {
	return &Error{HTTPStatus: httpStatus, Code: code, Message: message}
}

// Wrap 创建携带原始错误的业务错误，用于把内部错误和用户可见文案分离。
func Wrap(httpStatus int, code int, message string, cause error) *Error {
	return &Error{HTTPStatus: httpStatus, Code: code, Message: message, Cause: cause}
}

// WithData 创建带结构化 data 的业务错误，例如提示前端需要补图片验证码。
func WithData(httpStatus int, code int, message string, data interface{}) *Error {
	return &Error{HTTPStatus: httpStatus, Code: code, Message: message, Data: data}
}

// BadRequest 创建 400 类参数或业务规则错误。
func BadRequest(code int, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}

// Unauthorized 创建 401 类未登录或登录态失效错误。
func Unauthorized(code int, message string) *Error {
	return New(http.StatusUnauthorized, code, message)
}

// Forbidden 创建 403 类无权限错误。
func Forbidden(code int, message string) *Error {
	return New(http.StatusForbidden, code, message)
}

// Conflict 创建 409 类数据冲突错误。
func Conflict(code int, message string) *Error {
	return New(http.StatusConflict, code, message)
}

// RateLimited 创建 429 类频控错误。
func RateLimited(code int, message string) *Error {
	return New(http.StatusTooManyRequests, code, message)
}
