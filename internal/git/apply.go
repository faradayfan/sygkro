package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ApplyDiff applies a diff patch to the repository at repoPath.
// It pipes the provided diff string into the "git apply" command.
func ApplyDiff(repoPath string, diff string) error {
	// Prepare the git apply command.
	cmd := exec.Command("git", "apply")
	cmd.Dir = repoPath
	cmd.Stdin = bytes.NewBufferString(diff)

	// Capture any error output.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command.
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git apply failed: %v: %s", err, stderr.String())
	}
	return nil
}
