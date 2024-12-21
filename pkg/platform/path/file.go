package path

import (
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"os"
	"path/filepath"
	"time"
)

func CheckDir(dir string) {
	// 检查文件夹是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 如果文件夹不存在，则创建
		err := os.Mkdir(dir, 0755)
		if err != nil {
			logcdl.Error("failed to create directory: %v", err)
			return
		}
	}
}

func Backup(srcDir string) {
	baseName := filepath.Base(srcDir)
	parentDir := filepath.Dir(srcDir)
	dstDir := parentDir + string(os.PathSeparator) + "backup" + string(os.PathSeparator) + baseName + time.Now().Format("2006-01-02_15-04-05")

	// 遍历源文件夹
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logcdl.Error("error accessing path %q: %v\n", path, err)
			return err
		}
		// 检查文件是否是.txt文件且不是目录
		// if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt")
		if !info.IsDir() {
			// 构造目标路径
			relPath, err := filepath.Rel(srcDir, path)
			if err != nil {
				logcdl.Error("error getting relative path for %q: %v", path, err)
				return err
			}
			dstPath := filepath.Join(dstDir, relPath)

			// 创建目标文件的父目录
			err = os.MkdirAll(dstDir, 0755)
			if err != nil {
				logcdl.Error("error creating directory for %q: %v", dstPath, err)
				return err
			}

			// 移动文件
			logcdl.Info("moving %s to %s", path, dstPath)
			err = os.Rename(path, dstPath)
			if err != nil {
				logcdl.Error("error moving editor %q to %q: %v", path, dstPath, err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		logcdl.Error("error walking the path %q: %v", srcDir, err)
	}
}
