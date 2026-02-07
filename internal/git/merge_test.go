package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupMergeDir creates a temp directory and writes the given files into it.
// Files is a map of relative path → content.
func setupMergeDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for relPath, content := range files {
		fullPath := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", relPath, err)
		}
	}
	return dir
}

// readFile reads a file and returns its content, failing the test on error.
func readFileContent(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(content)
}

// --- ThreeWayMerge Tests ---

func TestThreeWayMerge_CleanMerge_TemplateOnlyChange(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\n", // unchanged
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nmodified\nline3\n", // template changed line2
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file result, got %d", len(result.Files))
	}
	if result.Files[0].Status != MergeClean {
		t.Errorf("expected MergeClean, got %d", result.Files[0].Status)
	}
	if result.HasConflict {
		t.Error("expected no conflicts")
	}

	// Apply and verify
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	content := readFileContent(t, filepath.Join(ours, "file.txt"))
	expected := "line1\nmodified\nline3\n"
	if content != expected {
		t.Errorf("merged content = %q, want %q", content, expected)
	}
}

func TestThreeWayMerge_CleanMerge_UserOnlyChange(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nuser_change\nline3\n", // user changed line2
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\n", // template unchanged
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	// Template didn't change, so this should be MergeUnchanged
	for _, f := range result.Files {
		if f.Status != MergeUnchanged {
			t.Errorf("expected MergeUnchanged for %s, got %d", f.RelPath, f.Status)
		}
	}

	// User's change should be preserved (file untouched)
	content := readFileContent(t, filepath.Join(ours, "file.txt"))
	expected := "line1\nuser_change\nline3\n"
	if content != expected {
		t.Errorf("file content = %q, want %q", content, expected)
	}
}

func TestThreeWayMerge_CleanMerge_BothChangeDifferentLines(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\nline4\nline5\nline6\nline7\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nuser_change\nline3\nline4\nline5\nline6\nline7\n", // user changed line2
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\nline4\nline5\nline6\ntemplate_change\n", // template changed line7
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	if result.HasConflict {
		t.Error("expected no conflicts")
	}

	// Find the file result
	found := false
	for _, f := range result.Files {
		if f.RelPath == "file.txt" {
			found = true
			if f.Status != MergeClean {
				t.Errorf("expected MergeClean, got %d", f.Status)
			}
		}
	}
	if !found {
		t.Error("file.txt not found in results")
	}

	// Apply and verify both changes are present
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	content := readFileContent(t, filepath.Join(ours, "file.txt"))
	if !strings.Contains(content, "user_change") {
		t.Error("user change not preserved in merged content")
	}
	if !strings.Contains(content, "template_change") {
		t.Error("template change not applied in merged content")
	}
}

func TestThreeWayMerge_Conflict_BothChangeSameLines(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nuser_change\nline3\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "line1\ntemplate_change\nline3\n",
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	if !result.HasConflict {
		t.Error("expected conflicts")
	}

	found := false
	for _, f := range result.Files {
		if f.RelPath == "file.txt" {
			found = true
			if f.Status != MergeConflict {
				t.Errorf("expected MergeConflict, got %d", f.Status)
			}
			if f.ConflictPath != "file.txt.sygkro-conflict" {
				t.Errorf("conflict path = %q, want %q", f.ConflictPath, "file.txt.sygkro-conflict")
			}
		}
	}
	if !found {
		t.Error("file.txt not found in results")
	}

	// Apply and verify original is preserved, conflict file created
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	// Original should be untouched
	original := readFileContent(t, filepath.Join(ours, "file.txt"))
	if original != "line1\nuser_change\nline3\n" {
		t.Errorf("original file was modified: %q", original)
	}

	// Conflict file should exist with conflict markers
	conflictContent := readFileContent(t, filepath.Join(ours, "file.txt.sygkro-conflict"))
	if !strings.Contains(conflictContent, "<<<<<<<") {
		t.Error("conflict file missing <<<<<<< markers")
	}
	if !strings.Contains(conflictContent, "=======") {
		t.Error("conflict file missing ======= markers")
	}
	if !strings.Contains(conflictContent, ">>>>>>>") {
		t.Error("conflict file missing >>>>>>> markers")
	}
	if !strings.Contains(conflictContent, "user_change") {
		t.Error("conflict file missing user's change")
	}
	if !strings.Contains(conflictContent, "template_change") {
		t.Error("conflict file missing template's change")
	}
}

