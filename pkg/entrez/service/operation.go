package service

import (
	"bufio"
	"context"
	"fmt"
	customeHttp "github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const (
	baseURL         = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/"
	BaseURLESearch  = baseURL + "esearch.fcgi"
	BaseURLEPost    = baseURL + "epost.fcgi"
	BaseURLESummary = baseURL + "esummary.fcgi"
)

// BaseOperation 包含所有 Entrez 操作共享的属性和方法
type BaseOperation struct {
	BaseURL     string
	Parameters  url.Values
	RateLimiter *customeHttp.RateLimiter
	useStream   bool
	httpClient  *customeHttp.Client
}

// NewBaseOperation 创建基础操作实例
func NewBaseOperation(baseURL string, httpClient *customeHttp.Client, rateLimiter *customeHttp.RateLimiter) *BaseOperation {
	return &BaseOperation{
		BaseURL:     baseURL,
		Parameters:  url.Values{},
		RateLimiter: rateLimiter,
		httpClient:  httpClient,
	}
}

// SetDB 设置数据库
func (b *BaseOperation) SetDB(db string) *BaseOperation {
	b.Parameters.Set("db", db)
	return b
}

// SetRetMax 设置 xml 返回的最大结果数(ID 数量,默认为 20)
// 如果需要获取所有结果，则需要将 RetMax 设置为 100000 或者指定 retmax=10000 并结合 retstart 分批获取，则可能会导致数据丢失
// 如果设置了 usehistory，则将查询到的所有 id 都上传到历史服务器，再从历史服务器获取数据
// pubmed 数据库单次最大返回 10000 个 ID，超过 10000 个 ID 需要使用 retstart 和 retmax 分批获取
func (b *BaseOperation) SetRetMax(retMax int) *BaseOperation {
	if retMax > 0 {
		b.Parameters.Set("retmax", fmt.Sprintf("%d", retMax))
	} else {
		b.Parameters.Del("retmax")
	}
	return b
}

// SetRetMode 设置返回数据的格式，xmljson
func (b *BaseOperation) SetRetMode(retMode string) *BaseOperation {
	if retMode != "" {
		b.Parameters.Set("retmode", retMode)
	} else {
		b.Parameters.Del("retmode")
	}
	return b
}

// SetUseHistory 设置是否使用历史
func (b *BaseOperation) SetUseHistory(useHistory bool) *BaseOperation {
	if useHistory {
		b.Parameters.Set("usehistory", "y")
	} else {
		b.Parameters.Del("usehistory")
	}
	return b
}

// SetEmail 设置邮箱
func (b *BaseOperation) SetEmail(email string) *BaseOperation {
	if email != "" {
		b.Parameters.Set("email", email)
	} else {
		b.Parameters.Del("email")
	}
	return b
}

// SetApiKey 设置 API 密钥
func (b *BaseOperation) SetApiKey(apiKey string) *BaseOperation {
	if apiKey != "" {
		b.Parameters.Set("api_key", apiKey)
	} else {
		b.Parameters.Del("api_key")
	}
	return b
}

// SetToolName 设置工具名称
func (b *BaseOperation) SetToolName(tool string) *BaseOperation {
	if tool != "" {
		b.Parameters.Set("tool", tool)
	} else {
		b.Parameters.Del("tool")
	}
	return b
}

// SetUseStream 设置是否使用流式处理
func (b *BaseOperation) SetUseStream(useStream bool) *BaseOperation {
	b.useStream = useStream
	return b
}

func (b *BaseOperation) GetRetMode() string {
	return b.Parameters.Get("retmode")
}

// BuildURL 构建完整的 URL
func (b *BaseOperation) BuildURL() (*url.URL, error) {
	u, err := url.Parse(b.BaseURL)
	if err != nil {
		return nil, errors.Wrapf(customerrors.ErrURL, "failed to parse base url: %v", err)
	}

	q := u.Query()
	for k, v := range b.Parameters {
		q[k] = v
	}
	u.RawQuery = q.Encode()

	return u, nil
}

// doRequest 执行 HTTP 请求
func (b *BaseOperation) doRequest(ctx context.Context, method, url string, params url.Values) ([]byte, error) {
	// 检查 context 是否已取消
	if ctx.Err() != nil {
		return nil, customerrors.NewTimeoutError(
			"context cancelled before request",
			ctx.Err(),
		)
	}

	// 在发送请求前等待速率限制
	if err := b.RateLimiter.WaitN(ctx, 1); err != nil {
		return nil, customerrors.NewTimeoutError(
			"rate limit wait failed",
			err,
		)
	}

	// 创建请求
	req, err := b.createRequest(ctx, method, url, params)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "github.com/iEchoxu/clinvarDL/1.0")

	// 获取 HTTP 客户端并执行请求
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 处理响应
	return b.handleResponse(resp)
}

// createRequest 创建 HTTP 请求
func (b *BaseOperation) createRequest(ctx context.Context, method, url string, params url.Values) (*http.Request, error) {
	var req *http.Request
	var err error

	switch method {
	case "GET":
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	case "POST":
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(params.Encode()))
		if err == nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	default:
		return nil, errors.Wrapf(customerrors.ErrInput, "unsupported http method: %s", method)
	}

	if err != nil {
		return nil, customerrors.NewNetError("failed to create request", err)
	}

	return req, nil
}

// handleResponse 处理 HTTP 响应
func (b *BaseOperation) handleResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, customerrors.NewHTTPError(customerrors.WithStatusCode(resp.StatusCode))
	}

	// TODO: 数据量过多时采用流式读取
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, customerrors.NewNetError("failed to read response body", err) // 读取响应体失败直接返回网络错误类型
	}

	if len(body) == 0 {
		return nil, customerrors.NewEmptyResultError("server returned empty response")
	}

	return body, nil
}

// streamResponse 流式处理 HTTP 响应
func (b *BaseOperation) streamResponse(resp *http.Response, handler func(chunk []byte) error) error {
	// 使用 bufio.Reader 进行缓冲读取
	reader := bufio.NewReaderSize(resp.Body, 128*1024) // 128KB 缓冲区

	buffer := make([]byte, 128*1024) // 128KB 缓冲区

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			// 处理当前读取的数据块
			chunk := buffer[:n]
			if err := handler(chunk); err != nil {
				return errors.Wrap(err, "failed to handle response chunk")
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return customerrors.NewNetError("failed to read response body", err)
		}
	}

	return nil
}

// doStreamRequest 执行流式 HTTP 请求
func (b *BaseOperation) doStreamRequest(ctx context.Context, method, url string, params url.Values, handler func(chunk []byte) error) error {
	// 检查 context 是否已取消
	if ctx.Err() != nil {
		return customerrors.NewTimeoutError(
			"context cancelled before request",
			ctx.Err(),
		)
	}

	// 在发送请求前等待速率限制
	if err := b.RateLimiter.WaitN(ctx, 1); err != nil {
		return customerrors.NewTimeoutError(
			"rate limit wait failed",
			err,
		)
	}

	// 创建请求
	req, err := b.createRequest(ctx, method, url, params)
	if err != nil {
		return err
	}

	// 获取 HTTP 客户端并执行请求
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return customerrors.NewNetError("failed to execute request", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return customerrors.NewHTTPError(customerrors.WithStatusCode(resp.StatusCode))
	}

	// 流式处理响应
	return b.streamResponse(resp, handler)
}
