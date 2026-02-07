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
		// print the path
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

		if _, err := os.Stat(projectDifFilePath); err == nil {
			// if the file exists in the projectDir
			// copy it to the currentTmpDir
			content, err := os.ReadFile(projectDifFilePath)
			if err != nil {
				return err
			}
			return os.WriteFile(currentTmpDirFilePath, content, info.Mode())
		} else if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	})

	if walkErr != nil {
		return "", fmt.Errorf("failed to walk template directory: %w", walkErr)
	}

	// compute the diff between the current and ideal temporary directories
	diff, err := gitDiff(currentTmpDir, idealTmpDir)
	if err != nil {
		return "", fmt.Errorf("failed to compute diff: %w", err)
	}

	return diff, nil
}

// ComputeTemplateDiff renders the template at both the old and new versions
// and returns a unified diff showing only what changed in the template.
// This is useful for previewing what a sync will bring in.
//
// templateDir should be a cloned repo with full history (use GetTemplateDirForSync).
// oldVersion is the commit SHA of the previously synced template version.
func ComputeTemplateDiff(templateDir string, oldVersion string, syncConfig *config.SyncConfig) (string, error) {
	// Render NEW template (current HEAD)
	newTmpDir, err := os.MkdirTemp("", "sygkro-diff-new-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(newTmpDir)

	if err := RenderTemplateAtPath(templateDir, newTmpDir, syncConfig.Inputs); err != nil {
		return "", fmt.Errorf("failed to render new template: %w", err)
	}

	// Render OLD template
	oldTmpDir, err := os.MkdirTemp("", "sygkro-diff-old-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(oldTmpDir)

	if oldVersion != "" {
		if err := GitCheckout(templateDir, oldVersion); err != nil {
			return "", fmt.Errorf("failed to checkout old version %s: %w", oldVersion, err)
		}
		if err := RenderTemplateAtPath(templateDir, oldTmpDir, syncConfig.Inputs); err != nil {
			return "", fmt.Errorf("failed to render old template: %w", err)
		}
	}
	// If oldVersion is empty (first sync), oldTmpDir stays empty — everything shows as added

	diff, err := gitDiff(oldTmpDir, newTmpDir)
	if err != nil {
		return "", fmt.Errorf("failed to compute diff: %w", err)
	}

	return diff, nil
}