func TestThreeWayMerge_NewFileInTemplate(t *testing.T) {
	base := setupMergeDir(t, map[string]string{})
	ours := setupMergeDir(t, map[string]string{})
	theirs := setupMergeDir(t, map[string]string{
		"newfile.txt": "new content\n",
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	if result.HasConflict {
		t.Error("expected no conflicts")
	}

	found := false
	for _, f := range result.Files {
		if f.RelPath == "newfile.txt" {
			found = true
			if f.Status != MergeNewFile {
				t.Errorf("expected MergeNewFile, got %d", f.Status)
			}
		}
	}
	if !found {
		t.Error("newfile.txt not found in results")
	}

	// Apply and verify file is created
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	content := readFileContent(t, filepath.Join(ours, "newfile.txt"))
	if content != "new content\n" {
		t.Errorf("new file content = %q, want %q", content, "new content\n")
	}
}

func TestThreeWayMerge_DeletedFileInTemplate(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"removed.txt": "old content\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"removed.txt": "old content\n",
	})
	theirs := setupMergeDir(t, map[string]string{})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	found := false
	for _, f := range result.Files {
		if f.RelPath == "removed.txt" {
			found = true
			if f.Status != MergeDeletedFile {
				t.Errorf("expected MergeDeletedFile, got %d", f.Status)
			}
		}
	}
	if !found {
		t.Error("removed.txt not found in results")
	}

	// Apply and verify file is NOT deleted
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	if !fileExists(filepath.Join(ours, "removed.txt")) {
		t.Error("file should NOT be deleted by ApplyMerge")
	}
}

func TestThreeWayMerge_UserDeletedFile_NoTemplateChange(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "content\n",
	})
	ours := setupMergeDir(t, map[string]string{}) // user deleted
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "content\n", // template unchanged
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	// Should respect user's deletion since template didn't change
	for _, f := range result.Files {
		if f.RelPath == "file.txt" && f.Status != MergeUnchanged {
			t.Errorf("expected MergeUnchanged for user-deleted file, got %d", f.Status)
		}
	}

	// File should NOT reappear
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	if fileExists(filepath.Join(ours, "file.txt")) {
		t.Error("user-deleted file should not be recreated")
	}
}

func TestThreeWayMerge_UserDeletedFile_TemplateChanged(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "old\n",
	})
	ours := setupMergeDir(t, map[string]string{}) // user deleted
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "new\n", // template changed
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	found := false
	for _, f := range result.Files {
		if f.RelPath == "file.txt" {
			found = true
			if f.Status != MergeNewFile {
				t.Errorf("expected MergeNewFile, got %d", f.Status)
			}
		}
	}
	if !found {
		t.Error("file.txt not found in results")
	}

	// Apply — file should be created with new content
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	content := readFileContent(t, filepath.Join(ours, "file.txt"))
	if content != "new\n" {
		t.Errorf("file content = %q, want %q", content, "new\n")
	}
}

func TestThreeWayMerge_AllIdentical(t *testing.T) {
	content := "same content\n"
	base := setupMergeDir(t, map[string]string{"file.txt": content})
	ours := setupMergeDir(t, map[string]string{"file.txt": content})
	theirs := setupMergeDir(t, map[string]string{"file.txt": content})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	if result.HasConflict {
		t.Error("expected no conflicts")
	}
	// All unchanged — Files list should be empty (unchanged files are not added)
	for _, f := range result.Files {
		if f.Status != MergeUnchanged {
			t.Errorf("expected MergeUnchanged, got %d for %s", f.Status, f.RelPath)
		}
	}
}

