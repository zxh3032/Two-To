package dao

const (
	// StatusNormal 表示业务记录处于正常可用状态。
	StatusNormal int32 = 1
	// StatusDisabled 表示业务记录被停用，数据仍保留用于审计和兼容历史引用。
	StatusDisabled int32 = 2
)

// BaseModel 是所有业务表复用的公共字段，不使用 deleted_at/delete_time。
type BaseModel struct {
	Status     int32 `gorm:"column:status" json:"status"`
	CreateTime int64 `gorm:"column:create_time" json:"createTime"`
	UpdateTime int64 `gorm:"column:update_time" json:"updateTime"`
}

// NewBaseModel 初始化公共字段，创建时间和更新时间在新增时保持一致。
func NewBaseModel(status int32, now int64) BaseModel {
	return BaseModel{
		Status:     status,
		CreateTime: now,
		UpdateTime: now,
	}
}

// Touch 刷新更新时间，用于逻辑删除、状态变更和普通资料更新。
func (m *BaseModel) Touch(now int64) {
	m.UpdateTime = now
}
