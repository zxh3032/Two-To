package dao

const (
	// ProfileStatusNormal 表示用户画像当前有效。
	ProfileStatusNormal int32 = 1
	// ProfileStatusDisabled 表示用户画像被停用，暂不参与后续匹配。
	ProfileStatusDisabled int32 = 2
)

// UserProfile 保存首期轻量画像，用 JSON 字段承载多选项。
type UserProfile struct {
	UserID             uint64 `gorm:"column:user_id;primaryKey" json:"userId"`
	PetStage           string `gorm:"column:pet_stage" json:"petStage"`
	InterestedPetTypes string `gorm:"column:interested_pet_types" json:"interestedPetTypes"`
	PetExperience      string `gorm:"column:pet_experience" json:"petExperience"`
	DailyCompanyTime   string `gorm:"column:daily_company_time" json:"dailyCompanyTime"`
	LivingSituation    string `gorm:"column:living_situation" json:"livingSituation"`
	PetConstraints     string `gorm:"column:pet_constraints" json:"petConstraints"`
	BaseModel
}

// TableName 指定 GORM 使用 user_profiles 表。
func (UserProfile) TableName() string {
	return "user_profiles"
}
