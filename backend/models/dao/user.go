package dao

const (
	// UserStatusNormal 表示账号可正常登录和使用。
	UserStatusNormal int32 = 1
	// UserStatusLocked 表示账号被临时锁定，通常由风控或安全策略触发。
	UserStatusLocked int32 = 2
	// UserStatusBanned 表示账号被封禁，不允许继续登录。
	UserStatusBanned int32 = 3
)

// User 映射 users 表，主表只保存账号稳定信息。
type User struct {
	ID                 uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Nickname           string `gorm:"column:nickname" json:"nickname"`
	PasswordHash       string `gorm:"column:password_hash" json:"-"`
	PasswordUpdateTime int64  `gorm:"column:password_update_time" json:"passwordUpdateTime"`
	BaseModel
}

// TableName 指定 GORM 使用 users 表，避免结构体名自动复数化带来的歧义。
func (User) TableName() string {
	return "users"
}
