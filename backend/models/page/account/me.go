package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/proto"
)

// Me 查询当前登录用户的账号资料、扩展画像和已绑定联系方式。
func (s *Service) Me(ctx context.Context, userID uint64) (*proto.AccountMeResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	if user == nil {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	profile, err := s.store.FindProfile(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	identities, err := s.store.ListIdentities(ctx, userID)
	if err != nil {
		return nil, internalError(err)
	}
	return &proto.AccountMeResponse{
		User:       toUserInfo(user),
		Profile:    toProfile(profile),
		Identities: toIdentities(identities),
	}, nil
}
