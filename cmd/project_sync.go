package cmd

import (
	"github.com/faradayfan/sygkro/internal/config"
	"github.com/spf13/cobra"
)

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project to a template",
	Long:  `Syncs a project to a template`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// syncFilePath := cmd.Flag("config").Value.String()

		// // Step 1: Ensure .sygkro.sync.yaml exists in the current directory.
		// syncConfig, err := config.ReadSyncConfig(syncFilePath)
		// if err != nil {
		// 	return err
		// }

		// // git-ref
		// ref := cmd.Flag("git-ref").Value.String()
		// if ref != "" {
		// 	syncConfig.Source.TemplateTrackingRef = ref
		// }

		// // Step 2: Clone the template repository at the latest commit for the tracking ref.
		// latest, err := git.GetTemplateDir(syncConfig.Source.TemplatePath, syncConfig.Source.TemplateTrackingRef)
		// if err != nil {
		// 	return fmt.Errorf("failed to clone template repository: %w", err)
		// }

		// defer latest.Cleanup()

		// // Step 3: Compute a diff between the rendered template and the actual project.
		// diff, err := git.ComputeDiff(latest.Path, ".", syncConfig)
		// if err != nil {
		// 	return fmt.Errorf("failed to compute diff: %w", err)
		// }
		// if diff == "" {
		// 	fmt.Println("No differences found.")
		// 	return nil
		// }

		// err = git.ApplyDiff(".", diff)
		// if err != nil {
		// 	return fmt.Errorf("failed to apply diff: %w", err)
		// }

		// // update the sync config with the latest commit SHA
		// syncConfig.Source.TemplateVersion = latest.CommitSHA
		// err = syncConfig.Write(syncFilePath)
		// if err != nil {
		// 	return fmt.Errorf("failed to write sync config: %w", err)
		// }

		// fmt.Println("Project synced successfully.")

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
	projectSyncCmd.Flags().StringP("config", "c", config.SyncConfigFileName, "Path to the sync config file")
	projectSyncCmd.Flags().StringP("git-ref", "r", "", "Git reference to use (branch, tag, or commit SHA)")
}
