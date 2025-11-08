package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/engine"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestComputeDiff_ProjectCreateFlow(t *testing.T) {
	// Create a template directory and config
	templateDir := t.TempDir()
	templateName := "test-template"
	templateInputs := map[string]string{
		"name":        "my-project",
		"slug":        "my-project",
		"description": "A new project created by sygkro",
		"author":      "Your Name",
	}
	templateFilesDir := filepath.Join(templateDir, "{{ .slug }}")
	if err := os.MkdirAll(templateFilesDir, 0755); err != nil {
		t.Fatalf("failed to create template directories: %v", err)
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
		t.Fatalf("failed to marshal config: %v", err)
	}
	// Create a README.md in the template
	tmplReadme := filepath.Join(templateFilesDir, "README.md")
	if err := os.WriteFile(tmplReadme, []byte("Hello, World!\n"), 0644); err != nil {
		t.Fatalf("failed to write template README.md: %v", err)
	}

	// Generate a project from the template (simulate project create command)
	projectDir := t.TempDir()
	renderedProjectDir, err := engine.RenderString("{{ .slug }}", templateInputs)
	if err != nil {
		t.Fatalf("failed to render project dir name: %v", err)
	}
	destination := filepath.Join(projectDir, renderedProjectDir)
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatalf("failed to create destination dir: %v", err)
	}
	if err := engine.ProcessTemplateDir(templateFilesDir, destination, templateInputs, templateConfig.Options); err != nil {
		t.Fatalf("failed to process template dir: %v", err)
	}

	// Introduce a change in the template
	if err := os.WriteFile(tmplReadme, []byte("Modified content\n"), 0644); err != nil {
		t.Fatalf("failed to modify template README.md: %v", err)
	}

	// Compute the diff
	diffOutput, err := ComputeDiff(templateDir, destination, "idealRevision", &config.SyncConfig{Inputs: templateInputs})
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}
	// Check that the diff output contains the modification
	// Assert diff contains exactly upstream-template-old/README.md\n+++ upstream-template-new/README.md\n@@ -1 +1 @@\n-Hello, World!\n+Modified content\n
	if !contains(diffOutput, "upstream-template-old/README.md\n+++ upstream-template-new/README.md\n@@ -1 +1 @@\n-Hello, World!\n+Modified content\n") {
		t.Errorf("diff output does not contain expected modification")
	}
}
