package config

import (
	"fmt"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"os"
	"path/filepath"
	"time"
)

// CacheConfig 定义缓存配置
type CacheConfig struct {
	Enabled bool          // 是否启用缓存
	Dir     string        // 缓存目录
	TTL     time.Duration // 缓存过期时间
	MaxSize int64         // 缓存最大大小（字节）
}

// NewCacheConfig 创建一个新的缓存配置
func NewCacheConfig(enabled bool, dir string, ttl time.Duration, maxSize int64) *CacheConfig {
	return &CacheConfig{
		Enabled: enabled,
		Dir:     dir,
		TTL:     ttl,
		MaxSize: maxSize,
	}
}

// validateCache 验证缓存配置
func (c *CacheConfig) validateCache() error {
	// 如果缓存未启用，则不进行验证
	if !c.Enabled {
		return nil
	}

	// 验证缓存目录
	if c.Dir == "" {
		return customerrors.NewParametersError("cache directory is required when cache is enabled")
	}

	// 验证目录是否存在或可创建
	if err := os.MkdirAll(c.Dir, 0755); err != nil {
		return customerrors.NewParametersError(fmt.Sprintf("failed to create cache directory: %v", err))
	}

	// 验证目录是否可写
	if err := checkDirWritable(c.Dir); err != nil {
		return customerrors.NewParametersError(fmt.Sprintf("cache directory is not writable: %v", err))
	}

	// 验证过期时间
	if c.TTL < time.Hour || c.TTL > 24*time.Hour {
		return customerrors.NewParametersError("cache TTL must be between 1 hour and 24 hours")
	}

	// 验证最大大小
	minSize := int64(50 << 20)  // 50MB
	maxSize := int64(500 << 20) // 500MB
	if c.MaxSize < minSize || c.MaxSize > maxSize {
		return customerrors.NewParametersError(fmt.Sprintf("cache max size must be between %d MB and %d MB",
			minSize>>20, maxSize>>20))
	}

	return nil
}

// checkDirWritable 检查目录是否可写
func checkDirWritable(dir string) error {
	testFile := filepath.Join(dir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	f.Close()
	return os.Remove(testFile)
}
