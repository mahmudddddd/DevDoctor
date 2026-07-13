# Terminal UX research

This note records public terminal-interface research used to guide DevDoctor. It is not a specification for copying another product. DevDoctor remains an original, deterministic, beginner-focused diagnostic tool rather than a conversational coding interface.

Research reviewed on 2026-07-13:

- Public OpenAI Codex CLI source and protocol documentation
- Publicly documented Claude Code terminal behavior from Anthropic

## Public Codex CLI observations

The public Codex CLI separates interactive and programmatic workflows. Its TUI exposes discoverable commands, persistent session concepts, explicit approvals, cancellation, and status information. Its app-server protocol models durable work as thread → turn → item, with started, progress, completed, failed, and interrupted lifecycle events.

Relevant public sources:

- [Codex TUI tooltips](https://github.com/openai/codex/blob/main/codex-rs/tui/tooltips.txt)
- [Codex app-server protocol](https://github.com/openai/codex/blob/main/codex-rs/app-server/README.md)
- [Codex non-interactive execution](https://github.com/openai/codex/blob/main/docs/exec.md)
- [Codex configuration schema](https://github.com/openai/codex/blob/main/codex-rs/core/config.schema.json)
- [Running Codex safely](https://openai.com/index/running-codex-safely/)

Useful patterns:

- Interactive and non-interactive modes are distinct rather than inferred late.
- Approval requests are associated with the exact operation that needs permission.
- Progress events and final state are separate; cancellation is an explicit state transition.
- A command palette and visible shortcuts improve discoverability.
- Alternate-screen behavior can be disabled so normal terminal scrollback remains available.
- Machine-oriented execution has stable lifecycle and completion states.

Public issue reports also demonstrate pitfalls worth avoiding: indefinite “working” states, completion events lost behind progress traffic, rendering corruption during terminal resize, unclear cancellation, and keyboard shortcuts that conflict with IMEs or terminal conventions. These reports are failure examples, not normative Codex behavior.

## Public Claude Code observations

Claude Code publicly documents separate interactive (`claude`) and print/non-interactive (`claude -p`) modes. Non-interactive output can be plain text, JSON, or streaming JSON. Permission behavior uses explicit allow, ask, and deny outcomes; actions that require unavailable approval do not wait forever in non-interactive mode.

Relevant official sources:

- [Claude Code quickstart](https://code.claude.com/docs/en/quickstart)
- [Interactive mode](https://code.claude.com/docs/en/interactive-mode)
- [CLI reference](https://code.claude.com/docs/en/cli-reference)
- [Run Claude Code programmatically](https://code.claude.com/docs/en/headless)
- [Configure permissions](https://code.claude.com/docs/en/permissions)
- [Permission modes](https://code.claude.com/docs/en/permission-modes)
- [Security](https://code.claude.com/docs/en/security)
- [Sandboxing](https://code.claude.com/docs/en/sandboxing)
- [Status line](https://code.claude.com/docs/en/statusline)

Useful patterns:

- Read-only behavior is the conservative default; edits, commands, and network activity have stronger permission boundaries.
- The current trust/permission state is visible rather than hidden in configuration.
- Human-readable and machine-readable output are both first-class.
- Interrupt, help, history, command discovery, and transcript visibility are documented.
- The working directory is a visible trust boundary.
- Automation fails closed when approval cannot be requested.
- Plain output and native terminal scrollback remain important fallbacks.

## DevDoctor-specific UX decisions

DevDoctor should borrow principles, not product identity or conversational structure.

### 1. Guided by default, scriptable by design

- `devdoctor` opens a short guided menu only when stdin and stdout are terminals.
- `devdoctor diagnose` is the explicit automation path.
- Redirected or CI execution never waits for a prompt.
- Text is the human default; versioned JSON is the stable machine format.

### 2. Show scope before action

Every run should make the selected project root visible. Future gated checks should show:

- Exact command or operation
- Working directory
- Why the check is needed
- Whether it can mutate files, use the network, or start a service
- Approval scope: once, this check, or this run

A generic spinner must never hide an approval wait.

### 3. Use diagnostic phases, not a chat transcript

DevDoctor's durable hierarchy should be:

```text
Diagnostic run
└── Phase or check
    └── Evidence, observation, finding, or skipped reason
```

Transient progress can be redrawn, but completed evidence and final findings must remain printable and recoverable. A final report must make sense without the animation or intermediate screen.

### 4. Prefer plain explanations over dense status chrome

Beginner-facing output should answer:

1. What DevDoctor inspected
2. What it detected
3. What was skipped and why
4. What failed
5. What evidence supports the conclusion
6. What safe next step is available

Status labels should use words such as `checking`, `waiting for approval`, `passed`, `warning`, `failed`, `skipped`, `cancelled`, and `timed out`. Color and symbols may reinforce these labels but never replace them.

### 5. Keep scrollback and copyability

- Avoid requiring an alternate terminal screen for primary results.
- Keep the final report in normal terminal scrollback.
- Make text output easy to copy into an issue or support request.
- Bound high-volume logs and show explicit truncation markers.
- Add verbose/transcript views later without making them necessary for the basic explanation.

### 6. Treat interruption as a normal outcome

Future long-running checks should support Ctrl+C, stop the full process tree, preserve completed findings, and report `cancelled` rather than a generic failure. Completion, failure, timeout, and cancellation events must have higher delivery priority than spinner updates.

### 7. Design for terminal diversity

- Respect `NO_COLOR`.
- Provide ASCII/plain-text fallbacks.
- Avoid relying on mouse interaction, animation, or Ctrl-only shortcuts.
- Test PowerShell, Windows Terminal, macOS terminals, Linux terminals, tmux, narrow widths, redirected logs, and multilingual input methods.
- Keep labels understandable to screen-reader users even when styling is removed.

### 8. Keep agent UX subordinate to deterministic diagnosis

Optional coding-agent assistance should appear only after local deterministic findings. The interface should identify the selected provider, show the exact redacted bundle, and distinguish agent hypotheses from DevDoctor evidence. Agent suggestions cannot change approval, path, privacy, or command policy.

## Phase 1 application

The current Phase 1 implementation intentionally uses a small Huh menu, normal terminal output, clear project-root display, text/JSON formats, no project-script execution, and a non-interactive fail-closed path. It does not implement long-running progress, command approvals, alternate-screen dashboards, conversational sessions, or agent UX; those belong to later approved phases.
