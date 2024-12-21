package config

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"time"
)

const (
	// 默认缓冲区大小, 用于结果通道和错误通道
	DefaultBufferSize    = 1000
	DefaultMaxBufferSize = 10000

	// 默认批处理大小
	DefaultBatchSize    = 1000
	DefaultMinBatchSize = 500
	DefaultMaxBatchSize = 2000

	// 默认超时时间
	DefaultMinTimeout = 10 * time.Second
	DefaultMaxTimeout = 120 * time.Minute

	// 默认最大响应大小
	DefaultMinResponseSize = 10 << 20  // 10MB
	DefaultMaxResponseSize = 100 << 20 // 100MB

	// 有 API Key 时的并发数 (10次/秒)
	// 并发控制配置 (总并发数不超过10)
	// 总请求数计算：
	// - esearch 请求数 = 3
	// - esummary 请求数 = 2 * 3 = 6 (MaxESummaryWorkers * queryWorker数量)
	// - 总请求数 = 3 + 6 = 9次/秒
	// 符合 NCBI 有 API Key 时的 10次/秒限制
	DefaultMaxQueryWorkersWithKey    = 3
	DefaultMaxEsummaryWorkersWithKey = 2

	// 无 API Key 时的并发数  (3次/秒)
	// 并发控制配置 (总并发数不超过3)
	// 总请求数计算：
	// - esearch 请求数 = 1
	// - esummary 请求数 = 1 * 1 = 1 (MaxESummaryWorkers * queryWorker数量)
	// - 总请求数 = 1 + 1 = 2次/秒
	// 符合 NCBI 无 API Key 时的 3次/秒限制
	DefaultMaxQueryWorkersWithoutKey    = 1
	DefaultMaxEsummaryWorkersWithoutKey = 1
)

// runtimeConfig 定义了运行时配置（私有）
type runtimeConfig struct {
	// 通道缓冲区配置
	BufferSize    int // 结果通道或错误通道缓冲区大小
	MaxBufferSize int // 最大缓冲区大小

	// 并发控制配置
	// 请求流程说明：
	// 1. 每个 pipeline 会启动 1 个协程执行 esearch 请求
	// 2. 每个 esearch 请求会启动 MaxEsummaryWorkers 个协程并发获取一批基因详细信息
	// 3. 每秒最多有 MaxQueryWorkers 个 esearch 请求
	//
	// 每秒总请求数计算：
	// - esearch 请求数 = MaxQueryWorkers
	// - esummary 请求数 = MaxESummaryWorkers * MaxQueryWorkers
	// - 总请求数 = MaxQueryWorkers + (MaxESummaryWorkers * MaxQueryWorkers)
	//
	// 速率限制：
	// - 无 API Key: 总请求数必须 <= 3次/秒
	// - 有 API Key: 总请求数必须 <= 10次/秒
	// Tip(有更高并发要求的情况下):
	// - 可通过 maxQueryWorkersWithFilter 和 maxQueryWorkersWithOutFilter 对 esearch 并发数进行细分控制
	// - 可通过 maxEsummaryWorkersWithFilter 和 maxEsummaryWorkersWithOutFilter 对 esummary 并发数进行细分控制
	MaxQueryWorkers    int // 最大查询并发数
	MaxEsummaryWorkers int // 最大 ESummary 并发数

	// 批处理配置
	BatchSize    int // 默认批量大小
	MinBatchSize int // 最小批量大小
	MaxBatchSize int // 最大批量大小

	// 响应大小配置
	MaxResponseSize int64 // 最大响应大小

	// 超时配置
	QueryTimeout       time.Duration // 总查询超时时间
	SingleQueryTimeout time.Duration // 单个查询超时时间
	WriteTimeout       time.Duration // 写入超时时间
}

// newRuntimeConfig 创建新的运行时配置
func newRuntimeConfig(hasAPIKey bool) *runtimeConfig {
	maxQueryWorkers := DefaultMaxQueryWorkersWithoutKey
	maxEsummaryWorkers := DefaultMaxEsummaryWorkersWithoutKey

	if hasAPIKey {
		maxQueryWorkers = DefaultMaxQueryWorkersWithKey
		maxEsummaryWorkers = DefaultMaxEsummaryWorkersWithKey
	}

	return &runtimeConfig{
		BufferSize:         DefaultBufferSize,
		MaxBufferSize:      DefaultMaxBufferSize,
		MaxQueryWorkers:    maxQueryWorkers,
		MaxEsummaryWorkers: maxEsummaryWorkers,
		BatchSize:          DefaultBatchSize,
		MinBatchSize:       DefaultMinBatchSize,
		MaxBatchSize:       DefaultMaxBatchSize,
		MaxResponseSize:    DefaultMaxResponseSize,
		QueryTimeout:       30 * time.Minute,
		SingleQueryTimeout: 20 * time.Minute,
		WriteTimeout:       10 * time.Minute,
	}
}

