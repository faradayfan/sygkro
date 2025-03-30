package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/engine"
	"github.com/faradayfan/sygkro/internal/git"
	"github.com/spf13/cobra"
)

var projectCreateCmd = &cobra.Command{
	Use:   "create --template [template-ref]",
	Short: "Generates a new project from a template directory or Git repo into a new project directory under the target directory",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get target directory from flag.
		targetDir, err := cmd.Flags().GetString("target")
		if err != nil {
			return err
		}

		// Get the template reference from the required --template flag.
		templateRef, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		gitRef, err := cmd.Flags().GetString("git-ref")
		if err != nil {
			return err
		}

		// Use the git package to get a local template directory and the HEAD commit SHA.
		// localTemplateDir, commitSHA, ref, cleanup, err := git.GetTemplateDir(templateRef, gitRef)
		templateResults, err := git.GetTemplateDir(templateRef, gitRef)
		if err != nil {
			return err
		}
		// Ensure cleanup of temporary directory if needed.
		defer templateResults.Cleanup()

		// Ensure the local template directory exists.
		if _, err := os.Stat(templateResults.Path); err != nil {
			return fmt.Errorf("template directory %s does not exist: %w", templateResults.Path, err)
		}

		// Validate that the template directory contains the required subdirectory "{{ .slug }}".
		expectedSubDir := filepath.Join(templateResults.Path, "{{ .slug }}")
		if stat, err := os.Stat(expectedSubDir); err != nil || !stat.IsDir() {
			return fmt.Errorf("template directory %s must contain a subdirectory named '{{ .slug }}'", templateResults.Path)
		}

		// Read the template configuration file.
		configFilePath := filepath.Join(templateResults.Path, config.TemplateConfigFileName)
		tmplConfig, err := config.ReadTemplateConfig(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read template config file: %w", err)
		}

		// Prompt the user for each input defined in the configuration.
		inputs := make(map[string]string)
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Please provide values for the following inputs:")
		for key, defaultVal := range tmplConfig.Templating.Inputs {
			fmt.Printf("%s (default: %s): ", key, defaultVal)
			userInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input for %s: %w", key, err)
			}
			userInput = strings.TrimSpace(userInput)
			if userInput == "" {
				userInput = defaultVal
			}
			inputs[key] = userInput
		}

		// Render the project directory name using the "{{ .slug }}" template.
		renderedProjectDir, err := engine.RenderString("{{ .slug }}", inputs)
		if err != nil {
			return fmt.Errorf("failed to render project directory name: %w", err)
		}

		// Build the full destination path.
		destination := filepath.Join(targetDir, renderedProjectDir)
		if _, err := os.Stat(destination); err == nil {
			return fmt.Errorf("destination directory %s already exists", destination)
		}
		if err := os.MkdirAll(destination, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}

		// Process the project template: render the "{{ .slug }}" subdirectory into the destination.
		if err := engine.ProcessTemplateDir(expectedSubDir, destination, inputs, tmplConfig.Options); err != nil {
			return fmt.Errorf("failed to process template subdirectory: %w", err)
		}

		// split string by "/" and return last element
		trackingRef := strings.Split(templateResults.HeadRef, "/")
		var trackingRefString string = ""
		if len(trackingRef) > 0 {
			trackingRefString = trackingRef[len(trackingRef)-1]
		}

		// Build the sync configuration.
		syncConfig := config.SyncConfig{
			Source: config.SourceConfig{
				TemplatePath:        templateRef, // Preserve the original reference.
				TemplateName:        tmplConfig.Name,
				TemplateVersion:     templateResults.CommitSHA,
				TemplateTrackingRef: trackingRefString,
			},
			Inputs: inputs,
		}
		syncConfigFilePath := filepath.Join(destination, config.SyncConfigFileName)
		if err := syncConfig.Write(syncConfigFilePath); err != nil {
			return fmt.Errorf("failed to write sync config file: %w", err)
		}

		fmt.Printf("Project created successfully in %s\n", destination)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCreateCmd.Flags().StringP("template", "s", "", "Path or Git repo reference to the template (required)")
	projectCreateCmd.Flags().StringP("target", "t", ".", "Target directory for the new project")
	projectCreateCmd.Flags().StringP("git-ref", "r", "", "Git reference (branch, tag, or commit SHA) to use for the template")
	projectCreateCmd.MarkFlagRequired("template")
}
