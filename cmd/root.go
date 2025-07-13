package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "nats-helper",
	Short: "A CLI tool for managing and interacting with NATS messaging system",
	Long:  `A CLI tool for managing and interacting with NATS messaging system.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Initialize configuration
	viper.SetEnvPrefix("NATS_HELPER")
	viper.AutomaticEnv()
}
