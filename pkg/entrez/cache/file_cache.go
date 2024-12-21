package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

// FileCache 实现基于文件的缓存
type FileCache struct {
	CacheDir string                        // 缓存目录
	Data     map[string]*types.QueryResult // 缓存数据
	mu       sync.RWMutex                  // 读写锁
	TTL      time.Duration                 // 缓存过期时间
}

// NewFileCache 创建新的文件缓存
func NewFileCache(cacheDir string, ttl time.Duration) (*FileCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &FileCache{
		CacheDir: cacheDir,
		Data:     make(map[string]*types.QueryResult),
		TTL:      ttl,
	}, nil
}

// Get 实现 Cache 接口
func (c *FileCache) Get(queryID string) (*types.QueryResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 先从内存缓存中查找
	if entry, ok := c.Data[queryID]; ok {
		// 检查是否过期
		if time.Since(entry.CreatedAt) > c.TTL {
			delete(c.Data, queryID)
			return nil, fmt.Errorf("cache expired for query '%v'", queryID)
		}
		// 检查结果是否有效
		if entry.Result == nil {
			return nil, fmt.Errorf("invalid cache entry for query '%v': nil result", queryID)
		}
		return entry, nil
	}

	// 从文件缓存中加载
	entry, err := c.loadFromFile(queryID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cache from file for query '%v': %w", queryID, err)
	}

	// 检查是否过期
	if time.Since(entry.CreatedAt) > c.TTL {
		// 删除过期的缓存文件
		filePath := filepath.Join(c.CacheDir, queryID+".json")
		if err := os.Remove(filePath); err != nil {
			logcdl.Warn("failed to remove expired cache file for query '%v': %v", queryID, err)
		}
		return nil, fmt.Errorf("cache expired for query '%v'", queryID)
	}

	// 检查结果是否有效
	if entry.Result == nil {
		return nil, fmt.Errorf("invalid cache entry loaded from file for query '%v': nil result", queryID)
	}

	// 加载到内存缓存
	c.Data[queryID] = entry
	logcdl.Info("loaded cache from file for query '%v'", queryID)

	return entry, nil
}

// Set 实现 Cache 接口
func (c *FileCache) Set(queryID string, entry *types.QueryResult) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 验证入参
	if entry == nil || entry.Result == nil {
		return fmt.Errorf("invalid cache entry: nil result")
	}

	// 更新内存缓存
	c.Data[queryID] = entry

	// 保存到文件
	return c.saveToFile(queryID, entry)
}

// loadFromFile 从文件加载缓存
func (c *FileCache) loadFromFile(queryID string) (*types.QueryResult, error) {
	filePath := filepath.Join(c.CacheDir, queryID+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var entry types.QueryResult
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	// 检查 Result  是否为 nil
	if entry.Result == nil {
		logcdl.Warn("invalid cache data for query '%v': nil result", queryID)
		return nil, fmt.Errorf("invalid cache data: nil result")
	}

	return &entry, nil
}

// saveToFile 保存缓存到文件
func (c *FileCache) saveToFile(queryID string, entry *types.QueryResult) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	filePath := filepath.Join(c.CacheDir, queryID+".json")
	return os.WriteFile(filePath, data, 0644)
}

// CleanExpired 清理所有过期的缓存
func (c *FileCache) CleanExpired() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 遍历缓存目录
	entries, err := os.ReadDir(c.CacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %v", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 加载缓存文件
		filePath := filepath.Join(c.CacheDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			logcdl.Warn("failed to read cache file %s: %v", filePath, err)
			continue
		}

		var result types.QueryResult
		if err := json.Unmarshal(data, &result); err != nil {
			logcdl.Warn("failed to unmarshal cache file %s: %v", filePath, err)
			continue
		}

		// 检查是否过期
		if now.Sub(result.CreatedAt) > c.TTL {
			// 删除内存缓存
			queryID := strings.TrimSuffix(entry.Name(), ".json")
			delete(c.Data, queryID)

			// 删除文件缓存
			if err := os.Remove(filePath); err != nil {
				logcdl.Warn("failed to remove expired cache file %s: %v", filePath, err)
			}
		}
	}

	return nil
}
