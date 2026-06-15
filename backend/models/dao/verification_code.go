package dao

const (
	// VerificationStatusUnused 表示验证码已发送但尚未成功使用。
	VerificationStatusUnused int32 = 1
	// VerificationStatusUsed 表示验证码已成功使用，不能再次复用。
	VerificationStatusUsed int32 = 2
	// VerificationStatusExpired 表示验证码超过有效期。
	VerificationStatusExpired int32 = 3
	// VerificationStatusRevoked 表示验证码因安全策略被撤销。
	VerificationStatusRevoked int32 = 4
)

// VerificationCode 是验证码审计表，明文验证码只短期存放在 cache。
type VerificationCode struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CodeType      int32  `gorm:"column:code_type" json:"codeType"`
	Scene         string `gorm:"column:scene" json:"scene"`
	Target        string `gorm:"column:target" json:"target"`
	CodeHash      string `gorm:"column:code_hash" json:"-"`
	AttemptCount  int32  `gorm:"column:attempt_count" json:"attemptCount"`
	ExpireTime    int64  `gorm:"column:expire_time" json:"expireTime"`
	UseTime       int64  `gorm:"column:use_time" json:"useTime"`
	IPHash        string `gorm:"column:ip_hash" json:"ipHash"`
	UserAgentHash string `gorm:"column:user_agent_hash" json:"userAgentHash"`
	Provider      string `gorm:"column:provider" json:"provider"`
	BaseModel
}

// TableName 指定 GORM 使用 verification_codes 表。
func (VerificationCode) TableName() string {
	return "verification_codes"
}
