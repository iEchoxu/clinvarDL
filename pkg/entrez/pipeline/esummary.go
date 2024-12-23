package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/config"
	customHttp "github.com/iEchoxu/clinvarDL/pkg/entrez/http"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/utils"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/service"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/pkg/errors"
)

type EsummaryExecutor struct {
	Esummary *service.ESummaryOperation
	Config   *config.Config
}

func NewEsummaryExecutor(config *config.Config, httpClient *customHttp.Client, rateLimiter *customHttp.RateLimiter) *EsummaryExecutor {
	esummaryExecutor := &EsummaryExecutor{
		Esummary: service.NewESummaryOperation(httpClient, rateLimiter),
		Config:   config,
	}

	esummaryExecutor.setupOperations()

	return esummaryExecutor
}

func (e *EsummaryExecutor) setupOperations() {
	// 配置 eSummary 查询参数
	e.Esummary.SetDB(e.Config.EntrezParams.DB).
		SetRetMode(e.Config.EntrezParams.RetMode.String()).
		SetEmail(e.Config.EntrezParams.Email).
		SetApiKey(e.Config.EntrezParams.ApiKey).
		SetToolName(e.Config.EntrezParams.ToolName).
		SetUseStream(e.Config.Stream.GetEnabled())
}

// processSummaryFlow 执行 ESummary 流程
func (e *EsummaryExecutor) ProcessSummaryFlow(ctx context.Context, query *types.Query, searchChan chan *types.ESearchResult, collector *types.QueryResult) {
	// 等待并验证 ESearch 结果
	searchResult, err := e.waitForSearchResult(ctx, query, searchChan)
	// 当 esearch 发生错误或 ctx 超时直接返回
	if err != nil {
		var emptyError *customerrors.EmptyResultError
		if !errors.As(err, &emptyError) {
			collector.SetStatusOnError(err, e.Config.EntrezParams.Filters != "")
		}

		return
	}

	// 根据结果数量选择处理方式
	result, err := e.processSearchResult(ctx, query, searchResult, collector)
	if err != nil {
		collector.Result = result // 返回已收集的结果
		return
	}

	// 发送结果
	select {
	case <-ctx.Done():
		collector.Error = ctx.Err()
		collector.ProcessedCount = 0
		collector.Progress = collector.GetProgressString()
		collector.Result = nil
	default:
		// 统计搜索结果
		collector.Result = result
		collector.Error = nil
		collector.Status = collector.GetQueryStatus()
		collector.Progress = collector.GetProgressString()
	}
}

// waitForSearchResult 等待搜索结果
func (e *EsummaryExecutor) waitForSearchResult(ctx context.Context, query *types.Query, searchChan <-chan *types.ESearchResult) (*types.ESearchResult, error) {
	select {
	case searchResult := <-searchChan:
		if searchResult == nil {
			return nil, customerrors.NewEmptyResultError(fmt.Sprintf("nil search result for query '%v'", query))
		}

		// 验证搜索结果
		if searchResult.Count == 0 {
			return nil, customerrors.NewEmptyResultError(fmt.Sprintf("no records found for query '%v'", query))
		}

		if searchResult.WebEnv == "" || searchResult.QueryKey == "" {
			return nil, errors.Wrapf(customerrors.ErrInvalidParameter, "invalid WebEnv or QueryKey for query '%v'", query)
		}

		return searchResult, nil
	case <-ctx.Done():
		err := customerrors.NewTimeoutError(
			fmt.Sprintf("esummary request timed out after %v for '%v'",
				e.Config.Runtime.QueryTimeout, query),
			ctx.Err())
		return nil, err
	}
}

// processSearchResult 处理搜索结果
func (e *EsummaryExecutor) processSearchResult(ctx context.Context, query *types.Query, searchResult *types.ESearchResult, collector *types.QueryResult) (*types.ESummaryResult, error) {
	if searchResult.Count > e.Config.Runtime.BatchSize {
		return e.executeSummaryBatches(ctx, query, searchResult, collector)
	}

	return e.executeSingleSummary(ctx, query, searchResult, collector)
}

