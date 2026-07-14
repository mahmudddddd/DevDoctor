# DebugDoc project instructions

DebugDoc is a local-first, deterministic, beginner-friendly diagnostic CLI. Preserve its safety, privacy, cross-platform, and non-interactive contracts.

## Required reading for terminal UI work

Before modifying any interactive terminal UI, every agent session must read:

1. `docs/tui-reference.md`
2. `docs/tui-design-system.md`

Also read the relevant existing architecture, privacy, UX, and contributing documentation before changing behavior. For interactive TUI design, `docs/tui-design-system.md` is the permanent specification; where older UX research describes a previous interaction direction, the design system governs the approved redesigned shell.

A reusable read-only research/review agent is defined at `.claude/agents/tui-researcher.md`.

## Permanent terminal UI constraints

- DebugDoc uses an original interface implemented with Bubble Tea and Lip Gloss.
- OpenAI Codex and Claude Code are research references, not templates to copy.
- Do not copy branding, logos, branded assets, exact wording, exact command catalogs, exact color palettes, distinctive layouts, or large implementation blocks.
- The root interactive command must use a full-screen, alternate-screen, in-place TUI.
- The full report must never be printed before the interactive shell starts.
- The interface must support at least 80x24 and must be tested at 80x24, 100x30, and 120x40.
- Keep one active primary view, a persistent composer, an allowlisted slash-command palette, a scrollable viewport, a compact project/status header, and a contextual footer.
- Interaction must be keyboard-first. Mouse support may be optional but never required.
- Running and lifecycle updates occur in place rather than as appended output.
- Running, success, warning, failure, skipped, cancelled, and timed-out states must be explicit in text.
- Honor `NO_COLOR`; provide ASCII/flat fallbacks; never rely on color, borders, symbols, or animation alone.
- Arbitrary shell input must never be accepted. Composer text can resolve only to registered, typed DebugDoc actions.

## Compatibility boundaries

- Non-interactive CLI behavior must remain stable.
- Redirected input/output must never open or wait on an interactive prompt.
- Keep `debugdoc diagnose --path <directory> --format text|json` behavior stable unless separately approved.
- Report schema `1.0` must remain unchanged unless a separate schema migration is explicitly approved.
- UI work must not begin Phase 3 deterministic diagnostic functionality, add diagnostic rules, select project commands, or expand production capability by accident.
- Keep terminal rendering in the presentation layer. Domain models, detectors, rules, privacy policy, consent, and runner packages must not depend on Bubble Tea or Lip Gloss.

## Safety boundaries

Do not weaken or bypass existing:

- Project-root containment and allowlisted metadata reads.
- Treatment of project files, logs, symlinks, scripts, terminal output, and agent output as untrusted.
- Structured executable and argument boundaries; never construct shell command strings.
- Exact consent and safe default denial.
- Canonical path and filesystem identity revalidation before process start.
- Minimal environment policy and prohibition on exposing environment values.
- Independent bounded stdout/stderr capture and explicit truncation metadata.
- Timeout and cancellation semantics.
- Unix process-group and Windows Job Object process-tree cleanup.
- Non-interactive approval-unavailable fail-closed behavior.
- Terminal control-character sanitization and redaction requirements.

A visual redesign must call existing safety and application boundaries rather than recreate, shortcut, or silently broaden them.

## Scope and change control

- Keep TUI redesign work separate from Phase 3 and report-schema work.
- Do not change production behavior during research or design-documentation tasks.
- Architectural changes require an explicit proposal and approval.
- Preserve useful existing instructions in `CONTRIBUTING.md`, `docs/architecture.md`, `docs/privacy.md`, `docs/ux-research.md`, `docs/rule-authoring.md`, and `docs/agent-integrations.md`.
- Do not commit or push unless the user explicitly asks.

## Testing expectations for future TUI implementation

Follow `docs/tui-design-system.md`. At minimum, future implementation must include:

- Pure Bubble Tea update/state-transition tests.
- Render snapshots and overflow assertions at 80x24, 100x30, and 120x40.
- Below-minimum, long-path, Unicode, `NO_COLOR`, ANSI-limited, and ASCII/flat coverage.
- Viewport bottom-follow, manual-scroll, jump-to-latest, and resize-anchor tests.
- Keymap conflict, modal priority, focus restoration, and footer-hint tests.
- Alternate-screen entry/restoration and PTY smoke tests where feasible.
- Compatibility tests proving non-interactive behavior and report schema `1.0` remain stable.
- Existing safety, consent, timeout, cancellation, output-bound, identity-revalidation, cleanup, and fail-closed tests.