func TestThreeWayMerge_SubdirectoryFiles(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"sub/dir/file.txt": "original\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"sub/dir/file.txt": "original\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"sub/dir/file.txt": "updated\n",
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	found := false
	for _, f := range result.Files {
		if f.RelPath == filepath.Join("sub", "dir", "file.txt") {
			found = true
			if f.Status != MergeClean {
				t.Errorf("expected MergeClean, got %d", f.Status)
			}
		}
	}
	if !found {
		t.Error("sub/dir/file.txt not found in results")
	}

	// Apply and verify
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	content := readFileContent(t, filepath.Join(ours, "sub", "dir", "file.txt"))
	if content != "updated\n" {
		t.Errorf("content = %q, want %q", content, "updated\n")
	}
}

func TestThreeWayMerge_NewFileInSubdirectory(t *testing.T) {
	base := setupMergeDir(t, map[string]string{})
	ours := setupMergeDir(t, map[string]string{})
	theirs := setupMergeDir(t, map[string]string{
		"new/sub/dir/file.txt": "new content\n",
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	// Apply and verify directory is created
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}
	content := readFileContent(t, filepath.Join(ours, "new", "sub", "dir", "file.txt"))
	if content != "new content\n" {
		t.Errorf("content = %q, want %q", content, "new content\n")
	}
}

func TestThreeWayMerge_EmptyBase_FirstSync(t *testing.T) {
	base := setupMergeDir(t, map[string]string{}) // empty base (first sync)
	ours := setupMergeDir(t, map[string]string{
		"existing.txt": "user content\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"existing.txt": "template content\n", // different from user's
		"newfile.txt":  "brand new\n",
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	// existing.txt: no base, different ours vs theirs → conflict
	// newfile.txt: no base, no ours → new file
	for _, f := range result.Files {
		switch f.RelPath {
		case "existing.txt":
			if f.Status != MergeConflict {
				t.Errorf("expected MergeConflict for existing.txt, got %d", f.Status)
			}
		case "newfile.txt":
			if f.Status != MergeNewFile {
				t.Errorf("expected MergeNewFile for newfile.txt, got %d", f.Status)
			}
		}
	}
}

func TestThreeWayMerge_EmptyBase_IdenticalFiles(t *testing.T) {
	base := setupMergeDir(t, map[string]string{}) // empty base
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "same content\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "same content\n", // identical to ours
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	// Same content, should be unchanged
	if result.HasConflict {
		t.Error("expected no conflicts for identical files")
	}
}

func TestThreeWayMerge_MultipleFiles_MixedOutcomes(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"clean.txt":     "line1\nline2\nline3\nline4\nline5\nline6\nline7\n",
		"conflict.txt":  "line1\nline2\nline3\n",
		"deleted.txt":   "will be deleted\n",
		"unchanged.txt": "stays the same\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"clean.txt":     "line1\nuser_line2\nline3\nline4\nline5\nline6\nline7\n", // user changed line2
		"conflict.txt":  "line1\nuser_change\nline3\n",                            // user changed line2
		"deleted.txt":   "will be deleted\n",
		"unchanged.txt": "stays the same\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"clean.txt":     "line1\nline2\nline3\nline4\nline5\nline6\ntemplate_line7\n", // template changed line7
		"conflict.txt":  "line1\ntemplate_change\nline3\n",                            // template also changed line2
		"unchanged.txt": "stays the same\n",
		"newfile.txt":   "brand new\n",
		// deleted.txt removed from template
	})

	result, err := ThreeWayMerge(base, ours, theirs)
	if err != nil {
		t.Fatalf("ThreeWayMerge failed: %v", err)
	}

	if !result.HasConflict {
		t.Error("expected HasConflict to be true")
	}

	statuses := make(map[string]MergeStatus)
	for _, f := range result.Files {
		statuses[f.RelPath] = f.Status
	}

	if s, ok := statuses["clean.txt"]; !ok || s != MergeClean {
		t.Errorf("clean.txt: expected MergeClean, got %v", s)
	}
	if s, ok := statuses["conflict.txt"]; !ok || s != MergeConflict {
		t.Errorf("conflict.txt: expected MergeConflict, got %v", s)
	}
	if s, ok := statuses["deleted.txt"]; !ok || s != MergeDeletedFile {
		t.Errorf("deleted.txt: expected MergeDeletedFile, got %v", s)
	}
	if s, ok := statuses["newfile.txt"]; !ok || s != MergeNewFile {
		t.Errorf("newfile.txt: expected MergeNewFile, got %v", s)
	}
	// unchanged.txt should not appear in results (it's filtered as MergeUnchanged)
	if s, ok := statuses["unchanged.txt"]; ok && s != MergeUnchanged {
		t.Errorf("unchanged.txt: expected not in results or MergeUnchanged, got %v", s)
	}

	// Apply and verify
	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	// clean.txt should have both changes
	cleanContent := readFileContent(t, filepath.Join(ours, "clean.txt"))
	if !strings.Contains(cleanContent, "user_line2") {
		t.Error("clean.txt missing user change")
	}
	if !strings.Contains(cleanContent, "template_line7") {
		t.Error("clean.txt missing template change")
	}

	// conflict.txt original should be untouched
	conflictOriginal := readFileContent(t, filepath.Join(ours, "conflict.txt"))
	if conflictOriginal != "line1\nuser_change\nline3\n" {
		t.Errorf("conflict.txt original modified: %q", conflictOriginal)
	}

	// conflict file should exist
	if !fileExists(filepath.Join(ours, "conflict.txt.sygkro-conflict")) {
		t.Error("conflict.txt.sygkro-conflict not created")
	}

	// newfile.txt should be created
	newContent := readFileContent(t, filepath.Join(ours, "newfile.txt"))
	if newContent != "brand new\n" {
		t.Errorf("newfile.txt content = %q", newContent)
	}

	// deleted.txt should still exist
	if !fileExists(filepath.Join(ours, "deleted.txt")) {
		t.Error("deleted.txt should not be auto-deleted")
	}
}

// --- mergeFile unit tests ---

func TestMergeFile_CleanMerge(t *testing.T) {
	baseDir := setupMergeDir(t, map[string]string{
		"file": "line1\nline2\nline3\nline4\nline5\nline6\nline7\n",
	})
	oursDir := setupMergeDir(t, map[string]string{
		"file": "line1\nuser\nline3\nline4\nline5\nline6\nline7\n",
	})
	theirsDir := setupMergeDir(t, map[string]string{
		"file": "line1\nline2\nline3\nline4\nline5\nline6\ntemplate\n",
	})

	merged, hasConflict, err := mergeFile(
		filepath.Join(baseDir, "file"),
		filepath.Join(oursDir, "file"),
		filepath.Join(theirsDir, "file"),
	)
	if err != nil {
		t.Fatalf("mergeFile failed: %v", err)
	}
	if hasConflict {
		t.Error("expected no conflict")
	}
	if !strings.Contains(string(merged), "user") {
		t.Error("merged missing user change")
	}
	if !strings.Contains(string(merged), "template") {
		t.Error("merged missing template change")
	}
}

func TestMergeFile_ConflictMerge(t *testing.T) {
	baseDir := setupMergeDir(t, map[string]string{
		"file": "line1\nline2\nline3\n",
	})
	oursDir := setupMergeDir(t, map[string]string{
		"file": "line1\nuser\nline3\n",
	})
	theirsDir := setupMergeDir(t, map[string]string{
		"file": "line1\ntemplate\nline3\n",
	})

	merged, hasConflict, err := mergeFile(
		filepath.Join(baseDir, "file"),
		filepath.Join(oursDir, "file"),
		filepath.Join(theirsDir, "file"),
	)
	if err != nil {
		t.Fatalf("mergeFile failed: %v", err)
	}
	if !hasConflict {
		t.Error("expected conflict")
	}
	if !strings.Contains(string(merged), "<<<<<<<") {
		t.Error("merged missing conflict markers")
	}
}

func TestMergeFile_EmptyBase(t *testing.T) {
	oursDir := setupMergeDir(t, map[string]string{
		"file": "user content\n",
	})
	theirsDir := setupMergeDir(t, map[string]string{
		"file": "template content\n",
	})

	merged, hasConflict, err := mergeFileWithEmptyBase(
		filepath.Join(oursDir, "file"),
		filepath.Join(theirsDir, "file"),
	)
	if err != nil {
		t.Fatalf("mergeFileWithEmptyBase failed: %v", err)
	}
	// With empty base and different content, expect conflict
	if !hasConflict {
		t.Error("expected conflict with empty base and different content")
	}
	if !strings.Contains(string(merged), "user content") {
		t.Error("merged missing user content")
	}
	if !strings.Contains(string(merged), "template content") {
		t.Error("merged missing template content")
	}
}

// --- ApplyMerge Tests ---

func TestApplyMerge_CleanFile(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "original\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "original\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "updated\n",
	})

	result := &MergeResult{
		Files: []MergeFileResult{
			{RelPath: "file.txt", Status: MergeClean},
		},
	}

	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	content := readFileContent(t, filepath.Join(ours, "file.txt"))
	if content != "updated\n" {
		t.Errorf("content = %q, want %q", content, "updated\n")
	}
}

