package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// MergeStatus represents the outcome of merging a single file.
type MergeStatus int

const (
	MergeClean       MergeStatus = iota // Clean merge applied
	MergeConflict                       // Conflict detected, .sygkro-conflict file created
	MergeNewFile                        // New file from template added to project
	MergeDeletedFile                    // File deleted in template, not auto-deleted
	MergeUnchanged                      // No changes needed
)

// MergeFileResult represents the outcome of merging a single file.
type MergeFileResult struct {
	RelPath      string      // Relative path within the project
	Status       MergeStatus // Outcome of the merge
	ConflictPath string      // Path to .sygkro-conflict file if Status == MergeConflict
}

// MergeResult represents the outcome of merging all template files.
type MergeResult struct {
	Files       []MergeFileResult
	HasConflict bool
}

// ThreeWayMerge performs a 3-way merge of template changes into the project.
//
// baseDir:   rendered old template (at the previously synced commit)
// oursDir:   current project directory (with user customizations)
// theirsDir: rendered new template (at the latest commit)
//
// For each file, it determines the appropriate action based on which
// directories contain the file and whether contents have changed.
func ThreeWayMerge(baseDir, oursDir, theirsDir string) (*MergeResult, error) {
	baseFiles, err := collectFiles(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect base files: %w", err)
	}

	theirsFiles, err := collectFiles(theirsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect theirs files: %w", err)
	}

	// Build union of all file paths
	allFiles := make(map[string]bool)
	for f := range baseFiles {
		allFiles[f] = true
	}
	for f := range theirsFiles {
		allFiles[f] = true
	}

	result := &MergeResult{}

	for relPath := range allFiles {
		_, inBase := baseFiles[relPath]
		_, inTheirs := theirsFiles[relPath]

		oursPath := filepath.Join(oursDir, relPath)
		oursExists := fileExists(oursPath)

		fileResult, err := mergeOneFile(baseDir, oursDir, theirsDir, relPath, inBase, oursExists, inTheirs)
		if err != nil {
			return nil, fmt.Errorf("failed to merge %s: %w", relPath, err)
		}

		if fileResult.Status != MergeUnchanged {
			result.Files = append(result.Files, *fileResult)
		}
		if fileResult.Status == MergeConflict {
			result.HasConflict = true
		}
	}

	return result, nil
}

