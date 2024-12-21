package response

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/service/response/xml"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/types"
)

// ESearchResponseParser 定义 ESearch 响应解析器接口
type ESearchResponseParser interface {
	ParseESearch(data []byte) (*types.ESearchResult, error)
}

// ESummaryResponseParser 定义 ESummary 响应解析器接口
type ESummaryResponseParser interface {
	ParseESummary(data []byte) (*types.ESummaryResult, error)
}

// EPostResponseParser 定义 EPost 响应解析器接口
type EPostResponseParser interface {
	ParseEPost(data []byte) (*types.EPostResult, error)
}

// ParserType 定义解析器类型
type ParserType string

const (
	ParserXML  ParserType = "xml"
	ParserJSON ParserType = "json"
)

// NewESearchResponseParser 根据类型创建 ESearch 响应解析器
func NewESearchResponseParser(parserType ParserType) (ESearchResponseParser, error) {
	switch parserType {
	case ParserXML:
		return &xml.ESearchResponseParser{}, nil
	case ParserJSON:
		return nil, fmt.Errorf("JSON parser not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported parser type: %s", parserType)
	}
}

// NewESummaryResponseParser 根据类型创建 ESummary 响应解析器
func NewESummaryResponseParser(parserType ParserType) (ESummaryResponseParser, error) {
	switch parserType {
	case ParserXML:
		return &xml.ESummaryResponseParser{}, nil
	case ParserJSON:
		return nil, fmt.Errorf("JSON parser not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported parser type: %s", parserType)
	}
}

// NewEPostResponseParser 根据类型创建 EPost 响应解析器
func NewEPostResponseParser(parserType ParserType) (EPostResponseParser, error) {
	switch parserType {
	case ParserXML:
		return &xml.EPostResponseParser{}, nil
	case ParserJSON:
		return nil, fmt.Errorf("JSON parser not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported parser type: %s", parserType)
	}
}
