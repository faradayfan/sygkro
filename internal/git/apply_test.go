package git

import (
	"testing"
)

func TestApplyDiff_InvalidRepo(t *testing.T) {
	// Use a temp dir that is not a git repo
	tempDir := t.TempDir()
	diff := "invalid diff"
	err := ApplyDiff(tempDir, diff)
	if err == nil {
		t.Errorf("expected error for invalid diff in non-git repo")
	}
}

func TestApplyDiff_EmptyDiff(t *testing.T) {
	tempDir := t.TempDir()
	err := ApplyDiff(tempDir, "")
	if err == nil {
		t.Errorf("expected error for empty diff in non-git repo")
	}
}

func TestApplyDiff_ValidDiff(t *testing.T) {

}
