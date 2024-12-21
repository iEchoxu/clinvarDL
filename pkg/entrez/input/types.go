package input

import (
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
	"io"
)

// Parser 定义解析器接口
type Parser interface {
	Parse(reader io.Reader) ([]*types.Query, error)
}
