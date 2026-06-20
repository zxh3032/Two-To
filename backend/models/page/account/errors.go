package account

import (
	"net/http"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
)

// requireDB 统一拦截未配置 MySQL 的账号中心请求。
func (s *Service) requireDB() error {
	if s.store.DBReady() {
		return nil
	}
	return apperror.New(http.StatusServiceUnavailable, response.CodeInternalError, "账号服务需要配置 MySQL 后才能使用")
}

// internalError 将底层错误包装为统一的接口错误结构。
func internalError(err error) *apperror.Error {
	return apperror.Wrap(http.StatusInternalServerError, response.CodeInternalError, "服务暂时不可用", err)
}
