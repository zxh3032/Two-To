package dao

const (
	// SecurityLogStatusNormal 表示安全日志处于正常可查询状态。
	SecurityLogStatusNormal int32 = 1
	// SecurityLogStatusArchived 表示安全日志已归档，默认业务查询可忽略。
	SecurityLogStatusArchived int32 = 2
)

// SecurityLog 保存账号安全事件，不记录验证码、token、密码等敏感明文。
type SecurityLog struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID        uint64 `gorm:"column:user_id" json:"userId"`
	EventType     string `gorm:"column:event_type" json:"eventType"`
	Detail        string `gorm:"column:detail" json:"detail"`
	IPHash        string `gorm:"column:ip_hash" json:"ipHash"`
	UserAgentHash string `gorm:"column:user_agent_hash" json:"userAgentHash"`
	BaseModel
}

// TableName 指定 GORM 使用 security_logs 表。
func (SecurityLog) TableName() string {
	return "security_logs"
}
