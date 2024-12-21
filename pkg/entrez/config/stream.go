package config

// StreamConfig 定义流式处理的配置
type StreamConfig struct {
	// 是否启用流式处理
	Enabled bool
}

// DefaultStreamConfig 返回默认的流式处理配置
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		Enabled: false, // 默认不启用流式处理
	}
}

// GetEnabled 获取是否启用流式处理
func (c *StreamConfig) GetEnabled() bool {
	return c.Enabled
}
