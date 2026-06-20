package auth

import (
	"net/http"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
)

// requireDB 检查账号接口是否具备 MySQL。没有数据库时只允许健康检查和验证码等非账号事实写入能力。
func (s *Service) requireDB() error {
	if s.store.DBReady() {
		return nil
	}
	return apperror.New(http.StatusServiceUnavailable, response.CodeInternalError, "账号服务需要配置 MySQL 后才能使用")
}

// internalError 把底层错误包装成统一业务错误，controller 会负责转换成 HTTP 响应。
func internalError(err error) *apperror.Error {
	return apperror.Wrap(http.StatusInternalServerError, response.CodeInternalError, "服务暂时不可用", err)
}
