package cache

import "github.com/iEchoxu/clinvarDL/pkg/entrez/types"

// Cache 定义缓存接口
type Cache interface {
	// Get 获取缓存的查询结果
	Get(queryID string) (*types.QueryResult, error)

	// Set 设置查询结果缓存
	Set(queryID string, entry *types.QueryResult) error

	// CleanExpired 清理过期的缓存
	CleanExpired() error
}
