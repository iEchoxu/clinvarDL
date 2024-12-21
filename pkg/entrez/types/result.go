package types

import (
	"fmt"
	"sync"
	"time"
)

const (
	QueryStatusSuccess QueryStatus = "success" // 完全成功
	QueryStatusPartial QueryStatus = "partial" // 部分成功
	QueryStatusFailed  QueryStatus = "failed"  // 完全失败
)

// QueryStatus 查询状态
type QueryStatus string

// QueryResult 单个查询的统计信息
type QueryResult struct {
	QueryID             string          `json:"query_id"`                 // 查询ID
	Query               string          `json:"query"`                    // 查询内容
	TotalRecords        int             `json:"total_records"`            // esearch 返回的总记录数
	ProcessedCount      int             `json:"processed_count"`          // 成功处理的记录数
	Status              QueryStatus     `json:"status"`                   // 查询状态（完全成功/部分成功/失败）
	FailedBatches       []BatchInfo     `json:"failed_batches,omitempty"` // 失败的批次信息（如果有）
	TotalBatches        int             `json:"total_batches"`            // 总批次数
	Error               error           `json:"error,omitempty"`          // 错误信息（如果有）
	CreatedAt           time.Time       `json:"created_at"`               // 开始时间
	EndTime             time.Time       `json:"end_time"`                 // 结束时间
	Duration            string          `json:"duration"`                 // 耗时
	Progress            string          `json:"progress"`                 // 进度
	Result              *ESummaryResult `json:"result"`                   // 查询结果
	LastQueryHasFilters bool            `json:"last_query_has_filters"`   // 上一次查询是否有过滤条件
	mu                  sync.Mutex      `json:"-"`
}

// NewQueryResult 创建新的查询结果
func NewQueryResult(queryID, query string) *QueryResult {
	now := time.Now()
	return &QueryResult{
		QueryID:             queryID,
		Query:               query,
		TotalRecords:        0,
		ProcessedCount:      0,
		Status:              QueryStatusFailed, // 默认为失败状态
		FailedBatches:       make([]BatchInfo, 0),
		TotalBatches:        0,
		Error:               nil,
		CreatedAt:           now,
		EndTime:             now, // 初始化为创建时间
		Duration:            "0s",
		Progress:            "0.00%",
		Result:              nil,
		LastQueryHasFilters: false,
	}
}

func (qr *QueryResult) SetTotalRecords(count int) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.TotalRecords = count
}

func (qr *QueryResult) SetTotalBatches(count int) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.TotalBatches = count
}

func (qr *QueryResult) AddProcessedRecords(count int) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.ProcessedCount += count
}

func (qr *QueryResult) AddFailedBatches(batchInfo BatchInfo) {
	qr.mu.Lock()
	defer qr.mu.Unlock()

	// 检查是否已存在相同批次号的记录
	for _, batch := range qr.FailedBatches {
		if batch.Start == batchInfo.Start {
			return
		}
	}

	qr.FailedBatches = append(qr.FailedBatches, batchInfo)
}

func (qr *QueryResult) IsComplete() bool {
	return qr.TotalRecords == qr.ProcessedCount
}

func (qr *QueryResult) GetProgress() float64 {
	if qr.ProcessedCount == 0 || qr.TotalRecords == 0 {
		return 0
	}

	if qr.ProcessedCount == qr.TotalRecords {
		return 100
	}

	return float64(qr.ProcessedCount) / float64(qr.TotalRecords) * 100
}

func (qr *QueryResult) GetProgressString() string {
	return fmt.Sprintf("%.2f%%", qr.GetProgress())
}

func (qr *QueryResult) GetQueryTime() float64 {
	return qr.EndTime.Sub(qr.CreatedAt).Seconds()
}

func (qr *QueryResult) GetQueryTimeString() string {
	return fmt.Sprintf("%.2fs", qr.GetQueryTime())
}

func (qr *QueryResult) GetQueryStatus() QueryStatus {
	switch qr.GetProgress() {
	case 0:
		return QueryStatusFailed
	case 100:
		return QueryStatusSuccess
	default:
		return QueryStatusPartial
	}
}

func (qr *QueryResult) GetQueryStatusContent() string {
	switch qr.GetQueryStatus() {
	case QueryStatusSuccess:
		return "success"
	case QueryStatusPartial:
		return "partial"
	case QueryStatusFailed:
		return "failed"
	default:
		return ""
	}
}

// RemoveFailedBatch 从失败批次列表中移除指定批次
func (qr *QueryResult) RemoveFailedBatch(start int) {
	qr.mu.Lock()
	defer qr.mu.Unlock()

	for i, batch := range qr.FailedBatches {
		if batch.Start == start {
			qr.FailedBatches = append(qr.FailedBatches[:i], qr.FailedBatches[i+1:]...)
			break
		}
	}
}

// updateBasicStatus 更新查询的基本状态（内部方法，调用前需要持有锁）
func (qr *QueryResult) updateBasicStatus(hasFilters bool) {
	qr.EndTime = time.Now()
	qr.Duration = qr.GetQueryTimeString()
	qr.Progress = qr.GetProgressString()
	qr.LastQueryHasFilters = hasFilters
}

// SetStatusOnError 在发生错误时更新查询状态
func (qr *QueryResult) SetStatusOnError(err error, hasFilters bool) {
	qr.mu.Lock()
	defer qr.mu.Unlock()

	qr.ProcessedCount = 0
	qr.Status = QueryStatusFailed
	qr.Error = err
	qr.Result = nil
	qr.updateBasicStatus(hasFilters)
}

// UpdateBasicStatus 更新查询的基本状态（公开方法）
func (qr *QueryResult) UpdateBasicStatus(hasFilters bool) {
	qr.mu.Lock()
	defer qr.mu.Unlock()
	qr.updateBasicStatus(hasFilters)
	qr.Status = qr.GetQueryStatus() // 正常情况下通过计算更新状态
}
