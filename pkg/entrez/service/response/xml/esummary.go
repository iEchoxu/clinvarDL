package xml

import (
	"encoding/xml"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/pkg/errors"

	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
)

type ESummaryResponseParser struct{}

// ParseESummary 解析 ESummary XML 响应
func (h *ESummaryResponseParser) ParseESummary(data []byte) (*types.ESummaryResult, error) {
	var result types.ESummaryResult
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrapf(customerrors.ErrParse, "esummary xml unmarshal failed: %v", err)
	}
	return &result, nil
}
