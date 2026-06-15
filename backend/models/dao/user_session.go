package dao

const (
	// SessionStatusNormal 表示设备会话可继续刷新 token。
	SessionStatusNormal int32 = 1
	// SessionStatusLogout 表示用户主动退出当前设备。
	SessionStatusLogout int32 = 2
	// SessionStatusRevoked 表示会话被改密、踢设备或安全策略撤销。
	SessionStatusRevoked int32 = 3
	// SessionStatusExpired 表示会话自然过期。
	SessionStatusExpired int32 = 4
)

// UserSession 映射登录设备和 refresh token 事实表。
type UserSession struct {
	ID               uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID           uint64 `gorm:"column:user_id" json:"userId"`
	RefreshTokenHash string `gorm:"column:refresh_token_hash" json:"-"`
	DeviceName       string `gorm:"column:device_name" json:"deviceName"`
	UserAgentHash    string `gorm:"column:user_agent_hash" json:"userAgentHash"`
	IPHash           string `gorm:"column:ip_hash" json:"ipHash"`
	LastActiveTime   int64  `gorm:"column:last_active_time" json:"lastActiveTime"`
	ExpireTime       int64  `gorm:"column:expire_time" json:"expireTime"`
	RevokeTime       int64  `gorm:"column:revoke_time" json:"revokeTime"`
	RevokeReason     string `gorm:"column:revoke_reason" json:"revokeReason"`
	BaseModel
}

// TableName 指定 GORM 使用 user_sessions 表。
func (UserSession) TableName() string {
	return "user_sessions"
}
