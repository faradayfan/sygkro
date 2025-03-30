// cmd/project_diff.go
package cmd

import (
	"fmt"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/git"
	"github.com/spf13/cobra"
)

var projectDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Outputs a git diff style output between the project and its template",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Step 1: Ensure .sygkro.sync.yaml exists in the current directory.
		syncConfigPath := config.SyncConfigFileName
		syncConfig, err := config.ReadSyncConfig(syncConfigPath)
		if err != nil {
			return err
		}
		// Step 2: Clone the template repository at the latest commit for the tracking ref.
		latest, err := git.GetTemplateDir(syncConfig.Source.TemplatePath, syncConfig.Source.TemplateTrackingRef)
		if err != nil {
			return fmt.Errorf("failed to clone template repository: %w", err)
		}

		defer latest.Cleanup()

		// Step 3: Compute a diff between the rendered template and the actual project.
		diff, err := git.ComputeDiff(latest.Path, ".", syncConfig)
		if err != nil {
			return fmt.Errorf("failed to compute diff: %w", err)
		}
		if diff == "" {
			fmt.Println("No differences found.")
		}

		fmt.Println(diff)

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectDiffCmd)
}
