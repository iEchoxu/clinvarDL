package service

import (
	"context"
	"net/url"
	"strings"

	"github.com/iEchoxu/clinvarDL/pkg/entrez/service/response"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"

	"github.com/pkg/errors"
)

// EPostOperation 实现了 epost 操作
type EPostOperation struct {
	BaseOperation
}

// NewEPostOperation 创建一个新的 EPostOperation 实例
func NewEPostOperation() *EPostOperation {
	return &EPostOperation{
		BaseOperation: BaseOperation{
			BaseURL:    BaseURLEPost,
			Parameters: url.Values{},
		},
	}
}

func (p *EPostOperation) Execute(ctx context.Context, ids []string, query *types.Query) (*types.EPostResult, error) {
	p.Parameters.Set("id", strings.Join(ids, ","))
	epostURL, err := p.BuildURL()
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to build epost url for query '%v'", query)
	}

	body, err := p.doRequest(ctx, "POST", epostURL.String(), p.Parameters)
	if err != nil {
		return nil, errors.WithMessagef(err, "epost http request failed for query '%v'", query)
	}

	// 创建解析器
	parser, err := response.NewEPostResponseParser(response.ParserType(p.GetRetMode()))
	if err != nil {
		return nil, err
	}

	// 解析响应``
	result, err := parser.ParseEPost(body)
	if err != nil {
		return nil, err
	}

	return result, nil
}