// validate 验证所有配置
func (r *runtimeConfig) validate(hasAPIKey bool) error {
	// 配置不能为空
	if r == nil {
		return customerrors.NewParametersError("runtime config is nil")
	}

	// 依次验证每个配置
	for _, fn := range []func() error{
		r.validateBufferSize,
		r.validateBatchSize,
		r.validateTimeout,
		r.validateResponseSize,
		func() error { return r.validateWorkers(hasAPIKey) },
	} {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// validateBufferSize 验证缓冲区大小
func (r *runtimeConfig) validateBufferSize() error {
	if r.BufferSize <= 0 || r.BufferSize > r.MaxBufferSize {
		return customerrors.NewParametersError(fmt.Sprintf("buffer size must be between 1 and %d", r.MaxBufferSize))
	}
	return nil
}

// validateBatchSize 验证批处理大小
func (r *runtimeConfig) validateBatchSize() error {
	if r.BatchSize < r.MinBatchSize || r.BatchSize > r.MaxBatchSize {
		return customerrors.NewParametersError(fmt.Sprintf("batch size must be between %d and %d",
			r.MinBatchSize, r.MaxBatchSize))
	}
	return nil
}

// validateTimeout 验证超时时间
func (r *runtimeConfig) validateTimeout() error {
	if r.QueryTimeout < DefaultMinTimeout || r.QueryTimeout > DefaultMaxTimeout {
		return customerrors.NewParametersError(fmt.Sprintf("query timeout must be between %v and %v",
			DefaultMinTimeout, DefaultMaxTimeout))
	}
	if r.SingleQueryTimeout < DefaultMinTimeout || r.SingleQueryTimeout > r.QueryTimeout {
		return customerrors.NewParametersError(fmt.Sprintf("single query timeout must be between %v and %v",
			DefaultMinTimeout, r.QueryTimeout))
	}
	if r.WriteTimeout < DefaultMinTimeout || r.WriteTimeout > r.QueryTimeout {
		return customerrors.NewParametersError(fmt.Sprintf("write timeout must be between %v and %v",
			DefaultMinTimeout, r.QueryTimeout))
	}
	return nil
}

// validateResponseSize 验证响应大小
func (r *runtimeConfig) validateResponseSize() error {
	if r.MaxResponseSize < DefaultMinResponseSize || r.MaxResponseSize > DefaultMaxResponseSize {
		return customerrors.NewParametersError(fmt.Sprintf("max response size must be between %d MB and %d MB",
			DefaultMinResponseSize>>20, DefaultMaxResponseSize>>20))
	}
	return nil
}

// validateWorkers 验证并发数
func (r *runtimeConfig) validateWorkers(hasAPIKey bool) error {
	if !hasAPIKey {
		// 无 API Key 时，单个参数的限制
		if r.MaxQueryWorkers <= 0 || r.MaxQueryWorkers > 1 {
			return customerrors.NewParametersError("max query workers must be 1 when no API key")
		}
		if r.MaxEsummaryWorkers <= 0 || r.MaxEsummaryWorkers > 1 {
			return customerrors.NewParametersError("max esummary workers must be 1 when no API key")
		}
	}

	// 计算总请求数: maxQueryWorkers + (maxEsummaryWorkers * maxQueryWorkers)
	totalRequests := r.MaxQueryWorkers * (1 + r.MaxEsummaryWorkers)

	// 获取速率限制
	rateLimit := http.DefaultRateLimit
	if hasAPIKey {
		rateLimit = http.ApiKeyRateLimit
	}

	// 验证总请求数是否超过速率限制
	if totalRequests > rateLimit {
		return customerrors.NewParametersError(fmt.Sprintf("total requests (%d) exceeds rate limit (%d)",
			totalRequests, rateLimit))
	}

	return nil
}

// GetQueryWorkers 获取查询并发数
func (r *runtimeConfig) GetQueryWorkers() int {
	return r.MaxQueryWorkers
}

// GetEsummaryWorkers 获取 ESummary 并发数
func (r *runtimeConfig) GetEsummaryWorkers() int {
	return r.MaxEsummaryWorkers
}
