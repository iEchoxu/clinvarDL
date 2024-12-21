package entrez

import (
	"context"
	"fmt"
	"sync"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/cache"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/config"
	customHttp "github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pipeline"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/utils"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

type QueryExecutor struct {
	stats       *types.Stats // 统计信息
	Config      *config.Config
	httpClient  *customHttp.Client
	rateLimiter *customHttp.RateLimiter
	cache       cache.Cache
}

func NewQueryExecutor(config *config.Config) *QueryExecutor {
	stats := types.NewStats()

	// 缓存启用时初始化缓存
	var cacheStore cache.Cache
	if config.Cache.Enabled {
		var err error
		cacheStore, err = cache.NewFileCache(config.Cache.Dir, config.Cache.TTL)
		if err != nil {
			logcdl.Warn("failed to create cache: %v", err)
		} else {
			logcdl.Info("cache enabled at: '%s'", config.Cache.Dir)
		}
	}

	return &QueryExecutor{
		stats:       stats,
		Config:      config,
		httpClient:  customHttp.GetHTTPClient(),
		rateLimiter: customHttp.NewRateLimiter(config.EntrezParams.ApiKey != ""),
		cache:       cacheStore, // 如果缓存未启用，则使用默认零值 nil
	}
}

// executeQueries 执行查询并返回结果通道
func (q *QueryExecutor) executeQueries(ctx context.Context, queries []*types.Query) (<-chan *types.QueryResult, error) {
	logcdl.Info("created query executor with request rate of %.2f requests/second", q.rateLimiter.GetCurrentRate())

	bufferSize := utils.GetWorkerCount(q.Config.Runtime.BufferSize, len(queries))

	results := make(chan *types.QueryResult, bufferSize)

	// 获取查询并发数，不超过查询数，避免多开协程造成资源浪费
	maxWorkers := q.Config.Runtime.GetQueryWorkers()
	queryWorkers := utils.GetWorkerCount(maxWorkers, len(queries))

	logcdl.Tip("starting to submit %d queries with %d concurrent workers",
		len(queries), queryWorkers)

	// 启动查询执行
	q.processQueries(ctx, queries, results, queryWorkers)

	// 设置总查询数
	q.stats.SetTotalQueries(len(queries))

	// 处理统计信息
	q.stats.PrintSummary()

	// 如果所有查询都失败了，返回 ErrEmptyResult
	if q.stats.AllQueriesFailed() {
		return nil, customerrors.NewEmptyResultError(fmt.Sprintf("all %d queries failed, please check the logs and try again later", len(queries)))
	}

	return results, nil
}

// processQueries 处理所有查询
func (q *QueryExecutor) processQueries(ctx context.Context, queries []*types.Query, results chan<- *types.QueryResult, queryWorkers int) {
	defer close(results)

	// 创建工作池
	semaphore := make(chan struct{}, queryWorkers)

	var wg sync.WaitGroup

	// 执行查询
	for _, query := range queries {
		select {
		case <-ctx.Done():
			logcdl.Warn("context cancelled for %v, stopping query execution", query)
			q.stats.FailedQueries.Store(query.GetQueryID(), ctx.Err())
			return
		case semaphore <- struct{}{}:
			wg.Add(1)
			go q.executeQuery(ctx, query, results, semaphore, &wg)
		}
	}

	wg.Wait()
}

// executeQuery 执行单个查询
func (q *QueryExecutor) executeQuery(ctx context.Context, query *types.Query, results chan<- *types.QueryResult, semaphore chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() { <-semaphore }()

	queryID := query.GetQueryID()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("worker panic for query %v: %v", queryID, r)
			q.stats.FailedQueries.Store(queryID, err)
			logcdl.Error("recovered from panic in worker for query %v: %v", queryID, r)
		}
	}()

	// 先尝试从缓存获取
	if result := q.tryGetFromCache(ctx, query); result != nil {
		// 处理缓存命中
		results <- result
		q.stats.AddProcessedRecords(result.ProcessedCount)
		q.stats.AddTotalRecords(result.TotalRecords)
		if result.Status == types.QueryStatusSuccess {
			q.stats.AddCompletedQuery()
		}
		return
	}

	// 缓存未命中，执行单个查询的完整流程
	queryStats, err := pipeline.NewPipeline(q.Config, q.httpClient, q.rateLimiter).ExecuteQuery(ctx, query)
	// 如果单个查询失败，则添加到失败查询列表且不生成缓存数据,下次查询时会发起新的 NewPipeline
	if err != nil {
		q.stats.FailedQueries.Store(queryID, err)
		return
	}

	// 缓存结果
	if q.cache != nil && queryStats != nil {
		if queryStats.Result == nil {
			q.stats.FailedQueries.Store(queryID, fmt.Errorf("query result is nil"))
			return
		}
		if err := q.cache.Set(queryID, queryStats); err != nil {
			logcdl.Warn("failed to cache result for query '%v': %v", queryID, err)
		}
	}

	// 发送结果
	select {
	case results <- &types.QueryResult{Query: query.Content, Result: queryStats.Result}:
		if queryStats.Status == types.QueryStatusSuccess {
			q.stats.AddCompletedQuery()
		}
		q.stats.AddProcessedRecords(queryStats.ProcessedCount)
		q.stats.AddTotalRecords(queryStats.TotalRecords)
		// 添加部分失败批次的打印信息
		if len(queryStats.FailedBatches) > 0 {
			failedBatches := types.Batch{
				Batches:    queryStats.TotalBatches,
				BatchInfos: queryStats.FailedBatches,
			}
			q.stats.AddPartialFailures(queryID, failedBatches)
		}
	case <-ctx.Done():
		logcdl.Error("context cancelled while sending results for query '%v'", query)
		q.stats.FailedQueries.Store(queryID, ctx.Err())
	}
}

