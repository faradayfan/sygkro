package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestTemplateConfig_WriteAndReadTemplateConfig(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, TemplateConfigFileName)
	original := &TemplateConfig{
		Name:        "basic",
		Description: "A basic template",
		Version:     "1.0.0",
		Templating: TemplatingConfig{
			Inputs: map[string]string{"foo": "bar"},
		},
		Options: &TemplateOptions{SkipRender: []string{"README.md"}},
	}

	// Write
	if err := original.Write(filePath); err != nil {
		t.Fatalf("TemplateConfig.Write failed: %v", err)
	}

	// Read
	readCfg, err := ReadTemplateConfig(filePath)
	if err != nil {
		t.Fatalf("ReadTemplateConfig failed: %v", err)
	}

	// Compare all fields
	if !reflect.DeepEqual(original, readCfg) {
		t.Errorf("TemplateConfig roundtrip mismatch: got %+v, want %+v", readCfg, original)
	}
}
