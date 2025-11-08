// internal/git/git.go
package git

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// commitRegex is precompiled to detect commit SHA strings.
var commitRegex = regexp.MustCompile("^[0-9a-fA-F]{7,40}$")

// enum for template reference types
type TemplateReferenceType int

const (
	TemplateReferenceTypeUnknown TemplateReferenceType = iota
	TemplateReferenceTypeSSH
	TemplateReferenceTypeHTTPS
	TemplateReferenceTypeSimpleGH
	TemplateReferenceTypeLocalPath
)

// TemplateDirResult holds the result of GetTemplateDir.
type TemplateDirResult struct {
	Path      string // Local directory path for the repository
	CommitSHA string // HEAD commit SHA (if available)
	HeadRef   string // HEAD reference (e.g., branch or tag name)
	Cleanup   func() // Function to clean up resources (e.g., remove temporary directory)
}

// GetTemplateReferenceType determines the type of the template reference.
func GetTemplateReferenceType(templateRef string) TemplateReferenceType {
	if strings.HasPrefix(templateRef, "git@") && strings.Contains(templateRef, ":") && strings.HasSuffix(templateRef, ".git") {
		return TemplateReferenceTypeSSH
	}

	if strings.HasPrefix(templateRef, "https://") && strings.HasSuffix(templateRef, ".git") {
		return TemplateReferenceTypeHTTPS
	}

	if strings.HasPrefix(templateRef, "gh:") {
		return TemplateReferenceTypeSimpleGH
	}

	if _, err := os.Stat(templateRef); err == nil {
		return TemplateReferenceTypeLocalPath
	}

	return TemplateReferenceTypeUnknown
}

// GetTemplateDir takes a template reference and a separate reference parameter,
// and returns a TemplateDirResult and an error.
// Supported templateRef formats:
//   - SSH clone URL: git@github.com:<repo-owner>/<repo>.git
//   - HTTPS clone URL: https://github.com/<repo-owner>/<repo>.git
//   - Simplified syntax: gh:<repo-owner>/<repo>
//
// The reference parameter specifies the branch, tag, or commit SHA to use. If it is an empty string,
// no ref is specified so the default branch will be checked out.
func GetTemplateDir(templateRef string, reference string) (*TemplateDirResult, error) {
	var (
		isCommit bool
		isTag    bool
	)
	gitRef := reference

	templateRefType := GetTemplateReferenceType(templateRef)

	switch templateRefType {
	case TemplateReferenceTypeUnknown:
		return nil, fmt.Errorf("unsupported template reference format: %s", templateRef)
	case TemplateReferenceTypeSimpleGH:
		repoSpec := strings.TrimPrefix(templateRef, "gh:")
		templateRef = fmt.Sprintf("git@github.com:%s.git", repoSpec)
	case TemplateReferenceTypeLocalPath:
		return &TemplateDirResult{
			Path:    templateRef,
			Cleanup: func() {},
		}, nil
	}

	tmpDir, err := os.MkdirTemp("", "sygkro-template-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	success := false
	defer func() {
		if !success {
			cleanup()
		}
	}()

	cloneOpts := &git.CloneOptions{
		URL: templateRef,
	}

	// Determine if the provided ref is a commit SHA.
	if gitRef != "" {
		isCommit = commitRegex.MatchString(gitRef)
		if isCommit {
			// For commit checkout, we need full history.
			cloneOpts.Depth = 0
		} else {
			// Assume it's a branch by default.
			cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(gitRef)
			cloneOpts.SingleBranch = true
			cloneOpts.Depth = 1
		}
	} else {
		// If no ref is provided, perform a shallow clone of the default branch.
		cloneOpts.Depth = 1
	}

	// print out clone options for debugging

	// Clone the repository.
	repo, err := git.PlainClone(tmpDir, false, cloneOpts)
	if err != nil && gitRef != "" && !isCommit {
		// If branch clone fails, try assuming it's a tag.
		cloneOpts.ReferenceName = plumbing.NewTagReferenceName(gitRef)
		repo, err = git.PlainClone(tmpDir, false, cloneOpts)
		isTag = true
	}
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository %s: %w", templateRef, err)
	}

	// If gitRef is a commit SHA, checkout that commit.
	if gitRef != "" && isCommit {
		wt, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("failed to get worktree: %w", err)
		}
		err = wt.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(gitRef),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to checkout commit %s: %w", gitRef, err)
		}
	}

	// Use the cloned repository handle to get the HEAD commit SHA and ref.
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}
	headRef := head.Name().String()
	if isTag {
		tagHead, err := repo.Tag(gitRef)
		if err != nil {
			return nil, fmt.Errorf("failed to get tag %s: %w", gitRef, err)
		}
		headRef = tagHead.Name().String()
	}
	commitSHA := head.Hash().String()

	success = true
	return &TemplateDirResult{
		Path:      tmpDir,
		CommitSHA: commitSHA,
		HeadRef:   headRef,
		Cleanup:   cleanup,
	}, nil
}
