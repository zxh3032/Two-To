package dao

const (
	// LoginAttemptStatusNormal 表示登录失败审计记录正常有效。
	LoginAttemptStatusNormal int32 = 1
	// LoginAttemptStatusClear 表示失败计数已被登录成功或安全动作清理。
	LoginAttemptStatusClear int32 = 2
)

// LoginAttempt 保留登录失败计数的审计记录，实时锁定状态以 cache 为准。
type LoginAttempt struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	IdentityType  int32  `gorm:"column:identity_type" json:"identityType"`
	IdentityValue string `gorm:"column:identity_value" json:"identityValue"`
	IPHash        string `gorm:"column:ip_hash" json:"ipHash"`
	FailCount     int32  `gorm:"column:fail_count" json:"failCount"`
	LockUntilTime int64  `gorm:"column:lock_until_time" json:"lockUntilTime"`
	LastFailTime  int64  `gorm:"column:last_fail_time" json:"lastFailTime"`
	BaseModel
}

// TableName 指定 GORM 使用 login_attempts 表。
func (LoginAttempt) TableName() string {
	return "login_attempts"
}
