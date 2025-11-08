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
	Long: `Outputs a git diff style output between the project and its template.
	1. Reads the sygkro.sync.yaml file to get the template source and inputs.
	2. Clones the template repository at the specified reference to a temporary location.
	3. Generates the ideal state of the project based on the template and inputs.
	4. Computes the diff between the current project and the ideal state for files tracked by the template.
	5. Outputs the diff in a git diff style format.
	`,
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
