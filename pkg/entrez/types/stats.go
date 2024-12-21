package types

import (
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"

	"sync"
	"sync/atomic"
)

// Stats 用于跟踪查询统计信息
// 完全成功的查询：增加 CompletedQueries 计数
// 部分成功的查询：记录在 PartialFailures 中，并保存失败批次信息
// 完全失败的查询：记录在 FailedQueries 中
type Stats struct {
	TotalQueries     int       // 总查询数
	CompletedQueries int32     // 执行成功的查询数
	TotalRecords     int32     // 每个 esearch 返回的记录数之和
	ProcessedRecords int32     // 成功获取的 esummary 记录数之和
	FailedQueries    *sync.Map // 失败的查询集合
	failedCount      int       // 失败的查询数量
	PartialFailures  *sync.Map // 部分失败的查询及其失败记录
}

// NewStats 创建新的统计对象
func NewStats() *Stats {
	s := &Stats{
		FailedQueries:   &sync.Map{},
		PartialFailures: &sync.Map{},
	}
	return s
}

// AddPartialFailures 记录部分成功的查询
func (s *Stats) AddPartialFailures(queryID string, failedBatches Batch) {
	s.PartialFailures.Store(queryID, failedBatches)
}

// SetTotalQueries 设置总查询数
func (s *Stats) SetTotalQueries(count int) {
	s.TotalQueries = count
}

// AddTotalRecords 添加记录数
func (s *Stats) AddTotalRecords(count int) {
	atomic.AddInt32(&s.TotalRecords, int32(count))
}

// AddProcessedRecords 添加处理的记录数
func (s *Stats) AddProcessedRecords(count int) {
	atomic.AddInt32(&s.ProcessedRecords, int32(count))
}

// AddCompletedQuery 增加完成的查询数
func (s *Stats) AddCompletedQuery() {
	atomic.AddInt32(&s.CompletedQueries, 1)
}

// PrintSummary 打印统计摘要
func (s *Stats) PrintSummary() {
	logcdl.Info("query statistics:")
	logcdl.Info("- total queries: %d", s.TotalQueries)
	logcdl.Info("- completed queries: %d", s.CompletedQueries)
	logcdl.Info("- total records: %d", s.TotalRecords)
	logcdl.Info("- records processed: %d", s.ProcessedRecords)

	// 打印部分成功的查询详情
	var partialCount int
	s.PartialFailures.Range(func(key, value interface{}) bool {
		if partialCount == 0 {
			logcdl.Warn("queries with missing batches:")
		}
		info := value.(Batch)
		logcdl.Warn("  - query '%v':", key)
		for _, batch := range info.BatchInfos {
			logcdl.Warn("    - failed batch %d/%d: (start=%d, size=%d)",
				batch.BatchNum, info.Batches, batch.Start, batch.Size)
		}
		partialCount++
		return true
	})

	// 打印失败的查询详情
	var count int
	s.FailedQueries.Range(func(key, value interface{}) bool {
		if count == 0 {
			logcdl.Warn("failed queries details:")
		}
		logcdl.Warn("  - query '%v': %v", key, value)
		count++
		return true
	})

	s.failedCount = count
}

// AllQueriesFailed 检查是否所有查询都失败了
func (s *Stats) AllQueriesFailed() bool {
	return s.failedCount == s.TotalQueries
}
