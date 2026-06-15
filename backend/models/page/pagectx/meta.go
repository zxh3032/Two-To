package pagectx

// RequestMeta 保存与 HTTP 相关但业务层需要用到的少量上下文信息。
type RequestMeta struct {
	IP        string
	UserAgent string
}
