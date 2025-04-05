// internal/git/diff.go
package git

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/engine"
)

// ComputeDiff renders the template using the inputs from the sync configuration into
// a temporary directory ("ideal") and computes a unified diff between that rendered output
// and the actual project directory.
func ComputeDiff(templateDir string, projectDir string, idealRevision string, syncConfig *config.SyncConfig) (string, error) {
	// create a temporary directory for the current state
	currentTmpDir, err := os.MkdirTemp("", "sygkro-current-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(currentTmpDir); err != nil {
			fmt.Printf("failed to remove current temporary directory: %v\n", err)
		}
	}()

	// create a temporary directory for the ideal state
	idealTmpDir, err := os.MkdirTemp("", "sygkro-ideal-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(idealTmpDir); err != nil {
			fmt.Printf("failed to remove ideal temporary directory: %v\n", err)
		}
	}()

	// read the template config file for the ideal revision
	idealTemplateConfig, err := config.ReadTemplateConfig(path.Join(templateDir, config.TemplateConfigFileName))
	if err != nil {
		return "", fmt.Errorf("failed to read template config file: %w", err)
	}

	expectedSubDir := filepath.Join(templateDir, "{{ .slug }}")

	// render the template the ideal revision into the ideal temporary directory
	if err := engine.ProcessTemplateDir(expectedSubDir, idealTmpDir, syncConfig.Inputs, idealTemplateConfig.Options); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// for every file in the templateDir, copy the files with the same name from the projectDir into currentTmpDir
	walkErr := filepath.Walk(idealTmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// remove the idealTmpDir prefix from the path
		relPath := strings.TrimPrefix(path, idealTmpDir+string(os.PathSeparator))

		projectDifFilePath := filepath.Join(projectDir, relPath)
		currentTmpDirFilePath := filepath.Join(currentTmpDir, relPath)

		// if the file is a directory, create the directory in the currentTmpDir
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(currentTmpDir, relPath), info.Mode())
		}

		// if the file exists in the projectDir and is not a directory, copy it to the currentTmpDir
		content, err := os.ReadFile(projectDifFilePath)
		if err != nil {
			return err
		}
		return os.WriteFile(currentTmpDirFilePath, content, info.Mode())
	})

	if walkErr != nil {
		return "", fmt.Errorf("failed to walk template directory: %w", err)
	}

	// compute the diff between the current and ideal temporary directories
	diff, err := gitDiff(currentTmpDir, idealTmpDir)
	if err != nil {
		return "", fmt.Errorf("failed to compute diff: %w", err)
	}

	return diff, nil
}
