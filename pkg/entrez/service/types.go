package service

import (
	"context"
	"net/url"
)

// SearchExecutor 接口定义了每个 Entrez 操作应该实现的方法
type SearchExecutor interface {
	Execute(ctx context.Context, input interface{}) (interface{}, error)
	BuildURL() (*url.URL, error)
}
