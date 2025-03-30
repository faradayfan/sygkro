package cmd

import (
	"fmt"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(projectSyncCmd)
}

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project to a template",
	Long:  `Syncs a project to a template`,
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
			return nil
		}

		err = git.ApplyDiff(".", diff)
		if err != nil {
			return fmt.Errorf("failed to apply diff: %w", err)
		}

		// update the sync config with the latest commit SHA
		syncConfig.Source.TemplateVersion = latest.CommitSHA
		err = syncConfig.Write(syncConfigPath)
		if err != nil {
			return fmt.Errorf("failed to write sync config: %w", err)
		}

		fmt.Println("Project synced successfully.")

		return nil
	},
}
