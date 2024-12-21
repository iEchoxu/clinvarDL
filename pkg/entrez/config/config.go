package config

import (
	"github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"time"
)

// Config 定义了 Pipeline 的所有配置项
type Config struct {
	// Entrez URL 参数
	EntrezParams *EntrezParams

	// 运行时配置
	Runtime *runtimeConfig

	// 缓存配置
	Cache *CacheConfig

	// 流式处理配置
	Stream *StreamConfig

	// 输出配置
	Output *OutputConfig

	// HTTP 客户端配置
	HTTP *http.HTTPClientConfig
}

// NewConfig 创建一个新的配置实例
func NewConfig(db string) *Config {
	return &Config{
		EntrezParams: &EntrezParams{
			DB: db,
		},
		Runtime: newRuntimeConfig(false),
		Cache:   NewCacheConfig(false, "", 0, 0),
		Stream:  DefaultStreamConfig(),
		Output:  NewOutputConfig(""),
		HTTP:    http.DefaultHTTPConfig(),
	}
}

func (c *Config) SetFilters(filters string) *Config {
	c.EntrezParams.Filters = filters
	return c
}

func (c *Config) SetRetMax(retMax int) *Config {
	c.EntrezParams.RetMax = retMax
	return c
}

func (c *Config) SetUseHistory(useHistory bool) *Config {
	c.EntrezParams.UseHistory = useHistory
	return c
}

func (c *Config) SetRetMode(retMode RetMode) *Config {
	c.EntrezParams.RetMode = retMode
	return c
}

// SetApiKey 设置 API Key 并更新运行时配置
func (c *Config) SetApiKey(apiKey string) *Config {
	c.EntrezParams.ApiKey = apiKey

	// 根据是否有 API Key 重新创建运行时配置
	c.Runtime = newRuntimeConfig(apiKey != "")

	return c
}

func (c *Config) SetEmail(email string) *Config {
	c.EntrezParams.Email = email
	return c
}

func (c *Config) SetToolName(toolName string) *Config {
	c.EntrezParams.ToolName = toolName
	return c
}

// SetCacheEnabled 设置是否启用缓存
func (c *Config) SetCacheEnabled(enabled bool) *Config {
	if c.Cache == nil {
		return c
	}

	c.Cache.Enabled = enabled

	return c
}

// SetCacheDir 设置缓存目录并启用缓存
func (c *Config) SetCacheDir(dir string) *Config {
	if c.Cache == nil {
		return c
	}

	c.Cache.Dir = dir
	return c
}

// SetCacheTTL 设置缓存过期时间
func (c *Config) SetCacheTTL(ttl time.Duration) *Config {
	if c.Cache == nil {
		return c
	}

	c.Cache.TTL = ttl
	return c
}

// SetCacheMaxSize 设置缓存最大大小
func (c *Config) SetCacheMaxSize(size int64) *Config {
	if c.Cache == nil {
		return c
	}

	c.Cache.MaxSize = size
	return c
}

// SetStreamEnabled 设置是否启用流式处理
func (c *Config) SetStreamEnabled(enabled bool) *Config {
	if c.Stream == nil {
		c.Stream = DefaultStreamConfig()
	}
	c.Stream.Enabled = enabled
	return c
}

// SetOutputDir 设置输出目录
func (c *Config) SetOutputDir(dir string) *Config {
	c.Output.Dir = dir
	return c
}

// SetQueryTimeout 设置查询超时时间
func (c *Config) SetQueryTimeout(timeout time.Duration) *Config {
	c.Runtime.QueryTimeout = timeout
	return c
}

// SetSingleQueryTimeout 设置单个查询超时时间
func (c *Config) SetSingleQueryTimeout(timeout time.Duration) *Config {
	c.Runtime.SingleQueryTimeout = timeout
	return c
}

// SetWriteTimeout 设置写入超时时间
func (c *Config) SetWriteTimeout(timeout time.Duration) *Config {
	c.Runtime.WriteTimeout = timeout
	return c
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	// 基本验证
	if c == nil {
		return customerrors.NewParametersError("config is nil")
	}

	// 验证必需参数
	if err := c.EntrezParams.validateRequired(); err != nil {
		return err
	}

	// 验证可选参数
	if err := c.EntrezParams.validateOptional(); err != nil {
		return err
	}

	// 验证运行时配置
	if err := c.Runtime.validate(c.EntrezParams.ApiKey != ""); err != nil {
		return err
	}

	// 验证缓存配置
	if c.Cache == nil {
		return customerrors.NewParametersError("cache config is nil")
	}

	if err := c.Cache.validateCache(); err != nil {
		return err
	}

	// 验证输出配置
	if err := c.Output.validateOutput(); err != nil {
		return err
	}

	return nil
}
