package path

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath 转换路径中的分隔符，并处理家目录和当前目录的符号
func NormalizePath(path string) (string, error) {
	// 替换 ~ 为家目录
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// 替换 . 为当前目录
	if strings.HasPrefix(path, "./") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(cwd, path[1:])
	}

	if os.PathSeparator == '/' {
		// 在Unix-like系统（如Linux和macOS）中，将\转换为/
		return filepath.ToSlash(path), nil
	} else if os.PathSeparator == '\\' {
		// 在Windows系统中，将/转换为\
		return filepath.FromSlash(path), nil
	}

	return path, nil
}