// tryGetFromCache 尝试从缓存获取结果
func (q *QueryExecutor) tryGetFromCache(ctx context.Context, query *types.Query) *types.QueryResult {
	// 如果缓存未启用，则直接返回 nil
	if q.cache == nil {
		return nil
	}

	queryID := query.GetQueryID()
	queryResult, err := q.cache.Get(queryID)
	if err != nil {
		logcdl.Debug("cache get failed for query '%v': %v", queryID, err)
		return nil
	}

	// 检查过滤条件是否发生变化, 如果发生变化，则返回 nil，更新缓存
	currentHasFilters := q.Config.EntrezParams.Filters != ""
	if queryResult.LastQueryHasFilters != currentHasFilters {
		logcdl.Info("filter conditions changed for query '%v', ignoring cache", queryID)
		return nil
	}

	// 如果缓存结果完整，直接返回
	if queryResult.IsComplete() {
		logcdl.Info("using complete cached result for query '%v'", queryID)
		return queryResult
	}

	// 尝试重试失败的批次, 如果成功，则更新缓存
	// 如果失败，则返回原有缓存结果
	updatedResult, err := q.retryFailedBatches(ctx, query, queryResult)
	if err != nil {
		logcdl.Warn("failed to retry failed batches for query '%v': %v", queryID, err)
		return queryResult
	}

	if updatedResult != nil && updatedResult.IsComplete() {
		logcdl.Success("successfully completed all batches for query '%v'", queryID)
		return updatedResult
	}

	return queryResult
}

// retryFailedBatches 重试失败的批次并更新缓存
func (q *QueryExecutor) retryFailedBatches(ctx context.Context, query *types.Query, cachedResult *types.QueryResult) (*types.QueryResult, error) {
	queryID := query.GetQueryID()
	logcdl.Info("retrying failed batches for query '%v'", queryID)

	// 创建新的 Pipeline 实例
	p := pipeline.NewPipeline(q.Config, q.httpClient, q.rateLimiter)

	// 重试失败的批次
	updatedResult := p.RetryFailedBatches(ctx, query, cachedResult)

	// 检查重试结果
	if updatedResult == nil {
		return nil, fmt.Errorf("retry failed: nil result")
	}

	// 如果重试失败，返回错误但不更新缓存
	if updatedResult.Error != nil {
		return nil, updatedResult.Error
	}

	// 更新缓存
	if q.cache != nil {
		// 只有在状态为成功或部分成功时才更新缓存
		if updatedResult.Status != types.QueryStatusFailed {
			if err := q.cache.Set(queryID, updatedResult); err != nil {
				logcdl.Warn("failed to update cache for query '%v': %v", queryID, err)
			}
		}
	}

	return updatedResult, nil
}
