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
		syncFilePath := cmd.Flag("config").Value.String()
		syncConfig, err := config.ReadSyncConfig(syncFilePath)
		if err != nil {
			return err
		}

		gitRef := cmd.Flag("git-ref").Value.String()
		if gitRef != "" {
			syncConfig.Source.TemplateTrackingRef = gitRef
		}

		templateDir, err := git.GetTemplateDir(syncConfig.Source.TemplatePath, syncConfig.Source.TemplateTrackingRef)
		if err != nil {
			return fmt.Errorf("failed to clone template repository: %w", err)
		}
		defer templateDir.Cleanup()

		diff, err := git.ComputeDiff(templateDir.Path, ".", syncConfig.Source.TemplateVersion, syncConfig)
		if err != nil {
			return fmt.Errorf("failed to compute diff: %w", err)
		}
		if diff == "" {
			fmt.Println("No differences found.")
			return nil
		}
		fmt.Println(diff)

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectDiffCmd)
	projectDiffCmd.Flags().StringP("config", "c", config.SyncConfigFileName, "Path to the sync config file")
	projectDiffCmd.Flags().StringP("git-ref", "r", "", "Git reference to use (branch, tag, or commit SHA)")
}
