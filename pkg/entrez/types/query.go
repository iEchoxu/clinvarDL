package types

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// Query 定义查询信息
type Query struct {
	Content string // 查询内容
}

// NewQuery 创建新的查询
func NewQuery(content string) *Query {
	return &Query{
		Content: content,
	}
}

// GetQueryID 生成并返回查询的唯一标识符
func (q *Query) GetQueryID() string {
	// 取查询内容的第一个基因名作为前缀
	prefix := extractFirstTerm(q.Content)

	// 使用 MD5 生成哈希（也可用 sha256）
	hash := md5.Sum([]byte(q.Content))
	// 取前 6 位作为 ID
	shortHash := hex.EncodeToString(hash[:])[:6]

	return fmt.Sprintf("%s-%s", prefix, shortHash)
}

// String 实现 Stringer 接口
func (q *Query) String() string {
	return q.GetQueryID()
}

// extractFirstTerm 提取第一个搜索词
func extractFirstTerm(content string) string {
	// 移除多余的空格
	content = strings.TrimSpace(content)

	// 处理空查询
	if content == "" {
		return "EMPTY"
	}

	// 正则表达式匹配模式
	patterns := []string{
		`(\w+)\[gene\]`,    // 匹配基因
		`(\w+)\[protein\]`, // 匹配蛋白质
		`(\w+)\[title\]`,   // 匹配标题
		`"([^"]+)"`,        // 匹配引号内的内容
		`(\w+)`,            // 匹配任意单词
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}

	// 如果没有匹配到任何模式，返回前 10 个字符
	if len(content) > 10 {
		return content[:10]
	}
	return content
}