// executeSingleSummary 执行单次 ESummary 请求
func (e *EsummaryExecutor) executeSingleSummary(ctx context.Context, query *types.Query, searchResult *types.ESearchResult, collector *types.QueryResult) (*types.ESummaryResult, error) {
	logcdl.Info("using single request for query '%v' with %d records", query, searchResult.Count)

	collector.SetTotalBatches(1)

	config := retry.DefaultConfig()
	result, err := retry.DoWithRetry(ctx, fmt.Sprintf("single esummary request for query: '%v'", query), config, func() (*types.ESummaryResult, error) {
		result, err := e.executeSummary(ctx, searchResult.WebEnv, searchResult.QueryKey, 0, searchResult.Count, query)
		if err != nil {
			return nil, err
		}

		if result == nil || result.DocumentSummarySet.DocumentSummary == nil {
			return nil, customerrors.NewEmptyResultError("empty result from single esummary request")
		}

		return result, nil
	})

	// 所有重试失败后，记录失败的批次
	if err != nil {
		// 这里不添加进失败批次, 因为 单次请求模式下，失败就意味着整个查询失败，和 esearch 失败时的处理一样
		// 且不会生成对应的缓存，如果没有缓存则在下次执行相同查询时会发起新请求
		collector.SetStatusOnError(err, e.Config.EntrezParams.Filters != "")

		logcdl.Error("single esummary request failed after all retries for query '%v'", query)
		return nil, err
	}

	recordCount := len(result.DocumentSummarySet.DocumentSummary)

	collector.AddProcessedRecords(recordCount)

	if recordCount < collector.TotalRecords {
		logcdl.Warn("single esummary request returned fewer records than expected (got %d, expected %d) for query '%v'",
			recordCount, collector.TotalRecords, query)

		return result, nil
	}

	logcdl.Success("single esummary request completed successfully for query '%v' with %d records",
		query, recordCount)

	return result, nil
}

// executeSummaryBatches 使用 retstart 分批执行 ESummary 请求
func (e *EsummaryExecutor) executeSummaryBatches(ctx context.Context, query *types.Query, searchResult *types.ESearchResult, collector *types.QueryResult) (*types.ESummaryResult, error) {
	totalCount := searchResult.Count
	batchSize := e.Config.Runtime.BatchSize

	batch := types.NewBatch(totalCount, batchSize)

	collector.SetTotalBatches(batch.Batches) // 批次总数

	// 创建通道和工作者
	resultChan := make(chan *types.ESummaryResult, batch.Batches)

	// 获取 ESummary 并发数，不超过批次数量
	maxBatchWorkers := e.Config.Runtime.GetEsummaryWorkers()
	esummaryWorkers := utils.GetWorkerCount(maxBatchWorkers, batch.Batches)

	semaphore := make(chan struct{}, esummaryWorkers)

	logcdl.Tip("starting batch processing for esummary query '%v' with %d records in %d batches (size=%d) with %d workers",
		query, totalCount, batch.Batches, batchSize, esummaryWorkers)

	var wg sync.WaitGroup

	// 并发获取批次结果
	for _, info := range batch.BatchInfos {
		wg.Add(1)
		go func(info types.BatchInfo) {
			defer wg.Done()
			e.processBatch(ctx, info, searchResult, query, resultChan, semaphore, collector)
		}(info)
	}

	//  等待所有批次完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	return e.collectResults(ctx, query, resultChan, collector)
}

