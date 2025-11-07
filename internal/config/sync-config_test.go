package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSyncConfig_WriteAndReadSyncConfig(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, SyncConfigFileName)
	original := &SyncConfig{
		Source: SourceConfig{
			TemplatePath:        "foo/path",
			TemplateName:        "basic",
			TemplateVersion:     "1.0.0",
			TemplateTrackingRef: "main",
		},
		Inputs: map[string]string{"key": "value"},
	}

	// Write
	if err := original.Write(filePath); err != nil {
		t.Fatalf("SyncConfig.Write failed: %v", err)
	}

	// Read
	readCfg, err := ReadSyncConfig(filePath)
	if err != nil {
		t.Fatalf("ReadSyncConfig failed: %v", err)
	}

	// Path field should be set
	if readCfg.Path != filePath {
		t.Errorf("Path field not set: got %q, want %q", readCfg.Path, filePath)
	}

	// Compare Source and Inputs
	if !reflect.DeepEqual(original.Source, readCfg.Source) {
		t.Errorf("Source mismatch: got %+v, want %+v", readCfg.Source, original.Source)
	}
	if !reflect.DeepEqual(original.Inputs, readCfg.Inputs) {
		t.Errorf("Inputs mismatch: got %+v, want %+v", readCfg.Inputs, original.Inputs)
	}
}
