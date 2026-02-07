package cmd

import (
	"fmt"
	"os"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/git"
	"github.com/spf13/cobra"
)

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project to a template",
	Long: `Syncs a project to a template using 3-way merge.
		1. Reads the sygkro.sync.yaml file to get the template source and inputs.
		2. Clones the template repository with full history.
		3. Renders the template at both the old and new versions.
		4. Performs a 3-way merge for each file (base=old template, ours=project, theirs=new template).
		5. Clean merges update project files. Conflicts create .sygkro-conflict files.
		6. Updates the sygkro.sync.yaml file with the new template version.
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

		oldVersion := syncConfig.Source.TemplateVersion

		// Clone with full history so we can access both old and new commits
		templateDir, err := git.GetTemplateDirForSync(syncConfig.Source.TemplatePath, syncConfig.Source.TemplateTrackingRef)
		if err != nil {
			return fmt.Errorf("failed to clone template repository: %w", err)
		}
		defer templateDir.Cleanup()

		// Render the NEW template (at HEAD)
		theirsTmpDir, err := os.MkdirTemp("", "sygkro-theirs-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(theirsTmpDir)

		if err := git.RenderTemplateAtPath(templateDir.Path, theirsTmpDir, syncConfig.Inputs); err != nil {
			return fmt.Errorf("failed to render new template: %w", err)
		}

		// Render the OLD template (at previously synced version)
		baseTmpDir, err := os.MkdirTemp("", "sygkro-base-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(baseTmpDir)

		if oldVersion != "" {
			if err := git.GitCheckout(templateDir.Path, oldVersion); err != nil {
				return fmt.Errorf("failed to checkout old template version %s: %w", oldVersion, err)
			}
			if err := git.RenderTemplateAtPath(templateDir.Path, baseTmpDir, syncConfig.Inputs); err != nil {
				return fmt.Errorf("failed to render old template: %w", err)
			}
		}

		// 3-way merge: base (old template) vs ours (project) vs theirs (new template)
		mergeResult, err := git.ThreeWayMerge(baseTmpDir, ".", theirsTmpDir)
		if err != nil {
			return fmt.Errorf("failed to merge: %w", err)
		}

		// Check if there are any changes
		if len(mergeResult.Files) == 0 {
			fmt.Println("No differences found.")
			return nil
		}

		// Apply merge results
		if err := git.ApplyMerge(".", baseTmpDir, theirsTmpDir, mergeResult); err != nil {
			return fmt.Errorf("failed to apply merge: %w", err)
		}

		// Print summary
		for _, f := range mergeResult.Files {
			switch f.Status {
			case git.MergeClean:
				fmt.Printf("  updated: %s\n", f.RelPath)
			case git.MergeConflict:
				fmt.Printf("  conflict: %s (see %s)\n", f.RelPath, f.ConflictPath)
			case git.MergeNewFile:
				fmt.Printf("  added: %s\n", f.RelPath)
			case git.MergeDeletedFile:
				fmt.Printf("  deleted in template (kept): %s\n", f.RelPath)
			}
		}

		// Update sync config
		syncConfig.Source.TemplateVersion = templateDir.CommitSHA
		if err := syncConfig.Write(syncFilePath); err != nil {
			return fmt.Errorf("failed to write sync config: %w", err)
		}

		if mergeResult.HasConflict {
			fmt.Println("Sync completed with conflicts. Review .sygkro-conflict files.")
		} else {
			fmt.Println("Sync completed successfully.")
		}

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
	projectSyncCmd.Flags().StringP("config", "c", config.SyncConfigFileName, "Path to the sync config file")
	projectSyncCmd.Flags().StringP("git-ref", "r", "", "Git reference to use (branch, tag, or commit SHA)")
}
