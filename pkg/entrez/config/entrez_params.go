package config

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	customerrors "github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/retry/errors"
)

// EntrezParams 包含所有 Entrez API 调用相关的基本参数
type EntrezParams struct {
	DB         string
	Filters    string
	RetMax     int
	UseHistory bool
	RetMode    RetMode
	ApiKey     string
	Email      string
	ToolName   string
}

// validateRequired 验证必需参数
func (e *EntrezParams) validateRequired() error {
	// 数据库不能为空
	if e.DB == "" {
		return customerrors.NewParametersError("database is required")
	}

	// RetMode 必须为 xml 或 json
	if !e.RetMode.IsValid() {
		return customerrors.NewParametersError(fmt.Sprintf("retmode %s is not supported, use xml or json", e.RetMode))
	}

	// RetMax 不能超过 10000
	if e.RetMax > 10000 {
		return customerrors.NewParametersError(fmt.Sprintf("retmax %d exceeds maximum allowed value (10000)", e.RetMax))
	}

	return nil
}

// validateOptional 验证可选参数并打印警告
func (e *EntrezParams) validateOptional() error {
	// 检查过滤器
	if e.Filters == "" {
		logcdl.Warn("no filters set: consider adding filters to reduce data volume")
	}

	// 检查历史记录功能
	if !e.UseHistory {
		logcdl.Warn("usehistory is disabled: this may result in slower processing")
	}

	// 检查 API Key
	if e.ApiKey == "" {
		logcdl.Warn("no api key provided: requests will be limited to 3 per second")
	}

	// 检查邮箱
	if e.Email == "" {
		logcdl.Warn("no email provided: please set email for better support")
	}

	// 检查工具名称
	if e.ToolName == "" {
		logcdl.Warn("no tool name provided: please set tool name for better support")
	}

	return nil
}
