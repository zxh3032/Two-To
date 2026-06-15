package dao

const (
	// IdentityTypeEmail 表示邮箱登录身份。
	IdentityTypeEmail int32 = 1
	// IdentityTypePhone 表示手机号登录身份。
	IdentityTypePhone int32 = 2

	// IdentityStatusNormal 表示联系方式当前可用于登录、找回密码和安全验证。
	IdentityStatusNormal int32 = 1
	// IdentityStatusUnbound 表示联系方式已解绑，保留历史记录但不再参与登录。
	IdentityStatusUnbound int32 = 2
)

// UserAuthIdentity 统一映射邮箱、手机号等可登录身份。
type UserAuthIdentity struct {
	ID            uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID        uint64  `gorm:"column:user_id" json:"userId"`
	IdentityType  int32   `gorm:"column:identity_type" json:"identityType"`
	IdentityValue *string `gorm:"column:identity_value" json:"identityValue"`
	VerifyTime    int64   `gorm:"column:verify_time" json:"verifyTime"`
	BaseModel
}

// TableName 指定 GORM 使用 user_auth_identities 表。
func (UserAuthIdentity) TableName() string {
	return "user_auth_identities"
}
