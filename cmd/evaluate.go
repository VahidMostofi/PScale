package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		fmt.Println("hi :D")
	},
}
