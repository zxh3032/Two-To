package auth

import (
	"strings"
	"unicode/utf8"

	"github.com/zxh3032/two-to/backend/library/apperror"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// validateNickname 校验昵称长度。昵称允许重复，系统内部以用户 ID 区分用户。
func validateNickname(nickname string) (string, error) {
	value := strings.TrimSpace(nickname)
	count := utf8.RuneCountInString(value)
	if count < 2 || count > 20 {
		return "", apperror.BadRequest(response.CodeBadRequest, "昵称需为 2-20 个字符")
	}
	return value, nil
}

// validateProfileFields 校验资料完善页必填项：养宠阶段和至少一种感兴趣宠物。
func validateProfileFields(petStage string, interestedPetTypes []string) error {
	if strings.TrimSpace(petStage) == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请选择当前养宠阶段")
	}
	if len(interestedPetTypes) == 0 {
		return apperror.BadRequest(response.CodeBadRequest, "请至少选择一种感兴趣的宠物")
	}
	return nil
}

// buildProfile 把前端资料完善/资料编辑入参转换为 user_profiles DAO。
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

// passwordUpdateTime 根据是否存在密码哈希决定密码更新时间。手机号验证码注册时密码可以为空。
func passwordUpdateTime(passwordHash string, now int64) int64 {
	if passwordHash == "" {
		return 0
	}
	return now
}

// toUserInfo 将 users DAO 转成接口返回结构，避免把 password_hash 等字段暴露给前端。
func toUserInfo(user *dao.User) *proto.UserInfo {
	if user == nil {
		return nil
	}
	return &proto.UserInfo{
		Id:       int64(user.ID),
		Nickname: user.Nickname,
		Status:   user.Status,
	}
}
