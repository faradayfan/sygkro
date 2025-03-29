package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	templateCmd.AddCommand(templateNewCmd)
}

var templateNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Generates a new project template",
	Long:  `Generates a new project template`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating a new project template...")
	},
}
