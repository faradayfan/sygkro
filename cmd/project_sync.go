package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(projectSyncCmd)
}

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project to a template",
	Long:  `Syncs a project to a template`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating a new project template...")
	},
}
