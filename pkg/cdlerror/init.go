package cdlerror

import (
	"fmt"
	"github.com/pkg/errors"
	"io/fs"
)

type Error struct {
	ErrorType ErrorType
	Cause     ConfigError
}

type CDLError struct {
	Errors []Error
}

func (c *CDLError) AddErrors() {
	for _, err := range c.Errors {
		addErrorToMapping(err.ErrorType, err.Cause)
	}
}
func addErrorToMapping(errType ErrorType, err ConfigError) {
	if _, exists := mapping[errType]; exists {
		fmt.Printf("Key '%T' exists\n", errType)
		return
	}
	mapping[errType] = err
}

func init() {
	ErrConfigFileCreationFailed := "无法创建配置文件"
	ErrEmailOrAPIKeyChangeFailed := "无法修改邮箱或API密钥"
	ErrConfigEditFailed := "无法编辑配置文件"

	cdlError := CDLError{
		Errors: []Error{
			{
				ErrorType: ErrTypeCreateConfig,
				Cause: NewCreateConfigError(ErrConfigFileCreationFailed).
					AddError(&fs.PathError{Err: errors.New("打开配置文件出错,配置文件不存在或没有权限读取")}).
					AddError(ErrConfigLoadMissing).
					AddError(ErrConfigSerializationFailed).
					AddError(ErrDirectoryCreationFailed).
					AddError(ErrConfigFileWriteFailed),
			},
			{
				ErrorType: ErrTypeSetEmailOrAPI,
				Cause: NewSetEmailOrAPIError(ErrEmailOrAPIKeyChangeFailed).
					AddError(&fs.PathError{Err: errors.New("打开配置文件出错,配置文件不存在或没有权限读取")}).
					AddError(ErrConfigParseFailed).
					AddError(ErrConfigLoadMissing).
					AddError(ErrConfigLoadOverflow).
					AddError(ErrConfigLoadNoValidConfigFound).
					AddError(ErrConfigAssertionFailed).
					AddError(ErrInvalidEmailAddress).
					AddError(ErrInvalidAPIKey).
					AddError(ErrUnknownKey),
			},
			{
				ErrorType: ErrTypeEditSettingConfig,
				Cause: NewEditSettingError(ErrConfigEditFailed).
					AddError(fs.ErrNotExist).
					AddError(&fs.PathError{Err: errors.New("启动进程失败,打开进程文件出错或进程文件不存在")}).
					AddError(ErrWaitingForProcessFailed),
			},
		},
	}

	cdlError.AddErrors()
}
