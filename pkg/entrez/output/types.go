package output

import (
	"context"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

// Writer 定义结果写入器接口
type Writer interface {
	SetHeaders(headers []string) error
	WriteResultStream(ctx context.Context, results <-chan *types.QueryResult) error
	Save(filename string) error
	Close() error
}
