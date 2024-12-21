package http

import (
	"time"
)

// HTTPConfig HTTP 客户端配置
// 活动连接数 = MaxConnsPerHost - MaxIdleConnsPerHost
// 总空闲连接数 = MaxIdleConns
// 每个 host 空闲连接数 = MaxIdleConnsPerHost
// 当前设置：
// - 最大活动连接数 = 40 - 15 = 25 个
// - 最大空闲连接数 = 15 个
// - 总连接数上限 = 40 个
type HTTPClientConfig struct {
	Timeout               time.Duration // 总超时时间
	MaxIdleConns          int           // 所有 host 的最大空闲连接数总和
	MaxIdleConnsPerHost   int           // 每个 host 的最大空闲连接数
	MaxConnsPerHost       int           // 每个 host 的最大连接数（包括活动和空闲）
	IdleConnTimeout       time.Duration // 空闲连接超时时间
	TLSHandshakeTimeout   time.Duration // TLS 握手超时时间
	ResponseHeaderTimeout time.Duration // 响应头超时时间
	KeepAlive             time.Duration // 保持连接时间
}

// DefaultHTTPConfig 返回默认配置
func DefaultHTTPConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout: 90 * time.Second,
		// 由于只连接单个 host，MaxIdleConns 应该等于 MaxIdleConnsPerHost
		MaxIdleConns:          15,
		MaxIdleConnsPerHost:   15,
		MaxConnsPerHost:       40,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 40 * time.Second,
		KeepAlive:             60 * time.Second,
	}
}
