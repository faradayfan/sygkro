package git

import (
	"fmt"
	"path/filepath"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/engine"
)

// RenderTemplateAtPath renders a template directory into a target directory
// using the given inputs. It reads the template config from templateDir,
// finds the "{{ .slug }}" subdirectory, and processes it into targetDir.
func RenderTemplateAtPath(templateDir string, targetDir string, inputs map[string]string) error {
	templateConfig, err := config.ReadTemplateConfig(filepath.Join(templateDir, config.TemplateConfigFileName))
	if err != nil {
		return fmt.Errorf("failed to read template config: %w", err)
	}

	slugDir := filepath.Join(templateDir, "{{ .slug }}")

	if err := engine.ProcessTemplateDir(slugDir, targetDir, inputs, templateConfig.Options); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return nil
}
