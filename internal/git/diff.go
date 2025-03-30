// internal/git/diff.go
package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/faradayfan/sygkro/internal/config"
	"github.com/faradayfan/sygkro/internal/engine"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CopyOptions struct {
	IgnoreFilesRegex []string
}

// ComputeDiff renders the template using the inputs from the sync configuration into
// a temporary directory ("ideal") and computes a unified diff between that rendered output
// and the actual project directory.
func ComputeDiff(templatePath string, projectPath string, syncConfig *config.SyncConfig) (string, error) {
	// 1. Determine the subdirectory to render.
	idealSource := filepath.Join(templatePath, "{{ .slug }}")
	templateConfigPath := filepath.Join(templatePath, config.TemplateConfigFileName)
	templateConfig, err := config.ReadTemplateConfig(templateConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template config: %w", err)
	}

	// 2. Create a temporary directory to render the ideal template.
	idealDir, err := os.MkdirTemp("", "sygkro-ideal-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory for ideal template: %w", err)
	}
	// Ensure cleanup of the temporary directory.
	defer os.RemoveAll(idealDir)

	// 3. Render the template from the idealSource using the stored inputs and options.
	if err := engine.ProcessTemplateDir(idealSource, idealDir, syncConfig.Inputs, templateConfig.Options); err != nil {
		return "", fmt.Errorf("failed to render ideal template: %w", err)
	}

	// 4. Compute a diff between the rendered "ideal" template and the actual project directory.
	diffOutput, err := DiffDirs(idealDir, projectPath, []string{syncConfig.Path})
	if err != nil {
		return "", fmt.Errorf("failed to compute diff: %w", err)
	}

	// 5. Return the diff output.
	return diffOutput, nil
}

// DiffDirs creates temporary Git repositories from two directories,
// commits their contents, and computes a unified diff between the resulting trees.
// It returns the diff as a unified diff string.
func DiffDirs(dir1, dir2 string, ignoreFilesRegex []string) (string, error) {
	// Commit the first directory.
	_, tree1, cleanup1, err := commitDirToTempRepo(dir1, ignoreFilesRegex)
	if err != nil {
		return "", fmt.Errorf("failed to commit directory %s: %w", dir1, err)
	}
	defer cleanup1()

	// Commit the second directory.
	_, tree2, cleanup2, err := commitDirToTempRepo(dir2, ignoreFilesRegex)
	if err != nil {
		return "", fmt.Errorf("failed to commit directory %s: %w", dir2, err)
	}
	defer cleanup2()

	// Compute the diff between the two trees.
	changes, err := object.DiffTree(tree2, tree1)
	if err != nil {
		return "", fmt.Errorf("failed to compute diff: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	return patch.String(), nil
}

// commitDirToTempRepo creates a temporary Git repository from the given directory.
// It copies all the files from the directory into the repository, stages, and commits them.
// Returns the repository, the commit's tree, a cleanup function, and an error.
func commitDirToTempRepo(dir string, ignoreFilesRegex []string) (*git.Repository, *object.Tree, func(), error) {
	// Create a temporary directory for the repository.
	tmpDir, err := os.MkdirTemp("", "diff-repo-*")
	if err != nil {
		return nil, nil, nil, err
	}
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Initialize a new repository.
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	copyOpts := CopyOptions{
		IgnoreFilesRegex: ignoreFilesRegex,
	}
	// Copy the entire contents of the source directory to the temporary repository.
	if err := copyDir(dir, tmpDir, copyOpts); err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	// Get the working tree.
	wt, err := repo.Worktree()
	if err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	// Stage all files.
	_, err = wt.Add(".")
	if err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	// Commit the changes.
	commitHash, err := wt.Commit("commit", &git.CommitOptions{All: true})
	if err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		cleanup()
		return nil, nil, nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		cleanup()
		return nil, nil, nil, err
	}
	return repo, tree, cleanup, nil
}

// copyDir recursively copies all files and directories from src to dst.
func copyDir(src, dst string, opts CopyOptions) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Compute the relative path from the source.
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Check if the file should be ignored based on the opts.IgnoreFilesRegex.
		for _, regex := range opts.IgnoreFilesRegex {
			if matched, _ := filepath.Match(regex, relPath); matched {
				return nil // Skip this file.
			}
		}

		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			// Create the directory.
			return os.MkdirAll(targetPath, info.Mode())
		}
		// Read and copy the file.
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}
