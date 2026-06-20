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

// UpdatePhone 绑定或换绑中国大陆手机号，当前版本只接受 +86 手机号。
func (s *Service) UpdatePhone(ctx context.Context, userID uint64, req *proto.UpdatePhoneRequest, meta pagectx.RequestMeta) error {
	phone, err := security.NormalizeMainlandPhone(req.GetPhone())
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, err.Error())
	}
	if err := s.verifyCode(ctx, verification.CodeTypeSMS, verification.SceneBindPhone, phone, req.GetSmsCode()); err != nil {
		return err
	}
	return s.updateIdentity(ctx, userID, dao.IdentityTypePhone, phone, meta)
}
