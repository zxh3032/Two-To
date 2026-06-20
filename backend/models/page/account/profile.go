package account

import (
	"strings"
	"unicode/utf8"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// validateNickname 校验昵称长度，昵称允许重复，系统内部通过 user_id 区分用户。
func validateNickname(nickname string) (string, error) {
	value := strings.TrimSpace(nickname)
	count := utf8.RuneCountInString(value)
	if count < 2 || count > 20 {
		return "", apperror.BadRequest(response.CodeBadRequest, "昵称需为 2-20 个字符")
	}
	return value, nil
}

// validateProfileFields 校验后续匹配依赖的必填画像字段。
func validateProfileFields(petStage string, interestedPetTypes []string) error {
	if strings.TrimSpace(petStage) == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请选择当前养宠阶段")
	}
	if len(interestedPetTypes) == 0 {
		return apperror.BadRequest(response.CodeBadRequest, "请至少选择一种感兴趣的宠物")
	}
	return nil
}

// buildProfile 将前端提交的画像字段转换为 user_profiles 表结构。
func buildProfile(userID uint64, petStage string, interestedPetTypes []string, petExperience string, dailyCompanyTime string, livingSituation string, petConstraints []string, now int64) *dao.UserProfile {
	return &dao.UserProfile{
		UserID:             userID,
		PetStage:           strings.TrimSpace(petStage),
		InterestedPetTypes: data.JSONStrings(interestedPetTypes),
		PetExperience:      strings.TrimSpace(petExperience),
		DailyCompanyTime:   strings.TrimSpace(dailyCompanyTime),
		LivingSituation:    strings.TrimSpace(livingSituation),
		PetConstraints:     data.JSONStrings(petConstraints),
		BaseModel:          dao.NewBaseModel(dao.ProfileStatusNormal, now),
	}
}

// toUserInfo 将用户主表模型转换为接口响应结构。
func toUserInfo(user *dao.User) *proto.UserInfo {
	if user == nil {
		return nil
	}
	return &proto.UserInfo{Id: int64(user.ID), Nickname: user.Nickname, Status: user.Status}
}

// toProfile 将画像模型转换为接口响应结构，兼容用户尚未填写画像的情况。
func toProfile(profile *dao.UserProfile) *proto.UserProfile {
	if profile == nil {
		return &proto.UserProfile{}
	}
	return &proto.UserProfile{
		PetStage:           profile.PetStage,
		InterestedPetTypes: data.ParseJSONStringArray(profile.InterestedPetTypes),
		PetExperience:      profile.PetExperience,
		DailyCompanyTime:   profile.DailyCompanyTime,
		LivingSituation:    profile.LivingSituation,
		PetConstraints:     data.ParseJSONStringArray(profile.PetConstraints),
	}
}

// toIdentities 将绑定身份转换为响应结构，同时补充脱敏展示值。
func toIdentities(items []dao.UserAuthIdentity) []*proto.UserIdentity {
	result := make([]*proto.UserIdentity, 0, len(items))
	for _, item := range items {
		value := ""
		if item.IdentityValue != nil {
			value = *item.IdentityValue
		}
		identityType := "phone"
		if item.IdentityType == dao.IdentityTypeEmail {
			identityType = "email"
		}
		result = append(result, &proto.UserIdentity{
			Id:            int64(item.ID),
			IdentityType:  identityType,
			IdentityValue: value,
			MaskedValue:   security.MaskIdentity(item.IdentityType, value),
			VerifyTime:    item.VerifyTime,
		})
	}
	return result
}
