package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(projectCreateCmd)
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new project from a template",
	Long:  `Creates a new project from a template`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating a new project template...")
	},
}
