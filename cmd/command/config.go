package command

import (
	"fmt"
	"github.com/iEchoxu/clinvarDL/configs"
	"github.com/iEchoxu/clinvarDL/configs/defaults"
	"github.com/iEchoxu/clinvarDL/configs/pkg/editor"
	"github.com/iEchoxu/clinvarDL/pkg/cdlerror"
	"os"

	"github.com/spf13/cobra"
)

// cfgCmd 生成配置文件
// Run: ./cdl config
var cfgCmd = &cobra.Command{
	Use:       "config",
	Short:     "config for clinvarDL",
	Long:      `config for clinvarDL`,
	Args:      cobra.MatchAll(cobra.OnlyValidArgs),
	ValidArgs: []string{"edit", "set"},
	Run: func(cmd *cobra.Command, args []string) {
		cf := configs.Config{
			Configs: []configs.ConfigFile{
				{
					Config:   configs.NewEntrezSettingConfig(configs.WithInputProcessor(true)),
					FilePath: defaults.SettingsConfigPath(),
				},
				{Config: configs.NewFiltersConfig(), FilePath: defaults.FiltersConfigPath()},
			},
		}

		if err := cf.Create(); err != nil {
			if handled := cdlerror.CheckError(cdlerror.ErrTypeCreateConfig, err); !handled {
				fmt.Println("无法修改邮箱或API密钥: 遇到未知错误。请咨询软件开发人员以获取进一步帮助")
				//fmt.Printf("%+v\n", err) // 可打印错误堆栈信息
				os.Exit(1)
			}
		}

		for _, c := range cf.Configs {
			fmt.Println("配置文件已写入:", c.FilePath)
		}
	},
}

// setCmd 作为 config 的子命令,用于修改 email 和 apiKey
// Run: ./cdl config set email/apiKey
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "set email and apiKey",
	Long:  `set email and apiKey for the clinvarDL`,
	Args:  cobra.MatchAll(cobra.RangeArgs(1, 2)),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := configs.Config{
			Configs: []configs.ConfigFile{
				{
					Config:   configs.NewEntrezSettingConfig(),
					FilePath: defaults.SettingsConfigPath(),
				},
			},
		}

		if err := cfg.SetEmailAndAPIKey(args[0], args[1]); err != nil {
			if handled := cdlerror.CheckError(cdlerror.ErrTypeSetEmailOrAPI, err); !handled {
				fmt.Println("无法修改邮箱或API密钥: 遇到未知错误。请咨询软件开发人员以获取进一步帮助")
				//fmt.Printf("%+v\n", err) // 可打印错误堆栈信息
				os.Exit(1)
			}

		}

		switch args[0] {
		case "email":
			fmt.Println("邮箱已修改")
		case "apiKey":
			fmt.Println("API密钥已修改")
		default:
			fmt.Println("未知的配置键")
		}

	},
}

// editCmd 作为 config 的子命令
// Run: ./cdl config edit
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit config",
	Long:  `edit the config for settings`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := editor.EditConfig(defaults.SettingsConfigPath())
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
	rootCmd.AddCommand(cfgCmd)
	cfgCmd.AddCommand(setCmd)
	cfgCmd.AddCommand(editCmd)
}
