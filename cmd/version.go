package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of nats-helper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nats-helper version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
} 