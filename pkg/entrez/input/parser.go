package input

import (
	"bufio"
	"fmt"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// FileParser 实现基因查询文件解析器
type FileParser struct {
	batchSize int
	flag      string
}

// NewFileParser 创建新的文件解析器
func NewFileParser(batchSize int, flag string) *FileParser {
	return &FileParser{
		batchSize: batchSize,
		flag:      flag,
	}
}

// ParseFile 解析输入文件生成查询列表
func (p *FileParser) ParseFile(filename string) ([]*types.Query, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(customerrors.ErrFailedOpenFile, "failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse 从 reader 解析查询
func (p *FileParser) Parse(reader io.Reader) ([]*types.Query, error) {
	scanner := bufio.NewScanner(reader)
	tempMap := make(map[string]struct{}) // 去重 map
	var searchItemList []*types.Query
	var lines []string

	sep := " OR " // 多个基因串之间的分隔符，必须是大写且左右两边留有空格

	for scanner.Scan() {
		line := scanner.Text()
		if _, ok := tempMap[line]; !ok {
			tempMap[line] = struct{}{}

			switch p.batchSize {
			case 1:
				searchItemList = append(searchItemList, types.NewQuery(line+p.flag))
			default:
				lines = append(lines, line)
				if len(lines) == p.batchSize {
					query := p.buildBatchQuery(lines, sep)
					searchItemList = append(searchItemList, types.NewQuery(query))
					lines = lines[:0]
				}
			}
		}
	}

	// 处理最后一批数据
	if p.batchSize != 1 && len(lines) > 0 {
		query := p.buildBatchQuery(lines, sep)
		searchItemList = append(searchItemList, types.NewQuery(query))
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("scan file error: %w", err)
	}

	return searchItemList, nil
}

// buildBatchQuery 构建批量查询字符串
func (p *FileParser) buildBatchQuery(lines []string, sep string) string {
	var builder strings.Builder
	for i, line := range lines {
		if i == 0 {
			builder.WriteString(line + p.flag)
		} else {
			builder.WriteString(sep + line + p.flag)
		}
	}
	return builder.String()
}
