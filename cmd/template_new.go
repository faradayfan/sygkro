package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/templates"
	"github.com/spf13/cobra"
)

var templateNewCmd = &cobra.Command{
	Use:   "new [template-name]",
	Short: "Creates a new template directory with necessary files",
	Long: `Creates a new template directory with necessary files.
	1. Creates a new directory with the specified template name.
	2. Creates a sygkro.template.yaml file with default configuration.
	3. Creates a subdirectory named '{{ .slug }}' for template files.
	4. Adds a sample README.md file in the template files directory.
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		templateInputs := map[string]string{
			"name":        "my-project",
			"slug":        "my-project",
			"description": "A new project created by sygkro",
			"author":      "Your Name",
		}

		templateDir := filepath.Join(templateName)
		templateFilesDir := filepath.Join(templateDir, "{{ .slug }}")

		if _, err := os.Stat(templateDir); err == nil {
			return fmt.Errorf("directory %s already exists", templateDir)
		}

		if err := os.Mkdir(templateDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := os.Mkdir(templateFilesDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		configFilePath := filepath.Join(templateDir, config.TemplateConfigFileName)
		templateConfig := config.TemplateConfig{
			Name:        templateName,
			Description: "A new template created by sygkro",
			Templating: config.TemplatingConfig{
				Inputs: templateInputs,
			},
		}

		err := templateConfig.Write(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		readmeFilePath := filepath.Join(templateFilesDir, "README.md")
		readmeFileContent, err := templates.GetTemplate("README.md.tpl")
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		if err := os.WriteFile(readmeFilePath, readmeFileContent, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("Template %s created successfully.\n", templateName)
		return nil
	},
}

func init() {

	templateCmd.AddCommand(templateNewCmd)
}
