package types

// BatchInfo 批次信息
type BatchInfo struct {
	BatchNum int    // 批次编号
	Start    int    // 起始记录数
	Size     int    // 批次大小
	ErrMsg   string `json:"err_msg,omitempty"` // 使用字符串存储错误信息
}

type Batch struct {
	Batches    int
	BatchInfos []BatchInfo
}

// NewBatch 计算批次信息
func NewBatch(totalCount, batchSize int) *Batch {
	batches := (totalCount + batchSize - 1) / batchSize
	batchInfos := make([]BatchInfo, batches)

	for i := 0; i < batches; i++ {
		start := i * batchSize
		size := batchSize
		if start+size > totalCount {
			size = totalCount - start
		}
		batchInfos[i] = BatchInfo{
			BatchNum: i + 1,
			Start:    start,
			Size:     size,
		}
	}

	return &Batch{
		Batches:    batches,
		BatchInfos: batchInfos,
	}
}
