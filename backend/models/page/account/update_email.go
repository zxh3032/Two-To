package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
)

// UpdateEmail 绑定或换绑邮箱，验证码校验成功后才更新身份表。
func (s *Service) UpdateEmail(ctx context.Context, userID uint64, req *proto.UpdateEmailRequest, meta pagectx.RequestMeta) error {
	emailValue, err := security.NormalizeEmail(req.GetEmail())
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := s.verifyCode(ctx, verification.CodeTypeEmail, verification.SceneBindEmail, emailValue, req.GetEmailCode()); err != nil {
		return err
	}
	return s.updateIdentity(ctx, userID, dao.IdentityTypeEmail, emailValue, meta)
}
