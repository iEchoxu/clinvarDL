package retry

import (
	"context"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"math"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

// RetryConfig 定义重试的配置
type RetryConfig struct {
	MaxRetries          int           // 最大重试次数
	BaseDelay           time.Duration // 基础延迟时间
	MaxDelay            time.Duration // 最大延迟时间
	Multiplier          float64       // 退避乘数
	RandomizationFactor float64       // 随机因子
}

// DefaultConfig 返回默认的重试配置
func DefaultConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:          5,
		BaseDelay:           2 * time.Second,
		MaxDelay:            10 * time.Second,
		Multiplier:          1.5,
		RandomizationFactor: 0.2,
	}
}

/*
	DoWithRetry 执行带重试的操作

具体的重试延迟计算：
第1次重试（attempt=2）:
基础延迟 = 1s (1.2^1) = 1.2s
随机范围：0.96s ~ 1.44s（±20%）
第2次重试（attempt=3）:
基础延迟 = 1s (1.2^2) = 1.44s
随机范围：1.152s ~ 1.728s
第3次重试（attempt=4）:
基础延迟 = 1s (1.2^3) = 1.728s
随机范围：1.382s ~ 2.074s
第4次重试（attempt=5）:
基础延迟 = 1s (1.2^4) = 2.074s
随机范围：1.659s ~ 2.489s
第5次重试（attempt=6）:
基础延迟 = 1s (1.2^5) = 2.488s
随机范围：1.990s ~ 2.986s
这种设计具有以下特点：
指数退避：每次重试的延迟时间会逐渐增加
最大限制：延迟时间不会超过10秒
随机抖动：避免多个请求同时重试
最多重试5次：如果仍然失败则返回错误
*/
func DoWithRetry[T any](ctx context.Context, operation string, config *RetryConfig, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
			result, lastErr = fn()
			if lastErr == nil {
				if attempt > 1 {
					logcdl.Success("%s succeeded after %d attempts", operation, attempt)
				}
				return result, nil
			}

			// 如果错误不可重试，则记录错误并跳出重试机制
			if !shouldRetry(lastErr) {
				logcdl.Warn("%s failed with non-retryable error: %v", operation, lastErr)
				return result, lastErr
			}

			// 如果达到最大重试次数，直接返回错误
			if attempt == config.MaxRetries {
				logcdl.Warn("attempt %d/%d for %s failed: %v",
					attempt, config.MaxRetries, operation, lastErr)
				return result, lastErr
			}

			// 计算延迟时间,并记录重试信息
			// 使用 %v 打印错误信息而不用 errors.Unwrap(lastErr) 打印,因为 errors.Unwrap(lastErr) 可能为 nil
			// 另外自定义错误类型实现了 Error() 方法, 直接使用 %v 打印即可
			delay := calculateDelay(attempt, config)
			logcdl.Warn("attempt %d/%d for %s failed: %v, retrying in %v",
				attempt, config.MaxRetries, operation, lastErr, delay)

			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return result, ctx.Err()
			}
		}
	}

	return result, lastErr
}

func calculateDelay(attempt int, config *RetryConfig) time.Duration {
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt-1))
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	if config.RandomizationFactor > 0 {
		delta := config.RandomizationFactor * delay
		minDelay := delay - delta
		maxDelay := delay + delta
		delay = minDelay + rand.Float64()*(maxDelay-minDelay)
	}

	return time.Duration(delay)
}

// shouldRetry 判断是否应该重试
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是已知的可重试错误类型
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return true
	}

	// 使用 errors.As 检查错误链中的可重试错误
	var retryErr customerrors.RetryableError
	if errors.As(err, &retryErr) {
		return retryErr.ShouldRetry()
	}

	return false
}
