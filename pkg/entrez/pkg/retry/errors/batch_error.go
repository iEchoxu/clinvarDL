package errors

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

// BatchError 定义批处理错误
type BatchError struct {
	BatchNum int          // 当前批次
	Batches  int          // 总批次
	Start    int          // 起始位置
	Size     int          // 批次大小
	Query    *types.Query // 查询ID
	Err      error        // 原始错误
}

// Error 实现 error 接口
func (e *BatchError) Error() string {
	return fmt.Sprintf("batch %d/%d (start=%d) failed: %v", e.BatchNum, e.Batches, e.Start, e.Err)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *BatchError) Unwrap() error {
	return e.Err
}

// NewBatchError 创建新的批处理错误
func NewBatchError(batchNum, batches, start, size int, query *types.Query, err error) *BatchError {
	return &BatchError{
		BatchNum: batchNum,
		Batches:  batches,
		Start:    start,
		Size:     size,
		Query:    query,
		Err:      err,
	}
}
