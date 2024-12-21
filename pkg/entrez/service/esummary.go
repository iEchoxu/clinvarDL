package service

import (
	"bytes"
	"context"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/service/response"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
	"net/url"

	"github.com/pkg/errors"
)

// ESummaryOperation 实现了 esummary 操作
type ESummaryOperation struct {
	BaseOperation
}

// NewESummaryOperation 创建一个新的 ESummaryOperation 实例
func NewESummaryOperation(httpClient *http.Client, rateLimiter *http.RateLimiter) *ESummaryOperation {
	return &ESummaryOperation{
		BaseOperation: *NewBaseOperation(BaseURLESummary, httpClient, rateLimiter),
	}
}

// Execute 执行 ESummary 操作
func (s *ESummaryOperation) Execute(ctx context.Context, webEnv, queryKey string, query *types.Query) (*types.ESummaryResult, error) {
	if s.BaseOperation.useStream {
		logcdl.Debug("using stream for esummary")
		return s.ExecuteStream(ctx, webEnv, queryKey, query)
	}
	logcdl.Debug("using non-stream for esummary")
	return s.ExecuteWithoutStream(ctx, webEnv, queryKey, query)
}

// ExecuteStream 执行 ESummary 流式操作
func (s *ESummaryOperation) ExecuteStream(ctx context.Context, webEnv, queryKey string, query *types.Query) (*types.ESummaryResult, error) {
	var buffer bytes.Buffer

	// 设置 ESummary 请求参数
	s.Parameters.Set("WebEnv", webEnv)
	s.Parameters.Set("query_key", queryKey)

	esummaryURL, err := s.BuildURL()
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to build esummary url for query: '%v'", query)
	}

	// 打印 URL 信息
	urlString, _ := url.QueryUnescape(esummaryURL.String())
	logcdl.Debug("esummary stream url for query '%v': '%s'", query, urlString)

	// 使用流式处理并写入缓冲区
	err = s.doStreamRequest(ctx, "GET", esummaryURL.String(), s.Parameters, func(chunk []byte) error {
		_, err := buffer.Write(chunk)
		return err
	})
	if err != nil {
		return nil, err
	}

	// 创建解析器并解析响应
	parser, err := response.NewESummaryResponseParser(response.ParserType(s.GetRetMode()))
	if err != nil {
		return nil, err
	}
	return parser.ParseESummary(buffer.Bytes())
}

// ExecuteWithoutStream 执行 ESummary 非流式操作
func (s *ESummaryOperation) ExecuteWithoutStream(ctx context.Context, webEnv, queryKey string, query *types.Query) (*types.ESummaryResult, error) {
	// 设置 ESummary 请求参数
	s.Parameters.Set("WebEnv", webEnv)
	s.Parameters.Set("query_key", queryKey)

	esummaryURL, err := s.BuildURL()
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to build esummary url for query: '%v'", query)
	}

	// 打印 URL 信息
	urlString, _ := url.QueryUnescape(esummaryURL.String())
	logcdl.Debug("esummary url for query '%v': '%s'", query, urlString)

	body, err := s.doRequest(ctx, "GET", esummaryURL.String(), s.Parameters)
	if err != nil {
		return nil, err
	}

	// Debug: 查看哪个批次的查询未获得有效结果
	if len(body) < 1024 {
		logcdl.Debug("response body length: %d bytes for esummary query '%v'", len(body), query)
		logcdl.Debug("response body: %s", string(body))
	}

	// 创建解析器
	parser, err := response.NewESummaryResponseParser(response.ParserType(s.GetRetMode()))
	if err != nil {
		return nil, err
	}

	// 解析响应
	result, err := parser.ParseESummary(body)
	if err != nil {
		return nil, err
	}

	return result, nil
}
