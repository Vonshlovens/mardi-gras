# Agent Integration

On Mardi Gras's direct local CLI path, press `a` on any selected issue to open the agent picker. Select a runtime with `j`/`k` and `enter`; `esc` cancels. The picker always shows every supported runtime, including tools that are not installed, so a missing CLI is clear before launch. Gas Town and Gas City retain their existing `a` sling/dispatch workflows.

In tmux, Mardi Gras creates a detached window and literally types a short Beads issue mention into the new agent's composer. It deliberately does **not** press `enter`, so you can review or expand the draft before submitting it. Outside tmux, Mardi Gras suspends into the selected interactive CLI without auto-submitting a prompt.

Mardi Gras supports multiple agent runtimes:

- **[OpenAI Codex](https://github.com/openai/codex)** — `codex --dangerously-bypass-approvals-and-sandbox`
- **[Claude Code](https://claude.com/claude-code)** — `claude --dangerously-skip-permissions`
- **[Cursor CLI](https://cursor.com)** — `cursor-agent --yolo`
- **[GitHub Copilot](https://github.com/features/copilot)** — `copilot --allow-all`

## Choosing a runtime

By default, mg highlights the first detected tool in the picker, preferring `claude`, then `cursor-agent`, `codex`, and `copilot`. Override the initial choice with the `--agent` flag or the `MG_AGENT_RUNTIME` env var:

```bash
mg --agent codex                  # use codex for this session
MG_AGENT_RUNTIME=cursor mg        # same shape via env var
MG_AGENT_RUNTIME=claude mg        # force claude even if you have other tools installed
MG_AGENT_RUNTIME=copilot mg       # highlight GitHub Copilot
```

Accepted values are `claude`, `cursor` (or `cursor-agent`), `codex`, and `copilot` (or `github-copilot`). The override is honored only if the matching binary is on PATH — if you request a runtime that isn't installed, mg falls back to the default detection order rather than failing silently. Unknown values are ignored.

The override applies only to mg's local launch path. When Gas Town is available, the `a` key dispatches through `gt sling` and the runtime is chosen by the Gas Town formula (see [Gas Town docs](https://github.com/steveyegge/gastown)). Gas Town v1.1.0+ has first-class codex support via `gt sling --agent codex`; configure the default in your formula or pass it manually until mg's gt-sling integration follows.

## Codex specifics

Codex normally prompts for sandbox and approval decisions. The picker launches it with `--dangerously-bypass-approvals-and-sandbox`, matching the explicitly permissive mode of the other picker choices. Use this only in a workspace you trust.

A few practical gotchas:

- **First-run auth**: run `codex login` once before mg dispatches a codex agent — codex authenticates with your ChatGPT account and stores credentials in `~/.codex/auth.json`. If you launch mg without logging in first, the agent will exit asking for setup.
- **Project trust**: codex prompts to trust a directory the first time it sees one. For unattended tmux dispatch, either run codex interactively in the project once (it'll remember in `[projects."<path>"]`), or pre-trust by editing `~/.codex/config.toml`.
- **nvm-installed codex**: if you installed codex via `npm install -g @openai/codex` under nvm, mg inherits PATH from the shell that launches it — make sure the right Node version is active (e.g. `nvm use 22`) before starting mg, or codex won't resolve. Separately, openai/codex#20906 documents a sandbox-PATH bug specific to nvm installs that can be worked around with `--add-dir $NVM_PATH`; consider a standalone or Homebrew install if you hit it.
- **AGENTS.md**: codex automatically reads `AGENTS.md` from the project root (and merges with `~/.codex/AGENTS.md`). Keep build/test/convention guidance there; the picker only pre-fills a short issue mention rather than sending a full generated task prompt.
- **Beads integration on Homebrew bd v1.0.4**: bd's `codex-hook` subcommand (which injects Beads context into codex sessions) shipped on bd main but is missing from the v1.0.4 release (steveyegge/beads#3924). Until v1.0.5 cuts, codex with bd-installed hooks will log "unknown command 'codex-hook'" once per prompt. This is cosmetic for the agent's work but worth knowing.

### Resuming a prior Codex session

When codex is the active runtime inside tmux, the command palette (`:` or `Ctrl+K`) shows a **"Resume last Codex session"** action. It launches `codex resume --last` in a new tmux split rooted at the project directory. The action is gated on a rollout file actually existing under `~/.codex/sessions/YYYY/MM/DD/*.jsonl` so a never-launched codex doesn't surface as a confusing empty pane.

### Gas Town routing for Codex

When Gas Town is on PATH and `MG_AGENT_RUNTIME=codex` (or `--agent codex`) is active, mg's sling commands pass `--agent codex` to `gt sling`, so the agent preference propagates from mg into Gas Town. Requires gt v1.1.0+ (earlier versions reject the `--agent` flag). For `claude` / `cursor-agent`, mg continues to let gt pick its default agent — the v0.19.0 behavior is unchanged.

## Tmux-native dispatch (multi-agent)

When running inside tmux, agents launch in **new tmux windows** instead of suspending the TUI. This means:

- The parade stays visible while agents work
- Multiple agents can run simultaneously on different issues
- The selected CLI starts in its permissive/yolo-equivalent mode and receives an unsent one-line issue draft
- Active agents show a `⚡` badge next to their issue in the parade
- The header displays the total active agent count
- Press `a` on an issue with an active agent to **switch** to its tmux window
- Press `A` to **kill** the active agent on the selected issue
- Agent status is polled automatically alongside the file watcher

## Fallback (non-tmux)

Outside tmux, the TUI suspends while the agent runs (using BubbleTea's `tea.ExecProcess`), giving the agent the full terminal. When you exit the session, Mardi Gras resumes and reloads data to pick up any changes.

## Requirements

- Supports `codex`, `claude`, `cursor-agent`, and `copilot` on your `PATH`
- The command palette opens the same runtime picker as `a`
- If no agent runtime is found, the picker still opens and identifies each unavailable choice
- Tmux dispatch requires both the `TMUX` env var and `tmux` binary on PATH
- The tmux draft is deliberately left unsubmitted; you decide when to send it
