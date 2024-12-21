package command

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:       "clinvarDL",
	Short:     "Download clinvar data",
	Long:      `Download clinvar data and generate excel`,
	Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.MinimumNArgs(1)),
	ValidArgs: []string{"configs", "filters", "run"},
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
