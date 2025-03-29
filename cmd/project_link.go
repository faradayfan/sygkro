package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(projectLinkCmd)
}

var projectLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Links an existing project to a template",
	Long:  `Links an existing project to a template`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating a new project template...")
	},
}
