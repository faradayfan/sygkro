package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sygkro",
	Short: "Sygkro is a project templating and synchronization tool",
	Long: `Sygkro is a project templating and synchronization tool that helps you manage your projects with ease.
	It allows you to create, update any git project.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
