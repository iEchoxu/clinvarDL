package command

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/configs/defaults"
	"github.com/iEchoxu/clinvarDL/configs/pkg/editor"
	"github.com/iEchoxu/clinvarDL/pkg/cdlerror"
	"github.com/spf13/cobra"
	"os"
)

// filtersCmd 过滤条件的命令
var filtersCmd = &cobra.Command{
	Use:       "filters",
	Short:     "filters editor for clinvarDL",
	Long:      `filters editor for clinvarDL`,
	Args:      cobra.MatchAll(cobra.OnlyValidArgs),
	ValidArgs: []string{"edit"},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

// editFiltersCmd 作为 filters 的子命令,用于编辑 filters 文件
// Run: ./cdl filters edit
var editFiltersCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit filters",
	Long:  `edit config file for filters`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := editor.EditConfig(defaults.FiltersConfigPath())
		if err != nil {
			if handled := cdlerror.CheckError(cdlerror.ErrTypeEditSettingConfig, err); !handled {
				fmt.Println("无法打开配置文件: 遇到未知错误。请咨询软件开发人员以获取进一步帮助")
				//fmt.Printf("%+v\n", err) // 可打印错误堆栈信息
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(filtersCmd)
	filtersCmd.AddCommand(editFiltersCmd)
}
