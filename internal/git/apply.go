package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ApplyDiff(repoPath string, diff string) error {
	cmd := exec.Command("git", "apply", "-3")
	cmd.Dir = repoPath
	cmd.Stdin = bytes.NewBufferString(diff)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git apply failed: %v: %s", err, stderr.String())
	}
	return nil
}
