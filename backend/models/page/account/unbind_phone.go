package account

import (
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/proto"
)

// UnbindPhone 解绑手机号，底层会校验当前密码并确保账号仍保留至少一种联系方式。
func (s *Service) UnbindPhone(ctx *servlet.Context, req *proto.UnbindPhoneRequest, _ *proto.UnbindPhoneResponse) error {
	return s.unbindIdentity(ctx, ctx.UserID, dao.IdentityTypePhone, req.GetCurrentPassword())
}
