package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(projectLinkCmd)
	projectLinkCmd.Flags().StringP("template", "s", "", "Path or Git repo reference to the template (required)")
	projectLinkCmd.Flags().StringP("target", "t", ".", "Target directory for the project to be linked to the template")
	projectLinkCmd.Flags().StringP("git-ref", "r", "", "Git reference (branch, tag, or commit SHA) to use for the template")
	projectLinkCmd.Flags().BoolP("quiet", "q", false, "Accepts default values for all inputs without prompting the user")
	projectLinkCmd.MarkFlagRequired("template")
}

var projectLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Links an existing project to a template",
	Long: `Links an existing project to a template
		1. Confirms there is no existing sygkro.sync.yaml file in the project.
		2. Clones the template repository provided as an input to a temporary location.
		3. Confirms the revision exists in the template repository.
		3. Prompts the user for input values defined in the template.
		4. Writes a sygkro.sync.yaml file to track the template source and inputs used.
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, err := cmd.Flags().GetString("target")
		if err != nil {
			return err
		}

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			return fmt.Errorf("target directory %s does not exist", targetDir)
		}

		templateRef, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		gitRef, err := cmd.Flags().GetString("git-ref")
		if err != nil {
			return err
		}

		templateResults, err := git.GetTemplateDir(templateRef, gitRef)
		if err != nil {
			return err
		}

		defer templateResults.Cleanup()

		if _, err := os.Stat(templateResults.Path); err != nil {
			return fmt.Errorf("template directory %s does not exist: %w", templateResults.Path, err)
		}

		expectedSubDir := filepath.Join(templateResults.Path, "{{ .slug }}")
		if stat, err := os.Stat(expectedSubDir); err != nil || !stat.IsDir() {
			return fmt.Errorf("template directory %s must contain a subdirectory named '{{ .slug }}'", templateResults.Path)
		}

		configFilePath := filepath.Join(templateResults.Path, config.TemplateConfigFileName)
		tmplConfig, err := config.ReadTemplateConfig(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read template config file: %w", err)
		}

		inputs := make(map[string]string)
		quietMode, err := cmd.Flags().GetBool("quiet")
		if err != nil {
			return err
		}

		if quietMode {
			for key, defaultVal := range tmplConfig.Templating.Inputs {
				inputs[key] = defaultVal
			}
		} else {
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
		}

		trackingRef := strings.Split(templateResults.HeadRef, "/")
		var trackingRefString string = ""
		if len(trackingRef) > 0 {
			trackingRefString = trackingRef[len(trackingRef)-1]
		}

		syncConfig := config.SyncConfig{
			Source: config.SourceConfig{
				TemplatePath:        templateRef,
				TemplateName:        tmplConfig.Name,
				TemplateVersion:     templateResults.CommitSHA,
				TemplateTrackingRef: trackingRefString,
			},
			Inputs: inputs,
		}
		syncConfigFilePath := filepath.Join(targetDir, config.SyncConfigFileName)
		if err := syncConfig.Write(syncConfigFilePath); err != nil {
			return fmt.Errorf("failed to write sync config file: %w", err)
		}

		fmt.Printf("Project linked to template %s successfully!\nRun 'sygkro project sync' to synchronize the project with the template.", templateRef)
		return nil
	},
}
