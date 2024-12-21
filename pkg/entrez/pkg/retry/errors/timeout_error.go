package errors

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
)

// TimeoutError 定义了超时错误
type TimeoutError struct {
	msg   string
	cause error // 保留原始错误以便追踪超时原因
}

// NewTimeoutError 创建新的超时错误
func NewTimeoutError(msg string, cause error) *TimeoutError {
	return &TimeoutError{
		msg:   msg,
		cause: cause,
	}
}

// Error 实现 error 接口
func (e *TimeoutError) Error() string {
	if e.cause == nil {
		return e.msg
	}

	// 如果是 url.Error，尝试获取底层错误
	var urlErr *url.Error
	if errors.As(e.cause, &urlErr) {
		// 返回底层错误消息
		return e.msg + ": " + urlErr.Err.Error()
	}

	// 其他错误直接返回
	return e.msg + ": " + e.cause.Error()
}

// Unwrap 实现 errors.Unwrap 接口
func (e *TimeoutError) Unwrap() error {
	return e.cause
}

func (e *TimeoutError) ShouldRetry() bool {
	if errors.Is(e.cause, context.DeadlineExceeded) {
		return true
	}
	return true
}

func (e *TimeoutError) IsRetryable() bool {
	return e.ShouldRetry()
}
