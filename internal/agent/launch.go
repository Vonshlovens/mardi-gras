// Package agent handles AI agent runtime detection and launch. It supports
// Codex, Claude Code, Cursor CLI, and GitHub Copilot, with tmux window
// dispatch for multi-agent sessions.
package agent

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/matt-wright86/mardi-gras/internal/data"
)

// Runtime identifies which AI agent binary to use.
type Runtime string

const (
	RuntimeClaude  Runtime = "claude"
	RuntimeCursor  Runtime = "cursor-agent"
	RuntimeCodex   Runtime = "codex"
	RuntimeCopilot Runtime = "copilot"
)

// SupportedRuntimes returns the choices shown in Mardi Gras's agent picker.
// Keep this list deliberate rather than deriving it from PATH: unavailable
// choices remain visible in the picker so users can see what needs installing.
func SupportedRuntimes() []Runtime {
	return []Runtime{
		RuntimeCodex,
		RuntimeClaude,
		RuntimeCursor,
		RuntimeCopilot,
	}
}

// Binary returns the executable name associated with a runtime.
func (r Runtime) Binary() string {
	switch r {
	case RuntimeClaude:
		return "claude"
	case RuntimeCursor:
		return "cursor-agent"
	case RuntimeCodex:
		return "codex"
	case RuntimeCopilot:
		return "copilot"
	default:
		return ""
	}
}

// Installed reports whether this runtime's executable can be found on PATH.
func (r Runtime) Installed() bool {
	binary := r.Binary()
	if binary == "" {
		return false
	}
	_, err := exec.LookPath(binary)
	return err == nil
}

// DetectRuntime returns the agent runtime to launch.
//
// If MG_AGENT_RUNTIME is set to "claude", "cursor" (or "cursor-agent"),
// "codex", or "copilot" and the corresponding binary is on PATH, that
// runtime wins. Unknown values or missing binaries fall through to the default
// detection order: claude, then cursor-agent, then codex, then copilot.
func DetectRuntime() Runtime {
	if pref := strings.ToLower(strings.TrimSpace(os.Getenv("MG_AGENT_RUNTIME"))); pref != "" {
		switch pref {
		case "claude":
			if RuntimeClaude.Installed() {
				return RuntimeClaude
			}
		case "cursor", "cursor-agent":
			if RuntimeCursor.Installed() {
				return RuntimeCursor
			}
		case "codex":
			if RuntimeCodex.Installed() {
				return RuntimeCodex
			}
		case "copilot", "github-copilot", "github copilot":
			if RuntimeCopilot.Installed() {
				return RuntimeCopilot
			}
		}
	}
	for _, runtime := range []Runtime{RuntimeClaude, RuntimeCursor, RuntimeCodex, RuntimeCopilot} {
		if runtime.Installed() {
			return runtime
		}
	}
	return ""
}

// Available returns true if any supported agent CLI is on PATH.
func Available() bool {
	return DetectRuntime() != ""
}

// RuntimeLabel returns a display name for the runtime.
func (r Runtime) RuntimeLabel() string {
	switch r {
	case RuntimeClaude:
		return "Claude Code"
	case RuntimeCursor:
		return "Cursor CLI"
	case RuntimeCodex:
		return "Codex"
	case RuntimeCopilot:
		return "GitHub Copilot"
	default:
		return "unknown"
	}
}

// PermissionModeLabel describes the non-interactive approval posture used
// when starting an interactive session through the picker.
func (r Runtime) PermissionModeLabel() string {
	switch r {
	case RuntimeCodex:
		return "bypass approvals + sandbox"
	case RuntimeClaude:
		return "skip permissions"
	case RuntimeCursor:
		return "yolo mode"
	case RuntimeCopilot:
		return "allow all"
	default:
		return ""
	}
}

