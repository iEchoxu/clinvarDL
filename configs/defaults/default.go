package defaults

import (
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/platform/path"
	"os"
	"path/filepath"
)

const (
	DefaultConfigDir       = ".clinvarDL"
	SettingsConfigFileName = "settings.yaml"
	FiltersConfigFileName  = "filters.yaml"
)

func BaseConfigDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		logcdl.Error("failed to get current directory: %v", err)
		return ""
	}

	return filepath.Join(pwd, DefaultConfigDir)
}

func CurrentDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		logcdl.Error("failed to get current directory: %v", err)
		return ""
	}

	return pwd
}

func SettingsConfigPath() string {
	return filepath.Join(BaseConfigDir(), SettingsConfigFileName)
}

func FiltersConfigPath() string {
	return filepath.Join(BaseConfigDir(), FiltersConfigFileName)
}

func Storage(dir, name string) string {
	urlDir := filepath.Join(CurrentDir(), dir)
	path.CheckDir(urlDir)
	return filepath.Join(urlDir, name)
}
