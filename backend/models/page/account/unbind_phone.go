package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
)

// UnbindPhone 解绑手机号，底层会校验当前密码并确保账号仍保留至少一种联系方式。
func (s *Service) UnbindPhone(ctx context.Context, userID uint64, currentPassword string, meta pagectx.RequestMeta) error {
	return s.unbindIdentity(ctx, userID, dao.IdentityTypePhone, currentPassword, meta)
}
