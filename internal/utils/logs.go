package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/felixtorres/az-dash/internal/azdo"
)

// BuildLogCommand returns an *exec.Cmd that displays build logs in a pager.
func BuildLogCommand(client *azdo.Client, buildID int) *exec.Cmd {
	logs, err := client.GetBuildLogs("", "", buildID)
	if err != nil {
		return echoCmd(fmt.Sprintf("Error fetching logs for build #%d: %v", buildID, err))
	}

	if len(logs) == 0 {
		return echoCmd(fmt.Sprintf("No logs found for build #%d", buildID))
	}

	// Fetch the last (most relevant) log — typically the main job output
	var content strings.Builder
	for _, log := range logs {
		text, err := client.GetBuildLogContent("", "", buildID, log.ID)
		if err != nil {
			content.WriteString(fmt.Sprintf("--- Log #%d (error: %v) ---\n", log.ID, err))
			continue
		}
		content.WriteString(fmt.Sprintf("--- Log #%d (%d lines) ---\n", log.ID, log.LineCount))
		content.WriteString(text)
		content.WriteString("\n")
	}

	tmpFile, err := os.CreateTemp("", "az-dash-logs-*.txt")
	if err != nil {
		return echoCmd("Failed to create temp file")
	}
	tmpFile.WriteString(content.String())
	tmpFile.Close()

	cmd := exec.Command("less", "-R", tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func echoCmd(msg string) *exec.Cmd {
	cmd := exec.Command("echo", msg)
	cmd.Stdout = os.Stdout
	return cmd
}
