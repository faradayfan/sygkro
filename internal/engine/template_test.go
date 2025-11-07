package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faradayfan/sygkro/internal/config"
)

func TestRenderString_Basic(t *testing.T) {
	tmpl := "Hello, {{.name}}!"
	data := map[string]string{"name": "World"}
	out, err := RenderString(tmpl, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}
	want := "Hello, World!"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRenderString_Error(t *testing.T) {
	tmpl := "Hello, {{.name" // malformed
	data := map[string]string{"name": "World"}
	_, err := RenderString(tmpl, data)
	if err == nil {
		t.Errorf("expected error for malformed template")
	}
}

func TestProcessTemplateDir_BasicFlow(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	// Create a template file
	fileName := "greet_{{.who}}.txt"
	filePath := filepath.Join(src, fileName)
	content := "Hello, {{.who}}!"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}
	inputs := map[string]string{"who": "Alice"}
	opts := &config.TemplateOptions{}
	err := ProcessTemplateDir(src, dst, inputs, opts)
	if err != nil {
		t.Fatalf("ProcessTemplateDir failed: %v", err)
	}
	// Check output file
	wantFile := filepath.Join(dst, "greet_Alice.txt")
	data, err := os.ReadFile(wantFile)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	wantContent := "Hello, Alice!"
	if string(data) != wantContent {
		t.Errorf("output content mismatch: got %q, want %q", string(data), wantContent)
	}
}

func TestProcessTemplateDir_SkipRender(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	// Create a file to skip rendering
	fileName := "static.txt"
	filePath := filepath.Join(src, fileName)
	content := "{{.should_not_render}}"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write static file: %v", err)
	}
	inputs := map[string]string{"should_not_render": "RENDERED"}
	opts := &config.TemplateOptions{SkipRender: []string{"static.txt"}}
	err := ProcessTemplateDir(src, dst, inputs, opts)
	if err != nil {
		t.Fatalf("ProcessTemplateDir failed: %v", err)
	}
	// Check output file
	wantFile := filepath.Join(dst, "static.txt")
	data, err := os.ReadFile(wantFile)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	// Should be unchanged
	if string(data) != content {
		t.Errorf("skip render failed: got %q, want %q", string(data), content)
	}
}
