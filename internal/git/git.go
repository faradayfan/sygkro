package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runCommand(execDir string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd.Dir = execDir

	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()

	if err != nil {
		return stdout.String(), fmt.Errorf("command failed: %w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func gitCheckout(repoPath string, commitish string) error {
	stdOut, err := runCommand(repoPath, "git", "checkout", commitish)
	if err != nil {
		return fmt.Errorf("git checkout failed: %w: %s", err, stdOut)
	}

	return nil
}

func gitDiff(current string, ideal string) (string, error) {
	// Prepare the git diff command.
	args := []string{
		"--no-pager",
		"-c",
		"diff.noprefix=",
		"diff",
		"--no-index",
		"--relative",
		"--binary",
		"--src-prefix=upstream-template-old/",
		"--dst-prefix=upstream-template-new/",
		"--no-ext-diff",
		"--no-color",
		current,
		ideal,
	}
	stdOut, _ := runCommand(current, "git", args...)

	oldProject := fmt.Sprintf("upstream-template-old%s", current)
	newProject := fmt.Sprintf("upstream-template-new%s", ideal)

	fixedDiff := strings.ReplaceAll(stdOut, oldProject, "upstream-template-old")
	fixedDiff = strings.ReplaceAll(fixedDiff, newProject, "upstream-template-new")

	return fixedDiff, nil
}
