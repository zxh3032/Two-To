package account

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"github.com/zxh3032/two-to/backend/proto"
	"gorm.io/gorm"
)

// UpdateProfile 更新昵称和养宠画像，供注册后补充资料以及个人资料页复用。
func (s *Service) UpdateProfile(ctx context.Context, userID uint64, req *proto.UpdateProfileRequest, meta pagectx.RequestMeta) (*proto.AccountMeResponse, error) {
	if err := s.requireDB(); err != nil {
		return nil, err
	}
	nickname, err := validateNickname(req.GetNickname())
	if err != nil {
		return nil, err
	}
	if err := validateProfileFields(req.GetPetStage(), req.GetInterestedPetTypes()); err != nil {
		return nil, err
	}
	now := data.Now()
	// 昵称在 users 表，画像在 user_profiles 表，使用同一个事务保证资料更新一致。
	err = s.store.WithTx(ctx, func(tx *data.Store) error {
		user, err := tx.FindUser(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return gorm.ErrRecordNotFound
		}
		user.Nickname = nickname
		user.UpdateTime = now
		if err := tx.UpdateUserProfileFields(ctx, user, buildProfile(userID, req.GetPetStage(), req.GetInterestedPetTypes(), req.GetPetExperience(), req.GetDailyCompanyTime(), req.GetLivingSituation(), req.GetPetConstraints(), now)); err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if err != nil {
		return nil, internalError(err)
	}
	s.logSecurity(ctx, userID, "update_profile", map[string]string{}, meta)
	return s.Me(ctx, userID)
}