func TestApplyMerge_ConflictFile(t *testing.T) {
	base := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nline2\nline3\n",
	})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "line1\nuser\nline3\n",
	})
	theirs := setupMergeDir(t, map[string]string{
		"file.txt": "line1\ntemplate\nline3\n",
	})

	result := &MergeResult{
		Files: []MergeFileResult{
			{RelPath: "file.txt", Status: MergeConflict, ConflictPath: "file.txt.sygkro-conflict"},
		},
		HasConflict: true,
	}

	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	// Original preserved
	original := readFileContent(t, filepath.Join(ours, "file.txt"))
	if original != "line1\nuser\nline3\n" {
		t.Errorf("original modified: %q", original)
	}

	// Conflict file created
	conflictContent := readFileContent(t, filepath.Join(ours, "file.txt.sygkro-conflict"))
	if !strings.Contains(conflictContent, "<<<<<<<") {
		t.Error("conflict file missing markers")
	}
}

func TestApplyMerge_NewFile(t *testing.T) {
	base := setupMergeDir(t, map[string]string{})
	ours := setupMergeDir(t, map[string]string{})
	theirs := setupMergeDir(t, map[string]string{
		"new.txt": "new content\n",
	})

	result := &MergeResult{
		Files: []MergeFileResult{
			{RelPath: "new.txt", Status: MergeNewFile},
		},
	}

	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	content := readFileContent(t, filepath.Join(ours, "new.txt"))
	if content != "new content\n" {
		t.Errorf("content = %q, want %q", content, "new content\n")
	}
}

