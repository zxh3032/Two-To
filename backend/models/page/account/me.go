package account

import (
	"context"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/proto"
)

// Me 查询当前登录用户的账号资料、扩展画像和已绑定联系方式。
func (s *Service) Me(ctx *servlet.Context, _ *proto.AccountMeRequest, resp *proto.AccountMeResponse) error {
	user, profile, identities, err := s.accountSnapshot(ctx, ctx.UserID)
	if err != nil {
		return err
	}
	resp.User = user
	resp.Profile = profile
	resp.Identities = identities
	return nil
}

func (s *Service) accountSnapshot(ctx context.Context, userID uint64) (*proto.UserInfo, *proto.UserProfile, []*proto.UserIdentity, error) {
	if err := s.requireDB(); err != nil {
		return nil, nil, nil, err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return nil, nil, nil, internalError(err)
	}
	if user == nil {
		return nil, nil, nil, apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	profile, err := s.store.FindProfile(ctx, userID)
	if err != nil {
		return nil, nil, nil, internalError(err)
	}
	identities, err := s.store.ListIdentities(ctx, userID)
	if err != nil {
		return nil, nil, nil, internalError(err)
	}
	return toUserInfo(user), toProfile(profile), toIdentities(identities), nil
}
