package xml

import (
	"encoding/xml"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/pkg/errors"
)

type EPostResponseParser struct{}

func (p *EPostResponseParser) ParseEPost(data []byte) (*types.EPostResult, error) {
	var result types.EPostResult
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, errors.Wrapf(customerrors.ErrParse, "epost xml unmarshal failed: %v", err)
	}
	return &result, nil
}
