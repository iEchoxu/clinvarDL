package http

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(hasAPIKey bool) *RateLimiter {
	rateLimit := DefaultRateLimit
	if hasAPIKey {
		rateLimit = ApiKeyRateLimit
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rateLimit), BurstSize),
	}
}

// WaitN 等待 n 个令牌可用
func (r *RateLimiter) WaitN(ctx context.Context, n int) error {
	return r.limiter.WaitN(ctx, n)
}

// GetCurrentRate 获取当前速率限制
func (r *RateLimiter) GetCurrentRate() float64 {
	return float64(r.limiter.Limit())
}
