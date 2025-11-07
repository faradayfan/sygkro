package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

type simpleStruct struct {
	Field string `yaml:"field"`
}

func TestWriteAndReadYAML_Simple(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "simple.yaml")
	original := simpleStruct{Field: "testvalue"}

	// Write YAML
	if err := WriteYAML(filePath, original); err != nil {
		t.Fatalf("WriteYAML failed: %v", err)
	}

	// Read YAML
	var result simpleStruct
	if err := ReadYAML(filePath, &result); err != nil {
		t.Fatalf("ReadYAML failed: %v", err)
	}

	if !reflect.DeepEqual(original, result) {
		t.Errorf("YAML roundtrip mismatch: got %+v, want %+v", result, original)
	}
}
