package configs

import (
	"github.com/iEchoxu/clinvarDL/configs/settings"
)

type EntrezSettingConfig struct {
	EntrezSetting  *settings.EntrezSettings  `yaml:"entrez_setting"`
	OutputSetting  *settings.OutputSettings  `yaml:"output_setting"`
	CacheSetting   *settings.CacheSettings   `yaml:"cache_setting"`
	TimeoutSetting *settings.TimeoutSettings `yaml:"timeout_setting"`
}

type EntrezSettingConfigOption func(option *EntrezSettingConfig)

// WithInputProcessor 是一个配置选项，用于设置 useInputProcessor 标志
func WithInputProcessor(use bool) EntrezSettingConfigOption {
	return func(cfg *EntrezSettingConfig) {
		if use {
			cfg.EntrezSetting = settings.NewEntrezSettings().InputProcessor()
		} else {
			cfg.EntrezSetting = settings.NewEntrezSettings()
		}
	}
}

// NewEntrezSettingConfig 创建并返回一个EntrezSettingConfig配置
func NewEntrezSettingConfig(options ...EntrezSettingConfigOption) *EntrezSettingConfig {
	cfg := &EntrezSettingConfig{
		OutputSetting:  settings.NewOutputSettings(),
		CacheSetting:   settings.NewCacheSettings(),
		TimeoutSetting: settings.NewTimeoutSettings(),
	}

	// 应用默认配置，确保EntrezSetting总是被初始化
	if len(options) == 0 {
		options = append(options, WithInputProcessor(false))
	}

	for _, option := range options {
		option(cfg)
	}

	return cfg
}

func (esc *EntrezSettingConfig) Write(configFile string) error {
	return write(esc, configFile)
}

func (esc *EntrezSettingConfig) Read(configFile string) (Configer, error) {
	return read(configFile, esc)
}

/*
searchType 定义查询类型
目前只实现了 gene symbol 类型
详细：https://www.ncbi.nlm.nih.gov/clinvar/docs/help/
*/
func (esc *EntrezSettingConfig) searchType() map[string]string {
	return map[string]string{
		"gene symbol": "[gene]",
	}
}

// GetSearchFlag 根据查询类型返回对应的 clinvar 查询标志
func (esc *EntrezSettingConfig) GetSearchFlag(searchType string) string {
	return esc.searchType()[searchType]
}
