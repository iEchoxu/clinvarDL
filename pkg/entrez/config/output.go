package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// OutputConfig 定义输出配置
type OutputConfig struct {
	Dir string // 输出目录
}

// NewOutputConfig 创建输出配置
func NewOutputConfig(dir string) *OutputConfig {
	return &OutputConfig{
		Dir: dir,
	}
}

// validateOutput 验证输出配置
func (o *OutputConfig) validateOutput() error {
	// 验证目录是否为空
	if o.Dir == "" {
		return fmt.Errorf("output directory is required")
	}

	// 验证并创建目录
	if err := os.MkdirAll(o.Dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	return nil
}

// GetOutputPath 获取完整的输出文件路径
func (o *OutputConfig) GetOutputPath(filename string) string {
	return filepath.Join(o.Dir, filename)
}
