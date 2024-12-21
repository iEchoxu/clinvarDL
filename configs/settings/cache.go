package settings

import "time"

// CacheSettings 定义缓存相关配置
type CacheSettings struct {
	Enabled bool          `yaml:"enabled"`  // 是否启用缓存
	Dir     string        `yaml:"dir"`      // 缓存目录
	TTL     time.Duration `yaml:"ttl"`      // 缓存过期时间
	MaxSize int64         `yaml:"max_size"` // 缓存最大大小（字节）
}

// NewCacheSettings 创建默认的缓存配置
func NewCacheSettings() *CacheSettings {
	return &CacheSettings{
		Enabled: true,          // 默认启用缓存
		Dir:     ".cache",      // 默认缓存目录
		TTL:     6 * time.Hour, // 默认6小时过期
		MaxSize: 200 << 20,     // 默认200MB
	}
}
