package utils

// Min 返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetWorkerCount 根据最大工作者数和任务数计算实际需要的工作者数
// 避免创建过多不必要的 goroutine
func GetWorkerCount(maxWorkers, taskCount int) int {
	return Min(maxWorkers, taskCount)
}
