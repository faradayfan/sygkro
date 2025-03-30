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

// GetTemplateDir takes a template reference and a separate reference parameter,
// and returns a local directory path, the HEAD commit SHA (if available), the HEAD ref name,
// a cleanup function, and an error.
// Supported templateRef formats:
//   - SSH clone URL: git@github.com:<repo-owner>/<repo>.git
//   - HTTPS clone URL: https://github.com/<repo-owner>/<repo>.git
//   - Simplified syntax: gh:<repo-owner>/<repo>
//
// The reference parameter specifies the branch, tag, or commit SHA to use. If it is an empty string,
// no ref is specified so the default branch will be checked out.
func GetTemplateDir(templateRef string, reference string) (string, string, string, func(), error) {
	// Check if the template reference itself contains a Git ref (indicated by '@').
	var (
		isCommit bool = false
		isTag    bool = false
	)
	gitRef := ""
	if strings.Contains(templateRef, "@") {
		parts := strings.Split(templateRef, "@")
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

	if strings.HasPrefix(templateRef, "git@github.com:") ||
		(strings.HasPrefix(templateRef, "https://github.com/") && strings.HasSuffix(templateRef, ".git")) {
		repoURL := templateRef

		// Create a temporary directory for cloning.
		tmpDir, err := os.MkdirTemp("", "sygkro-template-*")
		if err != nil {
			return "", "", "", nil, fmt.Errorf("failed to create temporary directory: %w", err)
		}

		// Prepare the clone options.
		cloneOpts := &git.CloneOptions{
			URL:   repoURL,
			Depth: 0,
		}

		// If a non-empty reference is provided, adjust the clone options.
		if gitRef != "" {
			// Determine if gitRef is a commit SHA.
			isCommit, _ = regexp.MatchString("^[0-9a-fA-F]{7,40}$", gitRef)
			if isCommit {
				// We'll handle commit checkout after cloning.
			} else {
				// Assume it's a branch
				cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(gitRef)
				cloneOpts.SingleBranch = true
			}
		}

		// Clone the repository.
		repo, err := git.PlainClone(tmpDir, false, cloneOpts)
		if err != nil && gitRef != "" {
			cloneOpts.ReferenceName = plumbing.NewTagReferenceName(gitRef)
			cloneOpts.SingleBranch = true
			repo, err = git.PlainClone(tmpDir, false, cloneOpts)
			isTag = true
		}

		if err != nil {
			os.RemoveAll(tmpDir)
			return "", "", "", nil, fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
		}

		// If gitRef is a commit SHA, checkout that commit.
		if gitRef != "" {
			if isCommit {
				wt, err := repo.Worktree()
				if err != nil {
					os.RemoveAll(tmpDir)
					return "", "", "", nil, fmt.Errorf("failed to get worktree: %w", err)
				}
				err = wt.Checkout(&git.CheckoutOptions{
					Hash: plumbing.NewHash(gitRef),
				})
				if err != nil {
					os.RemoveAll(tmpDir)
					return "", "", "", nil, fmt.Errorf("failed to checkout commit %s: %w", gitRef, err)
				}
			}
		}

		// Open the repository to obtain the HEAD commit SHA and ref.
		repo, err = git.PlainOpen(tmpDir)
		if err != nil {
			return "", "", "", nil, fmt.Errorf("failed to open cloned repository: %w", err)
		}
		head, err := repo.Head()

		if err != nil {
			return "", "", "", nil, fmt.Errorf("failed to get HEAD commit: %w", err)
		}

		headRef := head.Name().String()
		if isTag {
			tagHead, err := repo.Tag(gitRef)
			if err != nil {
				return "", "", "", nil, fmt.Errorf("failed to get tag %s: %w", gitRef, err)
			}
			headRef = tagHead.Name().String()

		}
		commitSHA := head.Hash().String()

		cleanup := func() {
			os.RemoveAll(tmpDir)
		}
		return tmpDir, commitSHA, headRef, cleanup, nil
	}

	// Otherwise, assume it's a local directory.
	return templateRef, "", "", func() {}, nil
}
