package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vahidmostofi/wise-auto-scaler/internal/starter"
)

func init() {
	rootCmd.AddCommand(evaluateCmd)

}

var evaluateCmd = &cobra.Command{
	Use:     "evaluate",
	Aliases: []string{"eval"},
	Short:   "evaluate the deployment",
	Long:    "evaluate the deployment",
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetDefault("evaluate_interval", 12)
		viper.SetDefault("evaluate_report_path", "/home/vahid/Desktop/")

		starter.StartEvaluator()
	},
}
