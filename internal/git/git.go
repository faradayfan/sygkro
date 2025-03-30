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

// TemplateDirResult holds the result of GetTemplateDir.
type TemplateDirResult struct {
	Path      string // Local directory path for the repository
	CommitSHA string // HEAD commit SHA (if available)
	HeadRef   string // HEAD reference (e.g., branch or tag name)
	Cleanup   func() // Function to clean up resources (e.g., remove temporary directory)
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
	gitRef := ""

	// Check if the templateRef includes a ref (using '@').
	if strings.Contains(templateRef, "@") {
		parts := strings.SplitN(templateRef, "@", 2)
		templateRef = parts[0]
		gitRef = parts[1]
	}

	// Override gitRef with the separate reference parameter if provided.
	if reference != "" {
		gitRef = reference
	}

	// Convert simplified syntax to SSH URL if needed.
	if strings.HasPrefix(templateRef, "gh:") {
		repoSpec := strings.TrimPrefix(templateRef, "gh:")
		templateRef = fmt.Sprintf("git@github.com:%s.git", repoSpec)
	}

	// Check if the templateRef is a valid Git URL.
	if strings.HasPrefix(templateRef, "git@github.com:") ||
		(strings.HasPrefix(templateRef, "https://github.com/") && strings.HasSuffix(templateRef, ".git")) {

		repoURL := templateRef

		// Create a temporary directory for cloning.
		tmpDir, err := os.MkdirTemp("", "sygkro-template-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary directory: %w", err)
		}
		// Define a cleanup function to remove the temporary directory.
		cleanup := func() {
			os.RemoveAll(tmpDir)
		}

		// Ensure that on error the temporary directory is cleaned up.
		success := false
		defer func() {
			if !success {
				cleanup()
			}
		}()

		// Set up clone options.
		cloneOpts := &git.CloneOptions{
			URL: repoURL,
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

		// Clone the repository.
		repo, err := git.PlainClone(tmpDir, false, cloneOpts)
		if err != nil && gitRef != "" && !isCommit {
			// If branch clone fails, try assuming it's a tag.
			cloneOpts.ReferenceName = plumbing.NewTagReferenceName(gitRef)
			repo, err = git.PlainClone(tmpDir, false, cloneOpts)
			isTag = true
		}
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
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

		// Mark the operation as successful so that the deferred cleanup is not triggered.
		success = true
		return &TemplateDirResult{
			Path:      tmpDir,
			CommitSHA: commitSHA,
			HeadRef:   headRef,
			Cleanup:   cleanup,
		}, nil
	}

	// Otherwise, assume it's a local directory.
	return &TemplateDirResult{
		Path:    templateRef,
		Cleanup: func() {},
	}, nil
}
