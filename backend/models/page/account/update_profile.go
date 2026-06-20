package account

import (
	"errors"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
	"gorm.io/gorm"
)

// UpdateProfile 更新昵称和养宠画像，供注册后补充资料以及个人资料页复用。
func (s *Service) UpdateProfile(ctx *servlet.Context, req *proto.UpdateProfileRequest, resp *proto.UpdateProfileResponse) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	nickname, err := validateNickname(req.GetNickname())
	if err != nil {
		return err
	}
	if err := validateProfileFields(req.GetPetStage(), req.GetInterestedPetTypes()); err != nil {
		return err
	}
	now := data.Now()
	// 昵称在 users 表，画像在 user_profiles 表，使用同一个事务保证资料更新一致。
	err = s.store.WithTx(ctx, func(tx *data.Store) error {
		user, err := tx.FindUser(ctx, ctx.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return gorm.ErrRecordNotFound
		}
		user.Nickname = nickname
		user.UpdateTime = now
		if err := tx.UpdateUserProfileFields(ctx, user, buildProfile(ctx.UserID, req.GetPetStage(), req.GetInterestedPetTypes(), req.GetPetExperience(), req.GetDailyCompanyTime(), req.GetLivingSituation(), req.GetPetConstraints(), now)); err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if err != nil {
		return internalError(err)
	}
	s.logSecurity(ctx, ctx.UserID, "update_profile", map[string]string{})
	user, profile, identities, err := s.accountSnapshot(ctx, ctx.UserID)
	if err != nil {
		return err
	}
	resp.User = user
	resp.Profile = profile
	resp.Identities = identities
	return nil
}
