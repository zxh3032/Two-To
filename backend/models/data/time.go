package data

import "time"

// Now 返回 Unix 秒级时间戳，统一匹配表字段的 create_time/update_time 约定。
func Now() int64 {
	return time.Now().Unix()
}

// StringPtr 生成字符串指针，便于区分身份值空字符串和 NULL。
func StringPtr(value string) *string {
	return &value
}
