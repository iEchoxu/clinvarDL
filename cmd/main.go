package main

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/cmd/command"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("请输入命令，例如: clinvarDL run -f ***.txt")
		fmt.Println("")
		fmt.Println("clinvarDL 命令:")
		fmt.Println("  config: 配置文件初始化")
		fmt.Println("  config set email ***@***.com: 配置文件设置邮箱")
		fmt.Println("  config set apiKey ***: 配置文件设置apiKey")
		fmt.Println("  config edit: 配置文件编辑")
		fmt.Println("  filters edit: 过滤条件编辑")
		fmt.Println("  run -f ***.txt: 运行搜索")
		fmt.Println("  help: 查看帮助")
		return
	}

	command.Execute()
}
