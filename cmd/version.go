package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "0.3.0-alpha.2" // x-release-please-version
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Pringt the version number of Sygkro",
	Long:  `Print the version number of Sygkro`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Sygkro version: %s\n", version)
	},
}
