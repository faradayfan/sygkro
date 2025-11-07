package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/faradayfan/sygkro/internal/config"
	"gopkg.in/yaml.v3"
)

func TestWriteSyncConfig_CreatesFileAndCleansUp(t *testing.T) {
	// Setup temp dir
	tempDir := t.TempDir()
	syncCfg := config.SyncConfig{
		// Add fields as needed for a minimal valid config
	}

	err := WriteSyncConfig(tempDir, syncCfg)
	if err != nil {
		t.Fatalf("WriteSyncConfig failed: %v", err)
	}

	syncFilePath := filepath.Join(tempDir, config.SyncConfigFileName)
	// Check file exists
	info, err := os.Stat(syncFilePath)
	if err != nil {
		t.Fatalf("sync config file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("sync config file is empty")
	}

	// Check file contents are valid YAML
	data, err := os.ReadFile(syncFilePath)
	if err != nil {
		t.Fatalf("failed to read sync config file: %v", err)
	}
	var out config.SyncConfig
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Errorf("sync config file is not valid YAML: %v", err)
	}
	// No need to cleanup, t.TempDir() is auto-removed
}