// BuildPrompt composes the initial prompt for a Claude Code session
// given a selected issue and its evaluated dependencies.
func BuildPrompt(issue data.Issue, deps data.DepEval, issueMap map[string]*data.Issue) string {
	var b strings.Builder

	b.WriteString("Work on this Beads issue:\n\n")
	fmt.Fprintf(&b, "## %s: %s\n\n", issue.ID, issue.Title)

	fmt.Fprintf(&b, "Status: %s | Type: %s | Priority: %s\n",
		issue.Status, issue.IssueType, data.PriorityLabel(issue.Priority))
	if issue.Owner != "" {
		fmt.Fprintf(&b, "Owner: %s\n", issue.Owner)
	}
	if issue.Assignee != "" {
		fmt.Fprintf(&b, "Assignee: %s\n", issue.Assignee)
	}

	if issue.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", issue.Description)
	}

	if issue.Notes != "" {
		fmt.Fprintf(&b, "\n### Notes\n%s\n", issue.Notes)
	}

	if issue.AcceptanceCriteria != "" {
		fmt.Fprintf(&b, "\n### Acceptance Criteria\n%s\n", issue.AcceptanceCriteria)
	}

	if len(deps.Edges) > 0 {
		b.WriteString("\n### Dependencies\n")
		for _, edge := range deps.Edges {
			switch edge.Status {
			case data.DepBlocking:
				if dep, ok := issueMap[edge.DependsOnID]; ok {
					fmt.Fprintf(&b, "- Blocked by: %s (%s) -- %s\n",
						edge.DependsOnID, dep.Title, dep.Status)
				}
			case data.DepMissing:
				fmt.Fprintf(&b, "- Missing: %s (not found)\n", edge.DependsOnID)
			case data.DepResolved:
				if dep, ok := issueMap[edge.DependsOnID]; ok {
					fmt.Fprintf(&b, "- Resolved: %s (%s) -- closed\n",
						edge.DependsOnID, dep.Title)
				}
			case data.DepNonBlocking:
				if dep, ok := issueMap[edge.DependsOnID]; ok {
					fmt.Fprintf(&b, "- Related: %s (%s) -- %s\n",
						edge.DependsOnID, dep.Title, edge.Type)
				}
			}
		}
	}

	fmt.Fprintf(&b, "\n---\nWhen you begin work, run: bd update %s --status=in_progress\n", issue.ID)
	fmt.Fprintf(&b, "When finished, run: bd close %s\n", issue.ID)
	b.WriteString("\nIf this task is complex enough to benefit from parallel work, consider using agent teams to spawn teammates for independent subtasks.")

	return b.String()
}

// BriefPrompt returns the one-line draft that Mardi Gras places in a new
// tmux agent window. It intentionally does not include a trailing newline:
// the user should review or extend it and press Enter themselves.
func BriefPrompt(issue data.Issue) string {
	title := strings.Join(strings.Fields(issue.Title), " ")
	const maxTitleRunes = 120
	if len([]rune(title)) > maxTitleRunes {
		title = string([]rune(title)[:maxTitleRunes-1]) + "…"
	}
	if title == "" {
		return fmt.Sprintf("Work on Beads issue %s", issue.ID)
	}
	return fmt.Sprintf("Work on Beads issue %s: %s", issue.ID, title)
}

// InteractiveCommand returns an agent command for a user-driven interactive
// session. The selected runtime starts with its least-interrupting permission
// mode, but no prompt argument is passed: tmux callers can prefill a draft
// without submitting it, and non-tmux callers retain a blank interactive box.
func InteractiveCommand(runtime Runtime, projectDir string) *exec.Cmd {
	if runtime == "" {
		runtime = DetectRuntime()
	}

	var c *exec.Cmd
	switch runtime {
	case RuntimeCodex:
		c = exec.Command("codex", "--dangerously-bypass-approvals-and-sandbox", "-C", projectDir)
	case RuntimeClaude:
		c = exec.Command("claude", "--dangerously-skip-permissions")
	case RuntimeCursor:
		c = exec.Command("cursor-agent", "--yolo")
	case RuntimeCopilot:
		c = exec.Command("copilot", "--allow-all")
	default:
		// The caller checks Installed before launch. Preserve an executable
		// command here so an unexpected runtime produces a normal process
		// error rather than a nil dereference in Bubble Tea.
		c = exec.Command(runtime.Binary())
	}
	c.Dir = projectDir
	return c
}

// Command returns an *exec.Cmd that launches the detected agent runtime
// with the given prompt, working directory set to projectDir.
//
// Codex defaults to sandboxed execution with interactive approval prompts,
// which would block unattended agent sessions. We pass --sandbox workspace-write
// and -a on-request to match the zero-friction posture Claude and Cursor have
// out of the box. Power users can override via codex profiles or
// MG_AGENT_RUNTIME=codex combined with a custom shell alias.
func Command(prompt, projectDir string) *exec.Cmd {
	rt := DetectRuntime()
	var c *exec.Cmd
	switch rt {
	case RuntimeCursor:
		c = exec.Command("cursor-agent", "-f", "-p", prompt)
	case RuntimeCodex:
		c = exec.Command("codex",
			"--sandbox", "workspace-write",
			"-a", "on-request",
			"-C", projectDir,
			prompt)
	case RuntimeCopilot:
		c = exec.Command("copilot", "-i", prompt, "--allow-all")
	default: // Claude Code
		c = exec.Command("claude", prompt)
	}
	c.Dir = projectDir
	return c
}
