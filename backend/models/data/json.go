package data

import "encoding/json"

// JSONStrings 把字符串数组转成 JSON 字符串，供 GORM 写入 MySQL JSON 字段。
func JSONStrings(values []string) string {
	raw, _ := json.Marshal(values)
	return string(raw)
}

// ParseJSONStringArray 把 MySQL JSON 字段解析为字符串数组，解析失败时按空值处理。
func ParseJSONStringArray(raw string) []string {
	if raw == "" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return values
}
