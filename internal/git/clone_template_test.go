package git

import (
	"os"
	"testing"
)

func TestGetTemplateReferenceType(t *testing.T) {
	tempDir := t.TempDir()
	cases := []struct {
		ref      string
		wantType TemplateReferenceType
	}{
		{"git@github.com:owner/repo.git", TemplateReferenceTypeSSH},
		{"https://github.com/owner/repo.git", TemplateReferenceTypeHTTPS},
		{"gh:owner/repo", TemplateReferenceTypeSimpleGH},
		{tempDir, TemplateReferenceTypeLocalPath},
		{"not_a_repo", TemplateReferenceTypeUnknown},
	}
	for _, tc := range cases {
		got := GetTemplateReferenceType(tc.ref)
		if got != tc.wantType {
			t.Errorf("ref %q: got %v, want %v", tc.ref, got, tc.wantType)
		}
	}
}

func TestGetTemplateDir_LocalPath(t *testing.T) {
	tempDir := t.TempDir()
	res, err := GetTemplateDir(tempDir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Path != tempDir {
		t.Errorf("expected Path to be %q, got %q", tempDir, res.Path)
	}
	if res.Cleanup == nil {
		t.Errorf("Cleanup should not be nil")
	}
	res.Cleanup()
	// Should not error
}

func TestGetTemplateDir_InvalidRepo(t *testing.T) {
	_, err := GetTemplateDir("git@github.com:nonexistent/repo.git", "")
	if err == nil {
		t.Errorf("expected error for invalid repo")
	}
}

func TestGetTemplateDir_ClonePublicRepo(t *testing.T) {
	res, err := GetTemplateDir("https://github.com/octocat/Hello-World.git", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Path == "" {
		t.Errorf("expected Path to be set")
	}
	if res.CommitSHA == "" {
		t.Errorf("expected CommitSHA to be set")
	}
	if res.HeadRef == "" {
		t.Errorf("expected HeadRef to be set")
	}
	res.Cleanup()
	if _, err := os.Stat(res.Path); !os.IsNotExist(err) {
		t.Errorf("expected temp dir to be cleaned up")
	}
}
