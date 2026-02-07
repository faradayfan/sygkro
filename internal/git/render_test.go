package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faradayfan/sygkro/internal/config"
)

func TestRenderTemplateAtPath_Basic(t *testing.T) {
	templateDir := t.TempDir()
	inputs := map[string]string{
		"name": "my-project",
		"slug": "my-project",
	}

	// Create template config
	cfg := config.TemplateConfig{
		Name:        "test-template",
		Description: "A test template",
		Templating: config.TemplatingConfig{
			Inputs: inputs,
		},
	}
	if err := cfg.Write(filepath.Join(templateDir, config.TemplateConfigFileName)); err != nil {
		t.Fatalf("failed to write template config: %v", err)
	}

	// Create slug directory with a template file
	slugDir := filepath.Join(templateDir, "{{ .slug }}")
	if err := os.MkdirAll(slugDir, 0755); err != nil {
		t.Fatalf("failed to create slug dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(slugDir, "README.md"), []byte("# {{ .name }}\n"), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Render
	targetDir := t.TempDir()
	if err := RenderTemplateAtPath(templateDir, targetDir, inputs); err != nil {
		t.Fatalf("RenderTemplateAtPath failed: %v", err)
	}

	// Verify rendered file
	content, err := os.ReadFile(filepath.Join(targetDir, "README.md"))
	if err != nil {
		t.Fatalf("failed to read rendered file: %v", err)
	}
	expected := "# my-project\n"
	if string(content) != expected {
		t.Errorf("rendered content = %q, want %q", string(content), expected)
	}
}

func TestRenderTemplateAtPath_MissingConfig(t *testing.T) {
	templateDir := t.TempDir()
	targetDir := t.TempDir()
	inputs := map[string]string{"name": "test"}

	err := RenderTemplateAtPath(templateDir, targetDir, inputs)
	if err == nil {
		t.Error("expected error for missing template config")
	}
}

func TestRenderTemplateAtPath_MissingSlugDir(t *testing.T) {
	templateDir := t.TempDir()
	targetDir := t.TempDir()
	inputs := map[string]string{"name": "test", "slug": "test"}

	// Create config but no slug directory
	cfg := config.TemplateConfig{
		Name: "test",
		Templating: config.TemplatingConfig{
			Inputs: inputs,
		},
	}
	if err := cfg.Write(filepath.Join(templateDir, config.TemplateConfigFileName)); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	err := RenderTemplateAtPath(templateDir, targetDir, inputs)
	if err == nil {
		t.Error("expected error for missing slug directory")
	}
}