// mergeOneFile determines and executes the merge strategy for a single file.
func mergeOneFile(baseDir, oursDir, theirsDir, relPath string, inBase, oursExists, inTheirs bool) (*MergeFileResult, error) {
	basePath := filepath.Join(baseDir, relPath)
	oursPath := filepath.Join(oursDir, relPath)
	theirsPath := filepath.Join(theirsDir, relPath)

	switch {
	case inBase && oursExists && inTheirs:
		// Normal case: file exists in all three — 3-way merge
		baseContent, _ := os.ReadFile(basePath)
		theirsContent, _ := os.ReadFile(theirsPath)

		// If template didn't change, nothing to do
		if bytes.Equal(baseContent, theirsContent) {
			return &MergeFileResult{RelPath: relPath, Status: MergeUnchanged}, nil
		}

		oursContent, _ := os.ReadFile(oursPath)

		// If user hasn't modified, just take theirs
		if bytes.Equal(baseContent, oursContent) {
			return &MergeFileResult{
				RelPath: relPath,
				Status:  MergeClean,
			}, nil
		}

		// Both changed — run git merge-file
		_, hasConflict, err := mergeFile(basePath, oursPath, theirsPath)
		if err != nil {
			return nil, err
		}

		if hasConflict {
			return &MergeFileResult{
				RelPath:      relPath,
				Status:       MergeConflict,
				ConflictPath: relPath + ".sygkro-conflict",
			}, nil
		}

		return &MergeFileResult{
			RelPath: relPath,
			Status:  MergeClean,
		}, nil

	case inBase && oursExists && !inTheirs:
		// Template deleted the file — report but don't auto-delete
		return &MergeFileResult{RelPath: relPath, Status: MergeDeletedFile}, nil

	case inBase && !oursExists && inTheirs:
		// User deleted the file
		baseContent, _ := os.ReadFile(basePath)
		theirsContent, _ := os.ReadFile(theirsPath)
		if bytes.Equal(baseContent, theirsContent) {
			// Template didn't change — respect user's deletion
			return &MergeFileResult{RelPath: relPath, Status: MergeUnchanged}, nil
		}
		// Template changed — treat as new file
		return &MergeFileResult{RelPath: relPath, Status: MergeNewFile}, nil

	case inBase && !oursExists && !inTheirs:
		// Both deleted — nothing to do
		return &MergeFileResult{RelPath: relPath, Status: MergeUnchanged}, nil

	case !inBase && oursExists && inTheirs:
		// File exists in project and new template but not in old template
		oursContent, _ := os.ReadFile(oursPath)
		theirsContent, _ := os.ReadFile(theirsPath)
		if bytes.Equal(oursContent, theirsContent) {
			return &MergeFileResult{RelPath: relPath, Status: MergeUnchanged}, nil
		}
		// Different contents, no common ancestor — conflict
		// Use empty base for merge-file
		_, hasConflict, err := mergeFileWithEmptyBase(oursPath, theirsPath)
		if err != nil {
			return nil, err
		}
		if hasConflict {
			return &MergeFileResult{
				RelPath:      relPath,
				Status:       MergeConflict,
				ConflictPath: relPath + ".sygkro-conflict",
			}, nil
		}
		return &MergeFileResult{RelPath: relPath, Status: MergeClean}, nil

	case !inBase && !oursExists && inTheirs:
		// New file from template — add to project
		return &MergeFileResult{RelPath: relPath, Status: MergeNewFile}, nil

	default:
		return &MergeFileResult{RelPath: relPath, Status: MergeUnchanged}, nil
	}
}

// mergeFile runs git merge-file on three files and returns the merged content.
// Returns (mergedContent, hasConflict, error).
func mergeFile(basePath, oursPath, theirsPath string) ([]byte, bool, error) {
	// git merge-file modifies the first file in-place, so we work with temp copies
	tmpDir, err := os.MkdirTemp("", "sygkro-merge-*")
	if err != nil {
		return nil, false, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	oursCopy := filepath.Join(tmpDir, "ours")
	baseCopy := filepath.Join(tmpDir, "base")
	theirsCopy := filepath.Join(tmpDir, "theirs")

	oursContent, err := os.ReadFile(oursPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read ours: %w", err)
	}
	baseContent, err := os.ReadFile(basePath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read base: %w", err)
	}
	theirsContent, err := os.ReadFile(theirsPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read theirs: %w", err)
	}

	if err := os.WriteFile(oursCopy, oursContent, 0644); err != nil {
		return nil, false, err
	}
	if err := os.WriteFile(baseCopy, baseContent, 0644); err != nil {
		return nil, false, err
	}
	if err := os.WriteFile(theirsCopy, theirsContent, 0644); err != nil {
		return nil, false, err
	}

	cmd := exec.Command("git", "merge-file", "-p", "--diff3",
		"-L", "project",
		"-L", "base",
		"-L", "template",
		oursCopy, baseCopy, theirsCopy,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// git merge-file exit codes:
	// 0 = clean merge
	// >0 = number of conflicts (but still produces output)
	// <0 = error
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() > 0 {
				// Conflicts detected, output still valid
				return stdout.Bytes(), true, nil
			}
		}
		return nil, false, fmt.Errorf("git merge-file failed: %v: %s", err, stderr.String())
	}

	return stdout.Bytes(), false, nil
}

