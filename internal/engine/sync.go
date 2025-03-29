package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/faradayfan/sygkro/internal/config"
	"gopkg.in/yaml.v3"
)

// WriteSyncConfig marshals the sync configuration and writes it to the destination directory.
func WriteSyncConfig(destination string, syncConfig config.SyncConfig) error {
	data, err := yaml.Marshal(syncConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal sync config: %w", err)
	}

	syncFilePath := filepath.Join(destination, config.SyncConfigFileName)
	if err := os.WriteFile(syncFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sync config file: %w", err)
	}

	return nil
}
