package cdlerror

import "github.com/pkg/errors"

const (
	Start ErrorType = iota + 1
	ErrTypeCreateConfig
	ErrTypeSetEmailOrAPI
	ErrTypeEditSettingConfig
	End
)

var (
	ErrConfigLoadMissing            = errors.New("配置文件缺失,请确保至少传入一个配置文件")
	ErrConfigOpenFailed             = errors.New("打开文件失败,文件不存在或没有权限读取")
	ErrConfigAssertionFailed        = errors.New("配置文件读取失败,无法断言为 Configer 类型")
	ErrConfigLoadNoValidConfigFound = errors.New("配置文件加载失败,未找到有效的配置")
	ErrConfigLoadOverflow           = errors.New("配置文件超量,仅允许传入一个配置文件")
	ErrInvalidEmailAddress          = errors.New("无效的邮箱地址")
	ErrInvalidAPIKey                = errors.New("无效的 API Key")
	ErrUnknownKey                   = errors.New("未知的配置键")
	ErrConfigSerializationFailed    = errors.New("配置文件序列化失败,请确保配置文件格式正确")
	ErrConfigParseFailed            = errors.New("配置文件解析失败,请确保配置文件格式正确")
	ErrDirectoryCreationFailed      = errors.New("创建目录失败,请确保目录存在且可写入")
	ErrConfigFileWriteFailed        = errors.New("配置文件写入失败,请确保配置文件存在且可写入")
	ErrWaitingForProcessFailed      = errors.New("等待进程失败,请确保进程文件存在且可读取")
)
