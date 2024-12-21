package configs

import (
	"errors"
	"fmt"
	"github.com/iEchoxu/clinvarDL/pkg/cdlerror"
	"sync"

	pkgErrors "github.com/pkg/errors"
)

// ConfigFile 结构体包含配置器和文件名
type ConfigFile struct {
	Config   Configer
	FilePath string
}

// Config 配置文件的集合
type Config struct {
	Configs []ConfigFile
}

// Create 创建一个或多个配置文件
func (c *Config) Create() error {
	if len(c.Configs) == 0 {
		return cdlerror.ErrConfigLoadMissing
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(c.Configs))

	wg.Add(len(c.Configs))
	for _, configFile := range c.Configs {
		go func(cf ConfigFile) {
			defer wg.Done()
			if err := cf.Config.Write(cf.FilePath); err != nil {
				errChan <- pkgErrors.WithMessage(err, "failed to write config file")
			}
		}(configFile)

	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// 使用 errors.Join 来组合多个错误并返回, 要求 go1.20 以上版本
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// LoadSettings 加载EntrezSettingConfig
func (c *Config) LoadSettings() (*EntrezSettingConfig, error) {
	if len(c.Configs) == 0 {
		return nil, cdlerror.ErrConfigLoadMissing
	}

	var cfg *EntrezSettingConfig
	for _, configFile := range c.Configs {
		cf, err := configFile.Config.Read(configFile.FilePath)
		if err != nil {
			return nil, err
		}

		// 断言cfg为*EntrezSettingConfig类型
		entrezCfg, ok := cf.(*EntrezSettingConfig)
		if !ok {
			return nil, cdlerror.ErrConfigAssertionFailed
		}

		cfg = entrezCfg
	}

	if cfg == nil {
		return nil, cdlerror.ErrConfigLoadNoValidConfigFound
	}

	return cfg, nil
}

// LoadFilters 加载FiltersConfig
func (c *Config) LoadFilters() (*FiltersConfig, error) {
	if len(c.Configs) == 0 {
		return nil, fmt.Errorf("missing config file: please ensure at least one config file is provided")
	}

	var cfg *FiltersConfig
	for _, configFile := range c.Configs {
		cf, err := configFile.Config.Read(configFile.FilePath)
		if err != nil {
			return nil, err
		}

		// 断言cfg为*FiltersConfig
		filtersCfg, ok := cf.(*FiltersConfig)
		if !ok {
			return nil, fmt.Errorf("type assertion failed: unable to assert config to *FiltersConfig type")
		}

		cfg = filtersCfg
		cfg.Path = configFile.FilePath // 设置默认配置文件路径,不设置此处会导致 BuildQueryStringWithTerm() 为空值
	}

	if cfg == nil {
		return nil, fmt.Errorf("failed to load config: no valid config found")
	}

	return cfg, nil
}

// LoadAll 加载所有配置
// 调用此方法时需要通过 cf.(type) 进行类型断言
func (c *Config) LoadAll() ([]Configer, error) {
	if len(c.Configs) == 0 {
		return nil, fmt.Errorf("missing config file: please ensure at least one config file is provided")
	}

	var configs []Configer
	for _, configFile := range c.Configs {
		cf, err := configFile.Config.Read(configFile.FilePath)
		if err != nil {
			return nil, err
		}

		configs = append(configs, cf)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("failed to load config: no valid config found")
	}

	return configs, nil
}

// SetEmailAndAPIKey 设置邮箱和API密钥
func (c *Config) SetEmailAndAPIKey(k, v string) error {
	if len(c.Configs) != 1 {
		return cdlerror.ErrConfigLoadOverflow
	}

	cfg, err := c.LoadSettings()
	if err != nil {
		return pkgErrors.WithMessage(err, "failed to load Settings config file")
	}

	if err = cfg.EntrezSetting.SetEmailOrAPIKey(k, v); err != nil {
		return pkgErrors.WithMessage(err, "failed to set Email or API Key")
	}

	return cfg.Write(c.Configs[0].FilePath)
}
