package agent

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// InTmux returns true if the current process is running inside a tmux session.
func InTmux() bool {
	return os.Getenv("TMUX") != ""
}

// TmuxAvailable returns true if the tmux binary is on PATH.
func TmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// WindowName returns the tmux window name for a given issue ID.
func WindowName(issueID string) string {
	return "mg-" + issueID
}

// LaunchInTmux opens a detached tmux window for the selected runtime. The
// issue draft is sent literally to the new pane without an Enter key, so the
// user remains in control of when the agent receives it.
func LaunchInTmux(runtime Runtime, projectDir, issueID, draft string) (string, error) {
	windowName := WindowName(issueID)
	agentArgs := InteractiveCommand(runtime, projectDir).Args
	tmuxArgs := newWindowArgs(projectDir, windowName, agentArgs)
	cmd := exec.Command("tmux", tmuxArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux new-window: %w", err)
	}
	paneID := strings.TrimSpace(string(out))

	// Tag the first pane in the agent window so state polling, capture, and
	// kill/select operations can continue to use the stable pane ID.
	_ = exec.Command("tmux", "set-option", "-p", "-t", paneID,
		"@mg_agent", windowName).Run()

	// -l sends the text literally. Crucially, do not send Enter: the draft
	// stays in the agent's composer until the user reviews and submits it.
	// tmux buffers terminal input while the CLI initializes, so this also works
	// for agents that take a moment to render their interactive prompt.
	if draft != "" {
		_ = exec.Command("tmux", "send-keys", "-t", paneID, "-l", "--", draft).Run()
	}

	return paneID, nil
}

// newWindowArgs constructs the tmux invocation used by LaunchInTmux. Keeping
// it separate makes the window (rather than split-pane) contract easy to test.
func newWindowArgs(projectDir, windowName string, agentArgs []string) []string {
	args := []string{
		"new-window",
		"-d", // don't switch focus away from Mardi Gras
		"-c", projectDir,
		"-n", windowName,
		"-P", "-F", "#{pane_id}", // print the first pane ID for tracking
		"--",
	}
	return append(args, agentArgs...)
}

// ListAgentWindows returns a map of issueID -> first-pane ID for all tmux
// agent windows tagged with the @mg_agent option.
func ListAgentWindows() (map[string]string, error) {
	// List all panes with their @mg_agent value and pane_id.
	out, err := exec.Command("tmux", "list-panes", "-a",
		"-F", "#{@mg_agent}\t#{pane_id}").Output()
	if err != nil {
		return nil, fmt.Errorf("tmux list-panes: %w", err)
	}
	return parseAgentPanes(string(out)), nil
}

// parseAgentPanes extracts agent panes from tmux list-panes output.
// Each line is "mg-<issueID>\t%<paneNum>" for tagged panes, or "\t%<paneNum>" for untagged.
func parseAgentPanes(output string) map[string]string {
	agents := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		tag := strings.TrimSpace(parts[0])
		paneID := strings.TrimSpace(parts[1])
		if strings.HasPrefix(tag, "mg-") && paneID != "" {
			issueID := strings.TrimPrefix(tag, "mg-")
			agents[issueID] = paneID
		}
	}
	return agents
}

// KillAgentWindow closes the tmux window for the given issue.
func KillAgentWindow(issueID string) error {
	// Find the first pane ID first; tmux accepts a pane target for kill-window
	// and resolves it to that pane's containing window.
	agents, err := ListAgentWindows()
	if err != nil {
		return err
	}
	paneID, ok := agents[issueID]
	if !ok {
		return fmt.Errorf("no agent window for %s", issueID)
	}
	return exec.Command("tmux", "kill-window", "-t", paneID).Run()
}

// CapturePane captures the last maxLines of output from an agent's tmux pane.
// Returns sanitized lines (ANSI stripped, trailing blanks trimmed).
// Returns nil if the agent pane is not found or capture fails.
func CapturePane(issueID string, maxLines int) []string {
	agents, err := ListAgentWindows()
	if err != nil {
		return nil
	}
	paneID, ok := agents[issueID]
	if !ok {
		return nil
	}
	// capture-pane -p prints to stdout, -S -N starts N lines from the end
	out, err := exec.Command("tmux", "capture-pane",
		"-t", paneID, "-p", "-S", fmt.Sprintf("-%d", maxLines+20)).Output()
	if err != nil {
		return nil
	}
	return sanitizeCaptureOutput(string(out), maxLines)
}

// sanitizeCaptureOutput strips ANSI codes, trims trailing blanks,
// and returns the last maxLines of non-empty content.
func sanitizeCaptureOutput(raw string, maxLines int) []string {
	// Strip ANSI escape sequences
	clean := stripANSI(raw)

	// Split into lines and trim trailing blanks
	allLines := strings.Split(clean, "\n")
	// Remove trailing empty lines
	for len(allLines) > 0 && strings.TrimSpace(allLines[len(allLines)-1]) == "" {
		allLines = allLines[:len(allLines)-1]
	}

	if len(allLines) == 0 {
		return nil
	}

	// Take last maxLines
	if len(allLines) > maxLines {
		allLines = allLines[len(allLines)-maxLines:]
	}

	return allLines
}

// stripANSI removes ANSI escape sequences and stray control characters from a string.
// Uses charmbracelet/x/ansi which handles all sequence types (CSI, OSC, DCS, etc.),
// then strips remaining C0/C1 control bytes that aren't part of escape sequences.
func stripANSI(s string) string {
	s = ansi.Strip(s)
	// Remove control characters (0x00-0x1F except \t, \n, \r) and DEL (0x7F).
	// These can leak from captured tmux output and should not reach the TUI.
	return strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return r
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

// SelectAgentWindow switches focus to the tmux window for the given issue.
func SelectAgentWindow(issueID string) error {
	agents, err := ListAgentWindows()
	if err != nil {
		return err
	}
	paneID, ok := agents[issueID]
	if !ok {
		return fmt.Errorf("no agent window for %s", issueID)
	}
	return exec.Command("tmux", "select-window", "-t", paneID).Run()
}