func TestApplyMerge_DeletedFile(t *testing.T) {
	base := setupMergeDir(t, map[string]string{})
	ours := setupMergeDir(t, map[string]string{
		"file.txt": "content\n",
	})
	theirs := setupMergeDir(t, map[string]string{})

	result := &MergeResult{
		Files: []MergeFileResult{
			{RelPath: "file.txt", Status: MergeDeletedFile},
		},
	}

	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	// File should NOT be deleted
	if !fileExists(filepath.Join(ours, "file.txt")) {
		t.Error("file should not be deleted")
	}
}

func TestApplyMerge_NewFileInSubdirectory(t *testing.T) {
	base := setupMergeDir(t, map[string]string{})
	ours := setupMergeDir(t, map[string]string{})
	theirs := setupMergeDir(t, map[string]string{
		"deep/nested/dir/file.txt": "content\n",
	})

	result := &MergeResult{
		Files: []MergeFileResult{
			{RelPath: filepath.Join("deep", "nested", "dir", "file.txt"), Status: MergeNewFile},
		},
	}

	if err := ApplyMerge(ours, base, theirs, result); err != nil {
		t.Fatalf("ApplyMerge failed: %v", err)
	}

	content := readFileContent(t, filepath.Join(ours, "deep", "nested", "dir", "file.txt"))
	if content != "content\n" {
		t.Errorf("content = %q, want %q", content, "content\n")
	}
}
