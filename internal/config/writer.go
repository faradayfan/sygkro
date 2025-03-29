// internal/config/yamlwriter.go
package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// WriteYAML writes the given data as YAML to the specified file path using a standard format.
// It uses 2 spaces per indent.
func WriteYAML(filePath string, data interface{}) error {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // Standardize on 2 spaces for indentation.
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}
	return nil
}