// processBatch 处理单个批次
func (e *EsummaryExecutor) processBatch(ctx context.Context, info types.BatchInfo, searchResult *types.ESearchResult, query *types.Query, resultChan chan<- *types.ESummaryResult, semaphore chan struct{}, collector *types.QueryResult) {
	select {
	case semaphore <- struct{}{}:
		defer func() { <-semaphore }()
	case <-ctx.Done():
		err := customerrors.NewTimeoutError(
			fmt.Sprintf("esummary request timed out after %v for '%v'",
				e.Config.Runtime.QueryTimeout, query),
			ctx.Err())
		collector.SetStatusOnError(err, e.Config.EntrezParams.Filters != "")
		return
	}

	logcdl.Info("processing batch %d/%d (start=%d, size=%d) for esummary query '%v'",
		info.BatchNum, collector.TotalBatches, info.Start, info.Size, query)

	config := retry.DefaultConfig()
	result, err := retry.DoWithRetry(ctx, fmt.Sprintf("esummary batch %d/%d (start=%d) for query '%v'", info.BatchNum, collector.TotalBatches, info.Start, query), config, func() (*types.ESummaryResult, error) {
		result, err := e.executeSummary(ctx, searchResult.WebEnv, searchResult.QueryKey, info.Start, info.Size, query)
		if err != nil {
			return nil, err
		}

		// 空结果检查
		if result == nil || result.DocumentSummarySet.DocumentSummary == nil || len(result.DocumentSummarySet.DocumentSummary) == 0 {
			return nil, customerrors.NewEmptyResultError("server returned empty result for query")
		}

		return result, nil
	})

	// 如果达到最大重试次数，记录该批次的信息
	if err != nil {
		logcdl.Error("esummary batch %d/%d (start=%d) failed after all retries for query '%v'",
			info.BatchNum, collector.TotalBatches, info.Start, query)

		// 记录失败批次
		// 不记录错误，因为单个批次失败不意味着整个查询失败
		collector.AddFailedBatches(types.BatchInfo{
			BatchNum: info.BatchNum,
			Start:    info.Start,
			Size:     info.Size,
			ErrMsg:   err.Error(),
		})

		return
	}

	select {
	case resultChan <- result:
		// 只有在重试模式下需要移除失败批次
		if len(collector.FailedBatches) > 0 {
			collector.RemoveFailedBatch(info.Start)
		}

		recordCount := len(result.DocumentSummarySet.DocumentSummary)

		// 验证每个批次返回的记录数是否符预期(与 retmax 较)
		// 对于批次返回的记录数小于预期值，则打印警告，未做处理
		if recordCount < info.Size {
			logcdl.Warn("esummary batch %d/%d (start=%d) returned fewer records than expected (got %d, expected %d) for query '%v'",
				info.BatchNum, collector.TotalBatches, info.Start, recordCount, info.Size, query)
			return
		}

		logcdl.Success("esummary batch %d/%d (start=%d) completed for query '%v' with %d records",
			info.BatchNum, collector.TotalBatches, info.Start, query, recordCount)
	case <-ctx.Done():
		collector.AddFailedBatches(types.BatchInfo{
			BatchNum: info.BatchNum,
			Start:    info.Start,
			Size:     info.Size,
			ErrMsg:   ctx.Err().Error(),
		})
	}
}

