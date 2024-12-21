package pipeline

import (
	"context"
	"fmt"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/config"
	customeHttp "github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/service"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

type EsearchExecutor struct {
	Esearch *service.ESearchOperation
	Config  *config.Config
}

func NewEsearchExecutor(config *config.Config, httpClient *customeHttp.Client, rateLimiter *customeHttp.RateLimiter) *EsearchExecutor {
	esearchExecutor := &EsearchExecutor{
		Esearch: service.NewESearchOperation(httpClient, rateLimiter),
		Config:  config,
	}

	esearchExecutor.setupOperations()

	return esearchExecutor
}

func (e *EsearchExecutor) setupOperations() {
	// 配置 eSearch 查询参数
	e.Esearch.SetQueryFilters(e.Config.EntrezParams.Filters).
		SetDB(e.Config.EntrezParams.DB).
		SetRetMax(e.Config.EntrezParams.RetMax).
		SetRetMode(e.Config.EntrezParams.RetMode.String()).
		SetUseHistory(e.Config.EntrezParams.UseHistory).
		SetEmail(e.Config.EntrezParams.Email).
		SetApiKey(e.Config.EntrezParams.ApiKey).
		SetToolName(e.Config.EntrezParams.ToolName)
}

// executeSearch 执行 ESearch 操作并有重试机制
func (e *EsearchExecutor) executeSearch(ctx context.Context, query *types.Query, searchChan chan<- *types.ESearchResult, collector *types.QueryResult) {
	config := retry.DefaultConfig()
	result, err := retry.DoWithRetry(ctx, fmt.Sprintf("esearch query for '%v'", query), config, func() (*types.ESearchResult, error) {
		searchResult, err := e.Esearch.Execute(ctx, query)
		if err != nil {
			return nil, err // 这里直接返回错误，错误由 Retry 机制处理
		}

		if searchResult == nil || searchResult.Count == 0 {
			return nil, customerrors.NewEmptyResultError(fmt.Sprintf("no records found for query '%v'", query))
		}

		return searchResult, nil
	})

	// 如果所有重试都失败或结果为空，则直接返回
	if err != nil || result == nil {
		logcdl.Error("execute esearch failed after all retries for query '%v'", query)

		// 这里不添加进失败批次, 因为 esearch 失败意味着整个查询失败，不是批次级别的失败
		// 且不会生成对应的缓存，如果没有缓存则在下次执行相同查询时会发起新请求
		collector.SetStatusOnError(err, e.Config.EntrezParams.Filters != "")

		searchChan <- nil // 发送 nil，防止 esummary 协程卡住

		return
	}

	collector.SetTotalRecords(result.Count) // 记录总记录数

	select {
	case searchChan <- result:
		logcdl.Info("esearch result for query '%v': id count: %d, WebEnv: '%s', QueryKey: '%s'",
			query, result.Count, result.WebEnv, result.QueryKey)
	case <-ctx.Done():
		err := customerrors.NewTimeoutError(
			fmt.Sprintf("esearch request timed out after %v for '%v'",
				e.Config.Runtime.QueryTimeout, query),
			ctx.Err())
		collector.SetStatusOnError(err, e.Config.EntrezParams.Filters != "")
	}
}

func (e *EsearchExecutor) RetryBatch(ctx context.Context, query *types.Query) (*types.ESearchResult, error) {
	return e.Esearch.Execute(ctx, query)
}
