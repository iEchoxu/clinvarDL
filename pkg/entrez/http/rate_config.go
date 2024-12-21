package http

const (
	// API 速率限制
	ApiKeyRateLimit  = 10 // 有 API Key 时每秒允许的请求数
	DefaultRateLimit = 3  // 无 API Key 时每秒允许的请求数
	BurstSize        = 1  // 突发请求数量限制
)
