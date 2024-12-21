package command

import (
	"context"
	"fmt"
	"github.com/iEchoxu/clinvarDL/configs"
	"github.com/iEchoxu/clinvarDL/configs/defaults"
	"github.com/iEchoxu/clinvarDL/pkg/entrez"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/config"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/input"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/output/excel"
	"github.com/iEchoxu/clinvarDL/pkg/entrez/pkg/logcdl"
	"github.com/iEchoxu/clinvarDL/pkg/platform/path"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// 启动命令 clinvarDL run -f gene.txt
// 目前只支持使用 gene symbol 查询

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start ClinvarDL",
	Long:  `Start clinvarDL and run tasks`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		// 初始化日志
		err := logcdl.InitLogger(logcdl.Options{
			MinLevel:    logcdl.INFO,
			LogDir:      "logs",
			LogFileName: "clinvarDL_%s.log",
			TimeFormat:  "2006-01-02",
		})
		if err != nil {
			logcdl.Error("failed to init logger: %v", err)
			return
		}
		defer logcdl.Close()

		searchFile, _ = path.NormalizePath(searchFile)

		// 校验输入文件是否合规范
		if err := validateArgs(searchFile); err != nil {
			logcdl.Error("Error in query file path or format: %v", err)
			return
		}

		cf := configs.Config{
			Configs: []configs.ConfigFile{
				{Config: configs.NewEntrezSettingConfig(), FilePath: defaults.SettingsConfigPath()},
			},
		}

		settings, err := cf.LoadSettings()
		if err != nil {
			logcdl.Error("Error loading settings: %v", err)
			return
		}

		// 创建配置
		entrezConfig := config.NewConfig(settings.EntrezSetting.DB)

		// 设置配置
		entrezConfig.SetFilters(configs.NewFiltersConfigWithPath(defaults.FiltersConfigPath()).BuildQueryStringWithTerm()).
			SetRetMax(settings.EntrezSetting.RetMax).
			SetUseHistory(settings.EntrezSetting.UseHistory).
			SetRetMode(config.RetMode(settings.EntrezSetting.RetMode)).
			SetApiKey(settings.EntrezSetting.ApiKey).
			SetEmail(settings.EntrezSetting.Email).
			SetToolName(settings.EntrezSetting.ToolName).
			SetCacheEnabled(settings.CacheSetting.Enabled).
			SetCacheDir(settings.CacheSetting.Dir).
			SetCacheTTL(settings.CacheSetting.TTL).
			SetCacheMaxSize(settings.CacheSetting.MaxSize).
			SetOutputDir(settings.OutputSetting.Storage). // 设置输出目录
			SetQueryTimeout(settings.TimeoutSetting.QueryTimeout).
			SetSingleQueryTimeout(settings.TimeoutSetting.SingleQueryTimeout).
			SetWriteTimeout(settings.TimeoutSetting.WriteTimeout).
			SetStreamEnabled(true) // 启用流式处理

		// 统一验证配置
		if err := entrezConfig.Validate(); err != nil {
			logcdl.Error("invalid config: %v", err)
			return
		}

		// 读取并解析输入文件
		parser := input.NewFileParser(settings.EntrezSetting.BatchSize, settings.GetSearchFlag(settings.EntrezSetting.SearchType))
		queries, err := parser.ParseFile(searchFile)
		if err != nil {
			logcdl.Error("failed to parse input file '%s': %v", searchFile, err)
			return
		}

		// 设置上下文和超时
		ctx, cancel := context.WithTimeout(context.Background(), entrezConfig.Runtime.QueryTimeout)
		defer cancel()

		// 执行查询
		service := entrez.NewEntrezService(entrezConfig)
		results, err := service.ExecuteQueries(ctx, queries)

		// 如果所有查询都失败了，退出程序
		if err != nil {
			logcdl.Error("%v", err)
			return
		}

		// 创建 ExcelWriter
		resultWriter, err := excel.NewWriter("ClinVar Results")
		if err != nil {
			logcdl.Error("failed to create excel writer: %v", err)
		}
		defer resultWriter.Close()

		defer func() {
			logcdl.Tip("total time taken: %s", time.Since(start))
		}()

		// 获取完整输出路径
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		outputFile := fmt.Sprintf("clinvar_results_%s.xlsx", timestamp)
		outputPath := entrezConfig.Output.GetOutputPath(outputFile)

		// 处理结果
		err = service.ProcessResults(ctx, results, outputPath, resultWriter)
		if err != nil {
			logcdl.Error("failed to process results: %v", err)
			return
		}

		logcdl.Success("results have been saved to %s", outputPath)
	},
}

var searchFile string

func init() {
	runCmd.Flags().StringVarP(&searchFile, "file", "f", "", "./clinvarDL run  -f gene.txt")
	rootCmd.AddCommand(runCmd)
}

func validateArgs(arg string) error {
	if arg == "" {
		return fmt.Errorf("参数为空: 请使用 -f 指定搜索文件路径")
	}

	if filepath.Ext(arg) != ".txt" {
		return fmt.Errorf("参数错误: 搜索文件必须为 txt 格式")
	}

	if os.PathSeparator == '\\' && strings.Contains(arg, "/") {
		return fmt.Errorf("参数错误: %s 中包含非法字符 /", arg)
	}

	if os.PathSeparator == '/' && strings.Contains(arg, "\\") {
		return fmt.Errorf("参数错误: %s 中包含非法字符 \\", arg)
	}

	if _, err := os.Stat(arg); os.IsNotExist(err) {
		return fmt.Errorf("参数错误: 未找到 %s", arg)
	}

	return nil
}