// mergeFileWithEmptyBase runs a merge with an empty base (no common ancestor).
func mergeFileWithEmptyBase(oursPath, theirsPath string) ([]byte, bool, error) {
	tmpDir, err := os.MkdirTemp("", "sygkro-merge-*")
	if err != nil {
		return nil, false, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	emptyBase := filepath.Join(tmpDir, "base")
	if err := os.WriteFile(emptyBase, []byte{}, 0644); err != nil {
		return nil, false, err
	}

	oursCopy := filepath.Join(tmpDir, "ours")
	theirsCopy := filepath.Join(tmpDir, "theirs")

	oursContent, err := os.ReadFile(oursPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read ours: %w", err)
	}
	theirsContent, err := os.ReadFile(theirsPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read theirs: %w", err)
	}

	if err := os.WriteFile(oursCopy, oursContent, 0644); err != nil {
		return nil, false, err
	}
	if err := os.WriteFile(theirsCopy, theirsContent, 0644); err != nil {
		return nil, false, err
	}

	cmd := exec.Command("git", "merge-file", "-p", "--diff3",
		"-L", "project",
		"-L", "base",
		"-L", "template",
		oursCopy, emptyBase, theirsCopy,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() > 0 {
				return stdout.Bytes(), true, nil
			}
		}
		return nil, false, fmt.Errorf("git merge-file failed: %v: %s", err, stderr.String())
	}

	return stdout.Bytes(), false, nil
}

// ApplyMerge applies the merge result to the project directory.
// For clean merges, the project file is overwritten with the merged content.
// For conflicts, the original file is kept and a .sygkro-conflict file is created.
// For new files, the file is created from the template.
// For deleted files, no action is taken (only reported).
func ApplyMerge(projectDir, baseDir, theirsDir string, result *MergeResult) error {
	for _, f := range result.Files {
		projectPath := filepath.Join(projectDir, f.RelPath)

		switch f.Status {
		case MergeClean:
			basePath := filepath.Join(baseDir, f.RelPath)
			oursPath := projectPath
			theirsPath := filepath.Join(theirsDir, f.RelPath)

			var merged []byte
			var err error

			if fileExists(basePath) {
				merged, _, err = mergeFile(basePath, oursPath, theirsPath)
			} else {
				// No base — but was determined clean (identical files)
				merged, err = os.ReadFile(theirsPath)
			}
			if err != nil {
				return fmt.Errorf("failed to merge %s: %w", f.RelPath, err)
			}

			// Preserve original file permissions
			info, err := os.Stat(oursPath)
			mode := os.FileMode(0644)
			if err == nil {
				mode = info.Mode()
			}

			if err := os.WriteFile(projectPath, merged, mode); err != nil {
				return fmt.Errorf("failed to write merged file %s: %w", f.RelPath, err)
			}

		case MergeConflict:
			basePath := filepath.Join(baseDir, f.RelPath)
			oursPath := projectPath
			theirsPath := filepath.Join(theirsDir, f.RelPath)

			var merged []byte
			var err error

			if fileExists(basePath) {
				merged, _, err = mergeFile(basePath, oursPath, theirsPath)
			} else {
				merged, _, err = mergeFileWithEmptyBase(oursPath, theirsPath)
			}
			if err != nil {
				return fmt.Errorf("failed to merge %s: %w", f.RelPath, err)
			}

			conflictPath := filepath.Join(projectDir, f.ConflictPath)
			if err := os.MkdirAll(filepath.Dir(conflictPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for conflict file: %w", err)
			}
			if err := os.WriteFile(conflictPath, merged, 0644); err != nil {
				return fmt.Errorf("failed to write conflict file %s: %w", f.ConflictPath, err)
			}

		case MergeNewFile:
			theirsPath := filepath.Join(theirsDir, f.RelPath)
			content, err := os.ReadFile(theirsPath)
			if err != nil {
				return fmt.Errorf("failed to read new template file %s: %w", f.RelPath, err)
			}

			if err := os.MkdirAll(filepath.Dir(projectPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for new file: %w", err)
			}
			if err := os.WriteFile(projectPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write new file %s: %w", f.RelPath, err)
			}

		case MergeDeletedFile:
			// Don't delete — just reported in the result

		case MergeUnchanged:
			// Nothing to do
		}
	}

	return nil
}

// collectFiles walks a directory and returns a set of relative file paths.
func collectFiles(dir string) (map[string]bool, error) {
	files := make(map[string]bool)
	if dir == "" {
		return files, nil
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return files, nil
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files[relPath] = true
		return nil
	})

	return files, err
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
