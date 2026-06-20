package auth

import (
	"errors"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
	"gorm.io/gorm"
)

// CompleteProfile 使用资料完善 token 创建正式账号，并在创建成功后直接登录。
func (s *Service) CompleteProfile(ctx *servlet.Context, req *proto.CompleteProfileRequest, resp *proto.CompleteProfileResponse) error {
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
	setupKey := s.cacheStore.Key("profile-setup", req.GetProfileSetupToken())
	payload, err := verification.LoadJSON[verification.SetupPayload](ctx, s.cacheStore, setupKey)
	if err != nil {
		return apperror.BadRequest(response.CodeBadRequest, "资料完善流程已过期，请重新开始")
	}
	// token payload 里只保存通过验证的联系方式和密码哈希，避免前端篡改联系方式类型。
	var identityType int32
	var identityValue string
	if payload.Kind == identityTypeEmailText {
		identityType = dao.IdentityTypeEmail
		identityValue = payload.Email
	} else if payload.Kind == identityTypePhoneText {
		identityType = dao.IdentityTypePhone
		identityValue = payload.Phone
	} else {
		return apperror.BadRequest(response.CodeBadRequest, "资料完善流程不正确")
	}
	existing, err := s.store.FindIdentity(ctx, identityType, identityValue)
	if err != nil {
		return internalError(err)
	}
	if existing != nil {
		return apperror.Conflict(response.CodeConflict, "该联系方式已被注册，请直接登录")
	}
	now := data.Now()
	// users、user_auth_identities、user_profiles 必须在同一事务创建，任何一步失败都回滚。
	user := &dao.User{
		Nickname:           nickname,
		PasswordHash:       payload.PasswordHash,
		PasswordUpdateTime: passwordUpdateTime(payload.PasswordHash, now),
		BaseModel:          dao.NewBaseModel(dao.UserStatusNormal, now),
	}
	identity := &dao.UserAuthIdentity{
		IdentityType:  identityType,
		IdentityValue: data.StringPtr(identityValue),
		VerifyTime:    now,
		BaseModel:     dao.NewBaseModel(dao.IdentityStatusNormal, now),
	}
	profile := buildProfile(0, req.GetPetStage(), req.GetInterestedPetTypes(), req.GetPetExperience(), req.GetDailyCompanyTime(), req.GetLivingSituation(), req.GetPetConstraints(), now)
	if err := s.store.CreateUserWithIdentityAndProfile(ctx, user, identity, profile); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperror.Conflict(response.CodeConflict, "该联系方式已被注册，请直接登录")
		}
		return internalError(err)
	}
	// 创建成功后立刻删除资料完善 token，避免同一 token 重复创建账号。
	_ = s.cacheStore.Del(ctx, setupKey)
	s.logSecurity(ctx, user.ID, "complete_profile", map[string]string{"identityType": payload.Kind})
	authResp, err := s.issueLogin(ctx, user)
	if err != nil {
		return err
	}
	fillCompleteProfileResponse(resp, authResp)
	return nil
}
