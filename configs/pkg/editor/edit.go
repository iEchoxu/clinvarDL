package editor

import (
	"github.com/iEchoxu/clinvarDL/pkg/cdlerror"
	"github.com/iEchoxu/clinvarDL/pkg/platform"
	"os"

	"github.com/pkg/errors"
)

func EditConfig(file string) error {
	// 检查文件是否存在
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errors.Wrapf(err, "file %s not found", file)
	}

	editorInfo := platform.DefaultEditor().Info()

	prog := editorInfo.Path
	args := []string{prog, file}

	// 环境变量
	env := os.Environ()

	// 使用 StartProcess 启动进程
	process, err := os.StartProcess(prog, args, &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env:   env,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to start process")
	}

	// 等待进程结束
	_, err = process.Wait()
	if err != nil {
		return errors.Wrapf(cdlerror.ErrWaitingForProcessFailed, "failed to wait for process|%v", err)
	}

	return nil
}