// collectResults 收集所有批次的结果
func (e *EsummaryExecutor) collectResults(ctx context.Context, query *types.Query, resultChan <-chan *types.ESummaryResult, collector *types.QueryResult) (*types.ESummaryResult, error) {
	// 完整初始化结构体
	combinedResult := types.ESummaryResult{
		DocumentSummarySet: types.DocumentSummarySet{
			DocumentSummary: make([]*types.DocumentSummary, 0),
		},
	}

	// 收集结果
	for i := 0; i < collector.TotalBatches; i++ {
		select {
		case result := <-resultChan:
			if result != nil && result.DocumentSummarySet.DocumentSummary != nil {
				recordCount := len(result.DocumentSummarySet.DocumentSummary)
				collector.AddProcessedRecords(recordCount) // 添加已处理记录数 (可以是缓存数据)

				combinedResult.DocumentSummarySet.DocumentSummary = append(
					combinedResult.DocumentSummarySet.DocumentSummary,
					result.DocumentSummarySet.DocumentSummary...)

				logcdl.Info("esummary progress for query '%v': %d/%d records (%.1f%%)",
					query, collector.ProcessedCount, collector.TotalRecords,
					float64(collector.ProcessedCount)/float64(collector.TotalRecords)*100)
			}
		case <-ctx.Done():
			// 返回已收集的结果和错误
			return &combinedResult, customerrors.NewTimeoutError(
				fmt.Sprintf("esummary request timed out after %v for '%v'",
					e.Config.Runtime.QueryTimeout, query),
				ctx.Err())
		}
	}

	// 打印最终统计信息
	logcdl.Info("summary of results for query '%v':", query)
	logcdl.Info("- total count from esearch: %d", collector.TotalRecords)
	logcdl.Info("- records processed: %d", collector.ProcessedCount)
	logcdl.Info("- failed to retrieve records: %d", collector.TotalRecords-collector.ProcessedCount)

	// 如果有部分批次的数据没获取到，返回已获取到的数据
	if collector.ProcessedCount < collector.TotalRecords {
		return &combinedResult, nil
	}

	// 所有批次都成功, 返回成功数据
	if collector.ProcessedCount == collector.TotalRecords {
		logcdl.Success("completed all esummary batches for query '%v', total records: %d",
			query, collector.ProcessedCount)
		return &combinedResult, nil
	}

	// 所有批次都失败，返回错误
	return nil, fmt.Errorf("all esummary batches failed for query '%v", query)
}

// executeSummary 执行单个批次的 ESummary 请求
func (e *EsummaryExecutor) executeSummary(ctx context.Context, webEnv, queryKey string, start, batchSize int, query *types.Query) (*types.ESummaryResult, error) {
	// 检查 context 是否已取
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 设置分页参数
	e.Esummary.Parameters.Set("retstart", fmt.Sprintf("%d", start))
	e.Esummary.Parameters.Set("retmax", fmt.Sprintf("%d", batchSize))

	// 执 ESummary 请求
	result, err := e.Esummary.Execute(ctx, webEnv, queryKey, query)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RetryBatches 重试多个批次的 ESummary 请求
func (e *EsummaryExecutor) RetryBatches(ctx context.Context, query *types.Query, batches []types.BatchInfo, searchChan <-chan *types.ESearchResult, collector *types.QueryResult) {
	searchResult, err := e.waitForSearchResult(ctx, query, searchChan)
	if err != nil {
		var emptyError *customerrors.EmptyResultError
		if !errors.As(err, &emptyError) {
			collector.SetStatusOnError(err, e.Config.EntrezParams.Filters != "")
		}
		return
	}

	resultChan := make(chan *types.ESummaryResult, len(batches))

	// 获取 ESummary 并发数
	maxBatchWorkers := e.Config.Runtime.GetEsummaryWorkers()
	esummaryWorkers := utils.GetWorkerCount(maxBatchWorkers, len(batches))

	// 创建工作池
	semaphore := make(chan struct{}, esummaryWorkers)
	var wg sync.WaitGroup

	// 并发处理每个批次
	for _, batch := range batches {
		wg.Add(1)
		go func(batch types.BatchInfo) {
			defer wg.Done()
			e.processBatch(ctx, batch, searchResult, query, resultChan, semaphore, collector)
		}(batch)
	}

	// 等待所有批次处理完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	result, err := e.collectResults(ctx, query, resultChan, collector)
	if err != nil {
		logcdl.Error("failed to collect some results: %v", err)
		collector.Error = err
		return
	}

	// 合并新的结果到现有结果中
	if result != nil && result.DocumentSummarySet.DocumentSummary != nil {
		collector.Result.DocumentSummarySet.DocumentSummary = append(
			collector.Result.DocumentSummarySet.DocumentSummary,
			result.DocumentSummarySet.DocumentSummary...)
	}
}
