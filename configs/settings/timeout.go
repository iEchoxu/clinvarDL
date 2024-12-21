package settings

import (
	"time"
)

// TimeoutSettings 定义超时相关配置
type TimeoutSettings struct {
	QueryTimeout       time.Duration `yaml:"query_timeout"`        // 总查询超时时间
	SingleQueryTimeout time.Duration `yaml:"single_query_timeout"` // 单个查询超时时间
	WriteTimeout       time.Duration `yaml:"write_timeout"`        // 写入超时时间
}

func NewTimeoutSettings() *TimeoutSettings {
	return &TimeoutSettings{
		QueryTimeout:       30 * time.Minute,
		SingleQueryTimeout: 20 * time.Minute,
		WriteTimeout:       10 * time.Minute,
	}
}
