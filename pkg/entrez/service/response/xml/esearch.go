package xml

import (
	"encoding/xml"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/pkg/errors"

	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
)

// ESearchResponseParser XML 格式的 ESearch 响应解析器
type ESearchResponseParser struct{}

// ParseESearch 解析 ESearch XML 响应
func (p *ESearchResponseParser) ParseESearch(data []byte) (*types.ESearchResult, error) {
	var result types.ESearchResult
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrapf(customerrors.ErrParse, "esearch xml unmarshal failed: %v", err)
	}
	return &result, nil
}
