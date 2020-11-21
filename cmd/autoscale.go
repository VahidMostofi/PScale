package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vahidmostofi/wise-auto-scaler/internal/starter"
)

func init() {
	rootCmd.AddCommand(autoscale)

}

var autoscale = &cobra.Command{
	Use:     "autoscale",
	Aliases: []string{"as"},
	Short:   "autoscale the deployment",
	Long:    "autoscale the deployment",
	Run: func(cmd *cobra.Command, args []string) {

		viper.SetDefault("autoscale_interval", 12)
		viper.SetDefault("evaluate_enable", true)
		starter.StartAutoscaler()
	},
}
