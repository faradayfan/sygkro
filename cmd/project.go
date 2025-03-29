package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Manage projects`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
