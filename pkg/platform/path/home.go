package path

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath 转换路径中的分隔符，并处理家目录和当前目录的符号
func NormalizePath(path string) (string, error) {
	// 处理 ~ 开头的路径
	if len(path) == 1 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		return homeDir, nil
	}

	if len(path) == 1 && path[0] == '.' {
		// 处理 . 开头的路径
		cwd, err := os.Getwd()
		if err != nil {
			return path, err
		}
		return cwd, nil
	}

	// 替换 ~、~/ 为家目录
	if len(path) > 1 && strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		join := filepath.Join(homeDir, path[2:])
		return filepath.Clean(join), nil
	}

	// 替换 ./ 为当前目录,不包括 .git 这样的隐藏目录
	if len(path) > 1 && strings.HasPrefix(path, "./") {
		cwd, err := os.Getwd()
		if err != nil {
			return path, err
		}
		join := filepath.Join(cwd, path[2:])
		return filepath.Clean(join), nil
	}

	//if os.PathSeparator == '/' {
	//	// 在Unix-like系统（如Linux和macOS）中且 path 中带有 \, 会将\转换为/
	//	return filepath.ToSlash(path), nil
	//} else if os.PathSeparator == '\\' {
	//	// 在Windows系统中且 path 中带有 /, 会将/转换为\
	//	return filepath.FromSlash(path), nil
	//}

	return filepath.Clean(path), nil
}
