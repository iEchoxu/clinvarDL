package errors

import (
	"net/http"
	"strconv"
	"strings"
)

// HTTPError 定义了 HTTP 错误
type HTTPError struct {
	msg        string
	statusCode int
}

// HTTPErrorOption 定义错误选项函数类型
type HTTPErrorOption func(*HTTPError)

// WithMessage 设置错误消息
func WithMessage(msg string) HTTPErrorOption {
	return func(e *HTTPError) {
		e.msg = msg
		// 根据错误消息设置状态码
		switch {
		case strings.Contains(msg, "timeout awaiting response headers"):
			e.statusCode = http.StatusRequestTimeout // 408
		case strings.Contains(msg, "server sent GOAWAY"):
			e.statusCode = http.StatusServiceUnavailable // 503
		default:
			e.statusCode = http.StatusInternalServerError // 500
		}
	}
}

// WithStatusCode 设置状态码
// link: https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
func WithStatusCode(code int) HTTPErrorOption {
	return func(e *HTTPError) {
		e.statusCode = code
		// 根据状态码设置消息
		switch {
		case code == http.StatusInternalServerError: // 500
			e.msg = "internal server error"
		case code == http.StatusNotImplemented: // 501
			e.msg = "not implemented"
		case code == http.StatusBadGateway: // 502
			e.msg = "bad gateway"
		case code == http.StatusServiceUnavailable: // 503
			e.msg = "service unavailable"
		case code == http.StatusGatewayTimeout: // 504
			e.msg = "gateway timeout"
		case code == http.StatusHTTPVersionNotSupported: // 505
			e.msg = "HTTP version not supported"
		case code == http.StatusTooManyRequests: // 429
			e.msg = "too many requests"
		case code == http.StatusRequestTimeout: // 408
			e.msg = "request timeout"
		case code >= 500:
			e.msg = "server error"
		case code >= 400:
			e.msg = "client error"
		default:
			e.msg = "HTTP request failed"
		}
	}
}

// NewHTTPError 创建 HTTP 错误
func NewHTTPError(opts ...HTTPErrorOption) *HTTPError {
	e := &HTTPError{
		statusCode: http.StatusInternalServerError, // 默认 500
		msg:        "HTTP request failed",          // 默认消息
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Error 实现 error 接口

func (e *HTTPError) Error() string {
	// 返回包含状态码的错误消息
	// 对于 HTTP/2 协议错误
	// "http2: timeout awaiting response headers (status code: 408)"
	// "server sent GOAWAY and closed the connection (status code: 503)"

	// 对于 HTTP 状态码错误
	// "service unavailable (status code: 503)"
	// "too many requests (status code: 429)"
	return e.msg + " (status code: " + strconv.Itoa(e.statusCode) + ")"
}

// ShouldRetry 判断 HTTP 错误是否应该重试
// "http2: server sent GOAWAY and closed the connection"、"http2: Transport received GOAWAY"、"http2: timeout awaiting response headers"、
// http2: timeout awaiting response headers 属于 urlErr.Timeout()
func (e *HTTPError) ShouldRetry() bool {
	switch {
	case e.statusCode >= 500:
		return true
	case e.statusCode == http.StatusTooManyRequests:
		return true
	case e.statusCode == http.StatusRequestTimeout,
		e.statusCode == http.StatusBadGateway,
		e.statusCode == http.StatusServiceUnavailable,
		e.statusCode == http.StatusGatewayTimeout:
		return true
	default:
		// 其它错误不重试，如果: 400、401、403、404 等
		return false
	}
}
