package configs

import (
	"github.com/iEchoxu/clinvarDL/pkg/cdlerror"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func write(configType Configer, configFile string) error {
	yamlData, err := yaml.Marshal(configType)
	if err != nil {
		return errors.Wrapf(cdlerror.ErrConfigSerializationFailed, "failed to serialize %T|%v", configType, err)
	}

	configBaseDir := filepath.Dir(configFile)
	err = os.MkdirAll(configBaseDir, 0744)
	if err != nil {
		return errors.Wrapf(cdlerror.ErrDirectoryCreationFailed, "failed to create config directory %s|%v", configBaseDir, err)
	}

	// 写入文件
	file, err := os.Create(configFile)
	if err != nil {
		return errors.Wrapf(err, "failed to create config file %s", configFile)
	}
	defer file.Close()

	_, err = file.Write(yamlData)
	if err != nil {
		return errors.Wrapf(cdlerror.ErrConfigFileWriteFailed, "failed to write config file %s|%v", configFile, err)
	}

	return nil
}

func read(configFile string, configType Configer) (Configer, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s|%s", configFile, err)
	}

	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(configType); err != nil {
		return nil, errors.Wrapf(cdlerror.ErrConfigParseFailed, "failed to parse %s|%s", configFile, err)
	}

	return configType, nil
}
