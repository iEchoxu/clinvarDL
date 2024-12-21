package http

import (
	"context"
	"crypto/tls"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	once       sync.Once
	httpClient *Client
)

// Client HTTP 客户端
type Client struct {
	client *http.Client
	config *HTTPClientConfig
}

// GetHTTPClient 获取全局 HTTP 客户端实例
func GetHTTPClient() *Client {
	once.Do(func() {
		httpClient = NewClient(DefaultHTTPConfig())
	})
	return httpClient
}

// NewClient 创建新的 HTTP 客户端
func NewClient(config *HTTPClientConfig) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // 代理设置
		// TCP 连接设置
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   15 * time.Second, // 增加连接超时
				KeepAlive: config.KeepAlive,
				DualStack: true, // 启用 IPv4/IPv6
			}
			conn, err := dialer.DialContext(ctx, network, addr)
			if tc, ok := conn.(*net.TCPConn); ok {
				tc.SetKeepAlive(true)
				tc.SetKeepAlivePeriod(config.KeepAlive)
				tc.SetNoDelay(true)           // 禁用 Nagle 算法
				tc.SetWriteBuffer(256 * 1024) // 增加 TCP 写缓冲
				tc.SetReadBuffer(256 * 1024)  // 增加 TCP 读缓冲
			}
			return conn, err
		},
		// 连接池配置 - 针对单一 NCBI 服务器优化
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		// TLS 配置
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			ClientSessionCache: tls.NewLRUClientSessionCache(64), // 启用 TLS 会话缓存
		},
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,

		// HTTP 设置
		ForceAttemptHTTP2: true,  // 启用 HTTP/2
		DisableKeepAlives: false, // 启用 keep-alive

		// 超时设置
		ResponseHeaderTimeout: config.ResponseHeaderTimeout, // 响应头超时
		ExpectContinueTimeout: 5 * time.Second,              // 100-continue 超时

		// 启用压缩
		DisableCompression: false,

		// 缓冲区设置
		WriteBufferSize: 128 * 1024, // 128KB
		ReadBufferSize:  128 * 1024, // 128KB
	}

	return &Client{
		client: &http.Client{
			Transport: transport,
			Timeout:   config.Timeout, // 总超时时间
		},
		config: config,
	}
}

// Do 执行 HTTP 请求
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// 先检查上下文是否已取消
	if err := req.Context().Err(); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logcdl.Debug("request cancelled before execution: %v", err)
			return nil, customerrors.NewTimeoutError("request cancelled before execution", err)
		case errors.Is(err, context.DeadlineExceeded):
			logcdl.Debug("request timeout before execution: %v", err)
			return nil, customerrors.NewTimeoutError("request timeout before execution", err)
		default:
			logcdl.Debug("context error before execution: %v", err)
			return nil, customerrors.NewTimeoutError("context error before execution", err)
		}
	}

	resp, err := c.client.Do(req)
	if err == nil {
		return resp, nil
	}

	// 处理错误 - http.Client.Do() 返回的错误都是 url.Error
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return nil, err // 这种情况很难出现
	}

	// 记录原始错误
	logcdl.Debug("original error: %v", err)

	// 记录 url.Error 详情
	logcdl.Debug("url.Error - op: %s, url: %s, err: %v", urlErr.Op, urlErr.URL, urlErr.Err)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		logcdl.Debug("deadline exceeded during execution: %v", err)
		return nil, customerrors.NewTimeoutError("deadline exceeded during execution", err)
	case errors.Is(err, context.Canceled):
		logcdl.Debug("request cancelled during execution: %v", err)
		return nil, customerrors.NewTimeoutError("request cancelled during execution", err)
	case urlErr.Timeout():
		logcdl.Debug("network timeout during execution: %v", err)
		return nil, customerrors.NewTimeoutError("network timeout during execution", err)
	case isHTTP2ProtocolError(urlErr.Err): // 这里需要用 urlErr.Err 因为要判断具体错误信息
		logcdl.Debug("HTTP/2 protocol error: %v", err)
		return nil, customerrors.NewHTTPError(customerrors.WithMessage(urlErr.Err.Error()))
	default:
		// 其他网络错误
		logcdl.Debug("network error during execution: %v", err)
		return nil, customerrors.NewNetError("network error during execution", err)
	}
}

// 辅助函数用于错误类型判断
func isHTTP2ProtocolError(err error) bool {
	errMsg := err.Error()
	logcdl.Debug("isHTTP2ProtocolError errMsg: %v", errMsg)
	return strings.Contains(errMsg, "http2: timeout awaiting response headers") ||
		strings.Contains(errMsg, "server sent GOAWAY") ||
		strings.Contains(errMsg, "Transport received GOAWAY")
}
