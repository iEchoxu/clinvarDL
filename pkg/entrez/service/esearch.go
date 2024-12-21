package service

import (
	"context"
	customeHttp "github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
	"net/url"
	"strings"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/service/response"

	"github.com/pkg/errors"
)

// ESearchConfig 定义 ESearch 的配置
type ESearchConfig struct {
	Filters string
}

// ESearchOperation 实现了 esearch 操作
type ESearchOperation struct {
	BaseOperation
	config ESearchConfig
}

// NewESearchOperation 创建一个新的 ESearchOperation 实例
func NewESearchOperation(httpClient *customeHttp.Client, rateLimiter *customeHttp.RateLimiter) *ESearchOperation {
	return &ESearchOperation{
		BaseOperation: *NewBaseOperation(BaseURLESearch, httpClient, rateLimiter),
	}
}

// SetQueryFilters 设置查询过滤器
func (eo *ESearchOperation) SetQueryFilters(filters string) *ESearchOperation {
	eo.config.Filters = filters

	return eo
}

// Execute 执行 ESearch 操作
func (eo *ESearchOperation) Execute(ctx context.Context, query *types.Query) (*types.ESearchResult, error) {
	// 设置搜索词: 拼接 filters 和 query
	// 重要逻辑：参考 clinvar advanced search 的搜索词格式
	if eo.config.Filters == "" {
		eo.Parameters.Set("term", "("+query.Content+")")
	} else {
		batchString := &strings.Builder{}
		batchString.WriteString("(")
		batchString.WriteString("(" + query.Content + ")")
		batchString.WriteString(" AND ")
		batchString.WriteString(eo.config.Filters)
		batchString.WriteString(")")
		eo.Parameters.Set("term", batchString.String())
	}

	esearchURL, err := eo.BuildURL()
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to build esearch url for query '%v'", query)
	}

	// 打印 ESearch URL
	esearchURLString, _ := url.QueryUnescape(esearchURL.String())
	logcdl.Debug("esearch url for query '%v': '%s'", query, esearchURLString)

	method := "GET"
	if len(esearchURL.String()) > 2048 {
		method = "POST"
	}

	body, err := eo.doRequest(ctx, method, esearchURL.String(), eo.Parameters)
	if err != nil {
		return nil, err
	}

	// 创建解析器
	parser, err := response.NewESearchResponseParser(response.ParserType(eo.GetRetMode()))
	if err != nil {
		return nil, err
	}

	// 解析响应
	result, err := parser.ParseESearch(body)
	if err != nil {
		return nil, err
	}

	return result, nil
}
