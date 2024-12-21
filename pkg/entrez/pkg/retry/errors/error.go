package errors

// 定义不可重试错误类型
// 用 const 而不是 var 是因为 const 是常量，不会被修改
const (
	ErrInput            = Error("input error")
	ErrURL              = Error("url error")
	ErrParse            = Error("parse error")
	ErrSaveResult       = Error("save result error")
	ErrInvalidBatchSize = Error("invalid batch size")
	ErrRateLimit        = Error("rate limit exceeded")
	ErrFailedOpenFile   = Error("open file error")
	ErrInvalidParameter = Error("invalid parameter")
	ErrRetryFailed      = Error("retry failed")
)

// Error 定义错误类型
type Error string

// Error 实现 error 接口
func (e Error) Error() string {
	return string(e)
}

// RetryableError 定义了可重试错误的接口
type RetryableError interface {
	error
	ShouldRetry() bool
}
