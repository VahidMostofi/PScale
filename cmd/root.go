package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "pscale",
		Short: "pscale",
		Long:  "pscale",
	}
)

// Execute ...
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	fmt.Println("root init")
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() {
	viper.Set("Verbose", true)
	viper.SetEnvPrefix("WA")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)

	} else {
		//TODO error
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using Config File:", viper.ConfigFileUsed())
	}
}
