# Terminal UI reference research

> Research date: 2026-07-13
>
> Purpose: record public terminal-interface research that can inform an original DebugDoc TUI. This is an architectural reference, not a template or permission to copy another product's implementation, wording, branding, assets, or visual identity.

## Research boundaries

The research pass examined:

- The official, public [OpenAI Codex repository](https://github.com/openai/codex), with emphasis on the current Rust terminal UI under `codex-rs/tui`.
- Official Claude Code documentation and behavior exposed by the locally installed `claude --help` command.
- DebugDoc's existing CLI, report renderer, safety model, and terminal behavior.

Claude Code source code was **not** inspected. No official public Claude Code source repository was found or used. Claude observations below are limited to official documentation and locally observable command-line behavior.

Upstream `main` branches change. Exact Codex paths below describe the repository as observed on the research date and should be revalidated before relying on a symbol or module.

## Sources examined

### OpenAI Codex

- [Repository](https://github.com/openai/codex)
- [TUI crate](https://github.com/openai/codex/tree/main/codex-rs/tui)
- [TUI source directory](https://github.com/openai/codex/tree/main/codex-rs/tui/src)
- [TUI tests](https://github.com/openai/codex/tree/main/codex-rs/tui/tests)
- [TUI style guidance](https://github.com/openai/codex/blob/main/codex-rs/tui/styles.md)
- [Repository license](https://github.com/openai/codex/blob/main/LICENSE)
- [Repository notice](https://github.com/openai/codex/blob/main/NOTICE)

### Claude Code

- [Overview](https://code.claude.com/docs/en/overview)
- [Interactive mode](https://code.claude.com/docs/en/interactive-mode)
- [Fullscreen mode](https://code.claude.com/docs/en/fullscreen)
- [Terminal configuration](https://code.claude.com/docs/en/terminal-config)
- [Commands](https://code.claude.com/docs/en/commands)
- [CLI reference](https://code.claude.com/docs/en/cli-reference)
- [Permissions](https://code.claude.com/docs/en/permissions)
- [Permission modes](https://code.claude.com/docs/en/permission-modes)
- [Security](https://code.claude.com/docs/en/security)
- Locally observable `claude --help` output on 2026-07-13

### DebugDoc

- `internal/cli/root.go`
- `internal/cli/interactive.go`
- `internal/cli/terminal.go`
- `internal/cli/consent.go`
- `internal/report/render.go`
- `internal/model/project.go`
- `docs/architecture.md`
- `docs/privacy.md`
- `docs/ux-research.md`
- `docs/agent-integrations.md`
- `go.mod`
- `LICENSE`

## Exact Codex source paths examined

### Application shell, event loop, and rendering lifecycle

- [`codex-rs/tui/src/app.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/app.rs)
  - `App`, `App::run`, `handle_event`, `handle_tui_event`, `handle_draw_pre_render`, `render_chat_widget_frame`, `handle_key_event`, and overlay event routing.
  - Multiplexes terminal, application, active-thread, and server events and coordinates redraw, resize reflow, cursor placement, and exit state.
- [`codex-rs/tui/src/chatwidget.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/chatwidget.rs)
  - `ChatWidget`, `TranscriptState`, `HistoryCell`, stream controllers, `pre_draw_tick`, `request_redraw`, history mutation, active-cell finalization, terminal resize handling, and task/status state.
- [`codex-rs/tui/src/tui.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/tui.rs)
  - Terminal initialization and restoration, raw mode, event translation, drawing, resize reflow, inline viewport behavior, alternate-screen entry/exit, focus reporting, bracketed paste, panic cleanup, and `TuiEvent`.

### Composer and input area

- [`codex-rs/tui/src/bottom_pane/chat_composer.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/chat_composer.rs)
  - `ChatComposer`, draft/attachment/popup/footer state, input results, key and paste handling, focus, cursor location, layout, history, and popup-specific routing.
- [`codex-rs/tui/src/bottom_pane/textarea.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/textarea.rs)
- [`codex-rs/tui/src/bottom_pane/chat_composer_history.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/chat_composer_history.rs)
- [`codex-rs/tui/src/bottom_pane/paste_burst.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/paste_burst.rs)
- [`codex-rs/tui/src/bottom_pane/pending_input_preview.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/pending_input_preview.rs)
- [`codex-rs/tui/src/bottom_pane/prompt_args.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/prompt_args.rs)

### Slash commands and command palette

- [`codex-rs/tui/src/slash_command.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/slash_command.rs)
  - `SlashCommand`, descriptions, command names, visibility and availability rules, inline argument support, aliases, presentation order, and tests.
- [`codex-rs/tui/src/bottom_pane/command_popup.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/command_popup.rs)
- [`codex-rs/tui/src/bottom_pane/slash_commands.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/slash_commands.rs)
- [`codex-rs/tui/src/bottom_pane/selection_popup_common.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/selection_popup_common.rs)
- [`codex-rs/tui/src/bottom_pane/selection_tabs.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/selection_tabs.rs)
- [`codex-rs/tui/src/bottom_pane/list_selection_view.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/list_selection_view.rs)
- [`codex-rs/tui/src/bottom_pane/multi_select_picker.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/multi_select_picker.rs)

### Viewport, scrolling, resize, and reflow

- [`codex-rs/tui/src/pager_overlay.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/pager_overlay.rs)
  - `Overlay`, `TranscriptOverlay`, `StaticOverlay`, and `PagerView`; scroll offset, height caches, bottom-follow behavior, manual-position preservation, paging, chunk visibility, percentage presentation, and width/revision-aware live-tail caching.
- [`codex-rs/tui/src/transcript_reflow.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/transcript_reflow.rs)
- [`codex-rs/tui/src/thread_transcript.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/thread_transcript.rs)
- [`codex-rs/tui/src/resize_reflow_cap.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/resize_reflow_cap.rs)
- [`codex-rs/tui/src/live_wrap.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/live_wrap.rs)

### Overlays, menus, approval, header, and status surfaces

- [`codex-rs/tui/src/bottom_pane/approval_overlay.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/approval_overlay.rs)
- [`codex-rs/tui/src/bottom_pane/pending_thread_approvals.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/pending_thread_approvals.rs)
- [`codex-rs/tui/src/bottom_pane/action_required_title.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/action_required_title.rs)
- [`codex-rs/tui/src/bottom_pane/custom_prompt_view.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/custom_prompt_view.rs)
- [`codex-rs/tui/src/bottom_pane/mcp_server_elicitation.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/mcp_server_elicitation.rs)
- [`codex-rs/tui/src/bottom_pane/footer.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/bottom_pane/footer.rs)
- [`codex-rs/tui/src/unified_exec_footer.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/unified_exec_footer.rs)
- [`codex-rs/tui/src/status_line_setup.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/status_line_setup.rs)
- [`codex-rs/tui/src/status_line_style.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/status_line_style.rs)
- [`codex-rs/tui/src/status_surface_preview.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/status_surface_preview.rs)
- [`codex-rs/tui/src/title_setup.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/title_setup.rs)
- [`codex-rs/tui/src/theme_picker.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/theme_picker.rs)
- [`codex-rs/tui/src/resume_picker.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/resume_picker.rs)
- [`codex-rs/tui/src/update_prompt.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/update_prompt.rs)
- [`codex-rs/tui/src/selection_list.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/selection_list.rs)

### Keyboard routing and focus management

- [`codex-rs/tui/src/keymap.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/keymap.rs)
  - Runtime keymap contexts, local/global resolution, keybinding parsing, primary binding selection, conflict and shadow validation, reserved keys, and modal precedence.
- [`codex-rs/tui/src/keymap_setup.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/keymap_setup.rs)
- [`codex-rs/tui/src/key_hint.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/key_hint.rs)
- [`codex-rs/tui/src/motion.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/motion.rs)

### Styling, color, and tests

- [`codex-rs/tui/styles.md`](https://github.com/openai/codex/blob/main/codex-rs/tui/styles.md)
- [`codex-rs/tui/src/color.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/src/color.rs)
- [`codex-rs/tui/Cargo.toml`](https://github.com/openai/codex/blob/main/codex-rs/tui/Cargo.toml)
- [`codex-rs/tui/tests/all.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/tests/all.rs)
- [`codex-rs/tui/tests/manager_dependency_regression.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/tests/manager_dependency_regression.rs)
- [`codex-rs/tui/tests/test_backend.rs`](https://github.com/openai/codex/blob/main/codex-rs/tui/tests/test_backend.rs)

## Architectural observations

### Separate state from terminal mechanics

Codex separates terminal setup/event translation, application coordination, content state, input composition, overlays, keymaps, and paging. The useful principle is the boundary, not the exact module graph. DebugDoc should keep the diagnostic/report domain independent of wrapping, styles, terminal capabilities, and Bubble Tea messages.

### Model committed content and active work separately

Codex distinguishes committed transcript history from the active streaming cell. DebugDoc has no need for a chat transcript, but it does need an equivalent distinction between stable completed report sections and transient running progress. A redraw or resize must not mutate the report or convert transient text into durable evidence.

### Preserve the user's scroll intent

A strong viewport invariant is visible in the pager design:

- If the user is at the bottom, new content keeps the view pinned to the bottom.
- If the user has scrolled up, updates preserve that position.
- Returning to live follow is explicit and discoverable.

DebugDoc should implement the same behavior for running checks and result updates.

### Route modal input before global input

Active overlays and local views receive key events before the application shell. Escape closes the topmost dismissible surface before it means cancel or quit. This prevents a command palette, consent screen, or error detail view from leaking keystrokes into the composer or global handler.

### Reflow logical content, not previously rendered strings

Codex retains width-aware logical renderables and recalculates wrapped height on resize. DebugDoc should retain report blocks or rows as structured data and render them at the current width. Re-wrapping strings that already contain padding or ANSI sequences produces drift, clipping, and incorrect scroll offsets.

### Centralize and validate keyboard behavior

Codex uses context-aware keymaps and validates conflicts and reserved bindings. DebugDoc's initial keymap can be much smaller, but every displayed hint should come from the same binding definitions used by event routing. Modal precedence and conflicts should be testable.

### Coalesce redraws and clean up terminal modes

The render lifecycle batches state changes into a draw, performs pre-draw updates, renders once, positions the cursor, and restores terminal state on normal exit or panic. Bubble Tea supplies much of this lifecycle, but DebugDoc still needs explicit cancellation, terminal restoration tests, and nonblocking commands.

## Codex test strategy observations

The TUI combines inline unit tests, rendering/snapshot-oriented tests, a test terminal backend, and regression tests. Particularly reusable coverage areas are:

- Scroll pinning and manual-scroll preservation.
- Wrapped-height and resize behavior.
- Live-tail/render cache invalidation.
- Key parsing, fallback, duplicate bindings, reserved keys, and shadowing.
- Popup selection and dismissal behavior.
- Status and active-cell state accessors for deterministic assertions.

DebugDoc should test its pure `Update` transitions and `View` output independently, then add a small PTY/terminal smoke-test layer for alternate-screen entry, resize, cancellation, and restoration.

## Accessibility and no-color observations

Codex style guidance recommends terminal-default foreground for ordinary text and textual or symbolic status in addition to color. `color.rs` includes terminal color and luminance utilities. The inspected source did not establish an explicit `NO_COLOR` path or a documented user-facing no-color flag. Do not infer one.

DebugDoc must go further:

- Honor `NO_COLOR` explicitly.
- Never use hue as the only state signal.
- Offer plain words and ASCII-safe symbols.
- Avoid motion as the only running indicator.
- Keep focus, selected item, errors, and consent choices understandable after ANSI styling is removed.
- Keep the non-interactive text and JSON modes fully usable without TUI capabilities.

## Claude Code public behavior observations

These are public or locally observable behaviors, not implementation claims.

- Interactive and print/non-interactive modes are distinct. Local help exposes `-p`/`--print` for non-interactive output.
- Official fullscreen documentation describes alternate-screen rendering with a fixed input area, keyboard and optional mouse scrolling, paused auto-follow after scrolling up, a jump-to-bottom affordance, transcript search, and an escape route back to native scrollback.
- Official interactive documentation describes keyboard-first navigation, slash-command discovery, history search, redraw, interrupt, transcript viewing, multiline input, and modal dismissal.
- The command palette filters as the user types `/` followed by text.
- Permission/trust behavior is explicit. The working directory and permission mode are meaningful trust boundaries; non-interactive operation cannot depend on an unavailable prompt.
- Local help documents a screen-reader-oriented mode that uses flatter text and removes decorative borders or animations.
- Official documentation exposes light, dark, daltonized, and ANSI-oriented theme behavior, but no dedicated `--no-color` flag was found in the sources examined.
- Public status surfaces expose permission state and attention needs. The documentation does not define a universal internal running/success/warning/failure state machine, so DebugDoc should define its own.

### Claude research limitations

- No Claude Code source code was inspected.
- No private or reverse-engineered implementation details were used.
- Documentation may describe preview functionality that changes independently of the installed CLI.
- Observed help text proves only the public interface present in the installed build, not its internal architecture.
- No user screenshots were supplied for this research pass.

## Reusable interaction patterns for DebugDoc

1. **Full-screen root shell:** launch the interactive root command directly into one alternate-screen application instead of printing a report and then opening controls.
2. **Persistent composer:** retain one compact input/action area at the bottom of the shell across views.
3. **Allowlisted slash palette:** typing `/` opens a filtered list of known DebugDoc actions. Slash commands resolve to typed internal actions; they are never shell strings.
4. **One active primary view:** show home, project selection, running diagnostics, results, details, consent, or error as the single primary view. Use overlays only for short choices or help.
5. **Scrollable viewport:** preserve bottom-follow and manual-scroll intent, expose position when not at the bottom, and keep wrapping deterministic across resize.
6. **Compact context header:** show product name, project/root context, and current run state without a dashboard-sized banner.
7. **Contextual footer:** derive hints from the active view and modal; do not display unavailable actions.
8. **Keyboard-first routing:** local modal, focused control, active view, then global bindings.
9. **In-place lifecycle updates:** represent `running`, `success`, `warning`, `failure`, `skipped`, `cancelled`, and `timed out` as explicit states rather than appended spinner lines.
10. **Trust and consent as dedicated views:** present scope and consequences before the choice, default to deny, and never hide a wait for approval behind progress.
11. **Responsive composition:** reduce decoration and optional metadata before truncating primary content or controls.
12. **Accessible fallback:** retain semantic labels, flat output, no-color behavior, and non-interactive commands.

## Patterns unsuitable for DebugDoc

- Chat-specific streaming, image attachment, mention, side-conversation, app-server, or agent-orchestration complexity.
- Copying Codex or Claude command names, branded text, logos, symbols, palette, spacing, or screen layouts.
- Treating a conversational transcript as the diagnostic information hierarchy.
- Accepting `!`, arbitrary shell text, or any composer input that bypasses typed command policy.
- Making mouse input required or capturing it by default.
- Allowing resize to reset scroll position or append duplicate content.
- Color-only, border-only, or animation-only state communication.
- A palette with actions unavailable in the current phase or actions that would silently begin Phase 3 functionality.
- Printing the full report before the interactive shell appears.
- Changing `diagnose --format text|json`, report schema `1.0`, or redirected behavior as a side effect of TUI work.

## Implementation recommendations for Go, Bubble Tea, and Lip Gloss

These are recommendations for a future approved implementation, not work performed by this research task.

### Package boundaries

Create a focused `internal/tui` package when implementation is approved:

- `model.go`: application state and top-level `Init`, `Update`, and `View`.
- `screen.go`: typed primary screens and transitions.
- `viewport.go`: logical content, wrapping, scrolling, bottom-follow, and resize invariants.
- `composer.go`: allowlisted action input, draft state, and command-palette trigger.
- `palette.go`: command registry, filtering, visibility predicates, and selection.
- `keymap.go`: contexts, bindings, priority, and footer hints.
- `styles.go`: semantic Lip Gloss tokens and terminal capability policy.
- `messages.go`: typed result, progress, consent, cancellation, and error messages.
- `terminal.go`: alternate-screen options and capability/fallback decisions.

Keep discovery, consent, process execution, report schemas, and safety policy in their existing packages. The TUI calls those boundaries; it does not recreate them.

### Bubble Tea lifecycle

- Use `tea.WithAltScreen()` for the interactive root shell.
- Keep `Update` fast and nonblocking. Run diagnosis through `tea.Cmd` and return typed messages.
- Carry `context.Context` cancellation into existing services and preserve current process-tree cleanup semantics.
- Handle `tea.WindowSizeMsg` as a first-class event and recompute every region from the new dimensions.
- Let the active overlay handle keys first, followed by focused control, active screen, and global bindings.
- Treat a terminal size below the supported minimum as a dedicated limited view, not a panic or corrupted layout.
- Do not emit the full report to stdout before starting the Bubble Tea program.

### Viewport

Use `bubbles/viewport` only if its wrapping and offset behavior can satisfy the documented invariants; otherwise keep a small project-specific viewport over logical report blocks. Cache rendered content by width, content revision, color mode, and fallback mode. Invalidate cache on any of those changes.

### Composer and palette

The composer is an action surface, not a shell. Parse only:

- Empty input/draft editing.
- Registered slash command identifiers.
- Explicit arguments defined by that command's schema.

Unknown commands show an inline error and suggestions. Never pass the raw string to `sh`, `cmd.exe`, PowerShell, or an executable resolver.

### Lip Gloss

- Build styles from semantic tokens, never hard-code presentation decisions throughout `View` functions.
- Prefer terminal-default background and ordinary foreground.
- Compute content widths after border, padding, and margin costs.
- Use `lipgloss.Width` and `lipgloss.Height` in render tests.
- Disable colors and decorative emphasis under `NO_COLOR`; keep labels and selection markers.
- Avoid assuming truecolor or a dark background.

### Compatibility

- Keep Cobra's explicit non-interactive commands unchanged.
- Keep `debugdoc diagnose --path ... --format text|json` stable.
- Keep report schema `1.0` unchanged until a separate schema migration is approved.
- Do not expose diagnostic rules or command selection that belongs to Phase 3.
- Preserve exact consent, revalidation, timeout, cancellation, bounded output, process-tree cleanup, and fail-closed behavior.

## Licensing and attribution notes

The Codex repository is licensed under Apache License 2.0 and includes a `NOTICE` file. Architecture and interaction ideas may be studied and independently reimplemented. Copying source or other protected material would introduce Apache notice, license, marking, and attribution obligations in addition to project policy concerns.

For DebugDoc:

- Keep the implementation original and compatible with DebugDoc's MIT license.
- Do not copy large or distinctive code blocks, exact text, logos, branded assets, command catalogs, or exact palettes.
- Do not imply endorsement by OpenAI or Anthropic.
- Attribute research inspiration in documentation where useful.
- Preserve license notices for every actual dependency.
- If future work proposes direct code reuse, stop and perform a separate legal/license review before merging it.
