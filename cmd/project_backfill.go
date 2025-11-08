package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(projectBackfillCmd)
}

var projectBackfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Backfills a template with changes in the project",
	Long: `Backfills a template with changes in the project.
		1. Generates a temporary ideal state of the project based on the template inputs found in the sygkro.sync.yaml
		2. Computes the diff between the current project and the ideal state for files tracked by the template.
		3. Applies the diff to the template project
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not yet implemented")
	},
}
