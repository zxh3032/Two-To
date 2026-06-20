package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
)

// UnbindEmail 解绑邮箱，底层会校验当前密码并确保账号仍保留至少一种联系方式。
func (s *Service) UnbindEmail(ctx context.Context, userID uint64, currentPassword string, meta pagectx.RequestMeta) error {
	return s.unbindIdentity(ctx, userID, dao.IdentityTypeEmail, currentPassword, meta)
}
