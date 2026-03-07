package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/felixtorres/az-dash/internal/azdo"
)

// DiffCommand returns an *exec.Cmd that displays the PR diff.
// Uses delta if available, falls back to less.
func DiffCommand(client *azdo.Client, repoID string, prID int) *exec.Cmd {
	diffContent, err := client.GetPullRequestDiff("", "", repoID, prID)
	if err != nil {
		// If we can't get diff, show error in pager
		diffContent = fmt.Sprintf("Error fetching diff for PR #%d: %v", prID, err)
	}

	tmpFile, tmpErr := os.CreateTemp("", "az-dash-diff-*.diff")
	if tmpErr != nil {
		return exec.Command("echo", "Failed to create temp file")
	}
	tmpFile.WriteString(diffContent)
	tmpFile.Close()

	// Try delta first, but fall back to less if delta rejects the diff.
	if deltaPath, err := exec.LookPath("delta"); err == nil {
		script := fmt.Sprintf("%q < %q || less -R %q", deltaPath, tmpFile.Name(), tmpFile.Name())
		cmd := exec.Command("sh", "-c", script)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}

	cmd := exec.Command("less", "-R", tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
