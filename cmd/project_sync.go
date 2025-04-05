package cmd

import (
	"fmt"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/spf13/cobra"
)

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a project to a template",
	Long:  `Syncs a project to a template`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Not yet implemented")
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
	projectSyncCmd.Flags().StringP("config", "c", config.SyncConfigFileName, "Path to the sync config file")
	projectSyncCmd.Flags().StringP("git-ref", "r", "", "Git reference to use (branch, tag, or commit SHA)")
}
