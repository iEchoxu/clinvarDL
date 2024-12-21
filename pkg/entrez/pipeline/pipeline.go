package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/config"
	customHttp "github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

type Pipeline struct {
	Search  *EsearchExecutor
	Summary *EsummaryExecutor
	Config  *config.Config
}

func NewPipeline(config *config.Config, httpClient *customHttp.Client, rateLimiter *customHttp.RateLimiter) *Pipeline {
	return &Pipeline{
		Search:  NewEsearchExecutor(config, httpClient, rateLimiter),
		Summary: NewEsummaryExecutor(config, httpClient, rateLimiter),
		Config:  config,
	}
}

// ExecuteQuery 按固定顺序执行操作管道
func (p *Pipeline) ExecuteQuery(ctx context.Context, query *types.Query) (*types.QueryResult, error) {
	logcdl.Info("starting query '%v': '%s'", query, query.Content)

	//  初始化查询统计信息
	queryStats := types.NewQueryResult(query.GetQueryID(), query.Content)

	// 执行新的查询
	return p.executeQuery(ctx, query, queryStats)
}

// executeQuery 执行查询
func (p *Pipeline) executeQuery(ctx context.Context, query *types.Query, queryStats *types.QueryResult) (*types.QueryResult, error) {
	// 使用单个查询的超时时间
	queryCtx, cancel := context.WithTimeout(ctx, p.Config.Runtime.SingleQueryTimeout)
	defer cancel()

	// 创建通道
	searchChan := make(chan *types.ESearchResult, 1)
	doneChan := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	// 启动 eSearch 查询
	go func() {
		defer wg.Done()
		p.Search.executeSearch(queryCtx, query, searchChan, queryStats)
	}()

	// 启动 eSummary 查询
	go func() {
		defer wg.Done()
		p.Summary.ProcessSummaryFlow(queryCtx, query, searchChan, queryStats)
	}()

	// 等待所有操作完成
	go func() {
		wg.Wait()
		close(doneChan) // 关闭通道，触发 select
	}()

	// 等待结果或超时
	return p.waitForResults(queryCtx, query, queryStats, doneChan)
}

// waitForResults 等待结果或超时
func (p *Pipeline) waitForResults(ctx context.Context, query *types.Query, queryStats *types.QueryResult, doneChan <-chan struct{}) (*types.QueryResult, error) {
	select {
	case <-ctx.Done():
		err := customerrors.NewTimeoutError(
			fmt.Sprintf("pipeline request timed out after %v for '%v'",
				p.Config.Runtime.QueryTimeout, query),
			ctx.Err())
		queryStats.SetStatusOnError(err, p.Config.EntrezParams.Filters != "")
		return queryStats, err
	case <-doneChan: // 等待所有操作完成
		return p.processQueryResults(queryStats)
	}
}

// processQueryResults 处理查询结果
func (p *Pipeline) processQueryResults(queryStats *types.QueryResult) (*types.QueryResult, error) {
	// 只在非错误状态时更新基本状态
	if queryStats.Error == nil {
		queryStats.UpdateBasicStatus(p.Config.EntrezParams.Filters != "")
	}

	return queryStats, queryStats.Error
}

// RetryFailedBatches 重试失败的批次
func (p *Pipeline) RetryFailedBatches(ctx context.Context, query *types.Query, cachedResult *types.QueryResult) *types.QueryResult {
	// 创建新的查询上下文
	queryCtx, cancel := context.WithTimeout(ctx, p.Config.Runtime.QueryTimeout)
	defer cancel()

	// 获取失败的批次信息
	failedBatches := cachedResult.FailedBatches
	if len(failedBatches) == 0 {
		return cachedResult
	}

	// 创建通道
	searchChan := make(chan *types.ESearchResult, 1)
	doneChan := make(chan struct{})

	// 启动查询和处理
	var wg sync.WaitGroup
	wg.Add(2)

	// 启动 Search
	go func() {
		defer wg.Done()
		p.Search.executeSearch(queryCtx, query, searchChan, cachedResult)
	}()

	// 启动 Summary
	go func() {
		defer wg.Done()
		p.Summary.RetryBatches(queryCtx, query, failedBatches, searchChan, cachedResult)
	}()

	// 等待所有操作完成
	go func() {
		wg.Wait()
		close(searchChan)
		close(doneChan)
	}()

	// 使用 waitForResults 处理结果，但保留原有的错误
	result, err := p.waitForResults(queryCtx, query, cachedResult, doneChan)
	if err != nil {
		// 如果发生错误，保留原有的失败批次信息
		return cachedResult
	}

	return result
}
