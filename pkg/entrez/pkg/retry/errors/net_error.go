package errors

import (
	"errors"
	"net"
	"net/url"
)

type NetError struct {
	msg   string
	cause error // 保留原始错误
}

func NewNetError(msg string, cause error) *NetError {
	return &NetError{
		msg:   msg,
		cause: cause,
	}
}

func (e *NetError) Error() string {
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

func (e *NetError) Unwrap() error {
	return e.cause
}

// ShouldRetry 判断网络错误是否应该重试
func (e *NetError) ShouldRetry() bool {
	if e.cause == nil {
		return false
	}

	var netErr net.Error
	if !errors.As(e.cause, &netErr) {
		return false
	}

	return shouldRetryNetError(netErr)
}

// ShouldRetryNetError 判断网络错误是否应该重试
func shouldRetryNetError(err net.Error) bool {
	// DNS 错误
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		// DNS 临时错误可以重试
		if dnsErr.Temporary() {
			return true
		}
		// DNS 查找失败可以重试
		if dnsErr.IsTimeout {
			return true
		}
		return false
	}

	// 操作错误
	// "connection refused"、"forcibly closed by the remote host"
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// 连接超时可以重试
		if opErr.Timeout() {
			return true
		}
		// 临时错误可以重试
		if opErr.Temporary() {
			return true
		}

		// 连接被拒绝可以重试
		// if strings.Contains(opErr.Error(), "connection refused") {
		// 	return true
		// }
		// 连接被强制关闭可以重试
		// if strings.Contains(opErr.Error(), "forcibly closed") {
		// 	return true
		// }

		return true
	}

	// 地址错误通常是配置问题，不应重试
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		return false
	}

	// 未知网络错误不重试
	var unknownErr net.UnknownNetworkError
	if errors.As(err, &unknownErr) {
		return false
	}

	// 其他错误只有超时才重试
	return err.Timeout()
}
