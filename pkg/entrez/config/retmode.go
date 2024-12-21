package config

const (
	RetModeXML  RetMode = "xml"  // XML 格式
	RetModeJSON RetMode = "json" // JSON 格式
)

// RetMode 定义了返回数据的格式类型
type RetMode string

// String 实现 Stringer 接口
func (r RetMode) String() string {
	return string(r)
}

// IsValid 检查 RetMode 是否有效
func (r RetMode) IsValid() bool {
	switch r {
	case RetModeXML, RetModeJSON:
		return true
	default:
		return false
	}
}
