package platform

import (
	"github.com/iEchoxu/clinvarDL/pkg/platform/editor"
	"os"
	"runtime"
)

func Get() string {
	sysType := runtime.GOOS
	return sysType
}

func IsRoot() (isRoot bool) {
	isRoot = true
	goos := Get()
	if goos != "windows" && os.Geteuid() != 0 {
		isRoot = false
	}
	return isRoot
}

func DefaultEditor() editor.Editorer {
	switch runtime.GOOS {
	case "linux":
		return editor.NewLinuxEditor()
	case "darwin":
		return editor.NewDarwinEditor()
	case "windows":
		return editor.NewWindowsEditor()
	default:
		panic("unsupported operating system")
	}
}
