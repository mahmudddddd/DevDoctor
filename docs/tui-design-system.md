# DevDoctor terminal UI design system

This document is the permanent interaction and presentation contract for DevDoctor's interactive terminal UI. It defines behavior for future implementation; it does not itself authorize new diagnostics, commands, report fields, or production behavior.

The interface must be visually original. Codex and Claude Code are research references, not templates.

## Design principles

1. **Diagnosis before decoration.** The current project, current state, evidence, and next safe action are always clearer than branding or chrome.
2. **One place, one purpose.** One primary view is active at a time. Short overlays may temporarily collect a choice or show help.
3. **Keyboard first.** Every action is reachable without a mouse. Mouse support, if later added, is optional and never changes the information model.
4. **Stable under change.** Resize, progress, and in-place updates preserve focus, draft text, selected item, and deliberate scroll position.
5. **Semantic without color.** Words and structure carry meaning; color only reinforces it.
6. **Safe by construction.** The composer accepts registered DevDoctor actions, never arbitrary shell input.
7. **Scriptable separately.** The alternate-screen root shell does not replace or alter non-interactive Cobra commands.
8. **Beginner legibility.** Prefer plain language, visible scope, and one recommended next step over dense telemetry.

## Information hierarchy

Present information in this order:

1. **Safety-critical interruption or consent**: what is requested, scope, consequence, and choice.
2. **Current run state**: running, waiting, success, warning, failure, cancelled, timed out, or skipped.
3. **Primary finding or current task**: what DevDoctor is checking or what needs attention.
4. **Evidence and explanation**: why the state or finding exists.
5. **Next safe action**: a registered action the user can take.
6. **Project context**: selected root, detected stack, and report metadata.
7. **Secondary help and shortcuts**.

Do not put decorative identity above consent or failure information. Do not display transient progress as if it were durable evidence.

## Application frame

The root interactive command uses a full-screen alternate-screen Bubble Tea application. The frame has four persistent regions:

```text
┌ header: product + project + state ┐
│                                  │
│ viewport: one active view        │
│                                  │
├ composer / active prompt         ┤
└ footer: contextual key hints     ┘
```

At compact sizes, borders may be removed, but the four regions and their order remain. The full report must not be printed before entering the interactive shell. Normal stdout/stderr report rendering remains available through explicit non-interactive commands.

The frame uses terminal-default background. Outer margins are at most one cell per side and disappear before primary content is truncated.

## Header

### Required content

- `DevDoctor` product name in plain text.
- Selected project name or `No project selected`.
- Short current state label.

### Optional content, in priority order

- Truncated project root.
- Current view name.
- Count summary such as `2 warnings`.
- Version or elapsed time only in wide/tall layouts and only when useful.

### Rules

- Compact header: one or two rows, no logo art.
- Standard/wide header: at most three rows.
- Project root uses middle elision so the drive/root and leaf remain visible.
- State labels are words: `READY`, `RUNNING`, `WAITING`, `OK`, `WARNING`, `FAILED`, `CANCELLED`, `TIMED OUT`, `SKIPPED`.
- Never use a spinner as the only state indicator.

## Viewport

The viewport owns all primary content. Valid primary views include:

- Home/project selection.
- First-run trust explanation.
- Diagnostic running view.
- Results summary.
- Finding/evidence detail.
- Consent request.
- Error/recovery view.
- Help.

Only one primary view is active. A result detail replaces the summary in the viewport rather than opening a second permanent column at compact and standard widths.

### Scrolling rules

- `Up`/`Down`: one logical row or wrapped line as appropriate.
- `PgUp`/`PgDn`: one viewport page minus one context row.
- `Home`/`End`: top/bottom when the focused control does not consume them.
- Scrolling up disables auto-follow.
- New running content follows only when already at the bottom.
- Show `Jump to latest` or an equivalent textual affordance when updates arrive above a manually scrolled position.
- Resize preserves the nearest logical content anchor, not a stale wrapped-line number.
- Scroll indicators include text or position, not color alone.

## Composer

The composer is persistent except when a full-screen safety-critical consent view temporarily replaces it with explicit choices.

### Purpose

- Invoke registered slash commands.
- Filter actions.
- Enter narrowly scoped arguments defined by an action.
- Preserve a draft while help, details, or a palette overlay is opened and closed.

### Prohibited behavior

- No arbitrary shell mode.
- No `!` escape.
- No raw command string passed to a shell, executable resolver, or process runner.
- No hidden Phase 3 diagnostic action.
- No command whose permission or phase availability is ambiguous.

### Layout

- Prompt marker: original DevDoctor marker such as `doctor>` in color mode and `>` in ASCII/flat mode.
- One row when empty; grows to a maximum of three text rows at 80x24 and four rows at larger sizes.
- Overflow scrolls inside the composer rather than consuming the viewport.
- Inline validation appears directly above or below the input and uses a full sentence.
- Cursor remains visible after resize and popup dismissal.

### Input behavior

- `/` as the first non-space character opens the command palette.
- `Enter` selects a palette item or submits a complete registered action.
- `Esc` closes the palette before clearing a draft.
- `Ctrl+C` interrupts active work; when idle it clears nonempty input before a second deliberate quit action.
- Multiline input is not required for the initial TUI. Do not add it without a defined DevDoctor use case.

## Slash-command palette

The palette is an overlay anchored to the composer when space allows and a centered bounded overlay otherwise.

Each command has:

- Stable internal ID.
- User-facing slash name.
- One-sentence description.
- Availability predicate by view, run state, phase, and terminal mode.
- Typed argument schema, if any.
- Key hint derived from the central keymap, if one exists.

### Palette rules

- Filter case-insensitively by command name and description.
- Show only currently available commands by default.
- Unknown input shows `No matching DevDoctor command` and does not submit.
- Selection is indicated by `>` in all modes; color or background may reinforce it.
- Maximum width: `min(72, terminal width - 4)`.
- Maximum height: `min(12, viewport height - 2)`.
- Long descriptions wrap or elide after the command remains fully visible.
- Palette order follows task relevance, then stable alphabetical order; it must not be an accidental enum/source order.
- Destructive or high-consequence actions never execute immediately from palette selection; they transition to a dedicated consent view.

## Footer

The footer shows only actions valid in the current context.

Examples:

- Home: `Enter select  / commands  ? help  Ctrl+C quit`
- Results: `↑↓ scroll  Enter details  / commands  Esc back`
- Palette: `↑↓ move  Enter choose  Esc close`
- Running: `PgUp/PgDn scroll  End latest  Ctrl+C cancel`
- Consent: `Tab move  Enter choose  Esc deny`

Rules:

- Hints come from the central keymap, not duplicated strings.
- Compact layouts show at most four actions and prioritize the focused control.
- Use spaces and separators that remain understandable in no-color mode.
- Never advertise unavailable or unsafe actions.

## Navigation and focus

### Input priority

1. Safety-critical consent surface.
2. Topmost overlay or menu.
3. Focused control, including palette or composer.
4. Active primary view.
5. Global bindings.

### Focus rules

- There is exactly one keyboard focus owner.
- Opening an overlay records prior focus; closing restores it when still valid.
- Resize never changes focus.
- Disabled controls cannot receive focus.
- Focus uses both a marker/label and style.
- `Esc` dismisses the topmost dismissible layer. It does not unexpectedly quit.
- Quitting with active work follows existing cancellation and cleanup policy.

### Baseline bindings

| Key | Default meaning |
| --- | --- |
| `Enter` | Activate or submit the focused safe action |
| `Esc` | Close overlay, go back, or deny consent |
| `Up` / `Down` | Move selection or scroll |
| `PgUp` / `PgDn` | Page viewport |
| `Home` / `End` | Top/bottom where not consumed by input |
| `/` | Open/filter command palette from composer |
| `?` | Contextual help when composer is empty |
| `Ctrl+L` | Redraw |
| `Ctrl+C` | Cancel active work; otherwise clear/quit through deliberate state |

Avoid Ctrl-only bindings as the sole way to perform ordinary navigation. Do not bind keys commonly required for text entry without a visible alternative.

## Responsive breakpoints

The required support matrix is 80x24, 100x30, and 120x40.

### Width classes

| Class | Terminal width | Behavior |
| --- | ---: | --- |
| Compact | 80-99 | No outer border; one-column views; shortest labels; path and metadata elision; palette may use most of viewport width |
| Standard | 100-119 | One-column primary view with fuller header/status and descriptions |
| Wide | 120+ | May add a secondary metadata rail only when it does not create two competing primary views |

### Height classes

| Class | Terminal height | Behavior |
| --- | ---: | --- |
| Compact | 24-29 | Two-row header maximum; composer maximum three text rows; reduced padding; viewport gets all remaining rows |
| Standard | 30-39 | Header up to three rows; composer maximum four text rows; normal spacing |
| Tall | 40+ | More viewport context and detail; chrome does not grow merely because space exists |

### Below minimum

Below 80 columns or 24 rows, render a stable limited view:

- State that DevDoctor needs at least 80x24 for the interactive shell.
- Show current size.
- Preserve active work, focus intent, and draft state.
- Allow cancellation and quit.
- Do not panic, overlap controls, or silently switch to non-interactive behavior.

## Width and height budgets

All calculations use terminal cells after accounting for ANSI-free rendered width.

### Horizontal budget

- Compact: zero outer margin; zero or one-cell internal side padding.
- Standard: one-cell outer margin or frame padding per side.
- Wide: at most two total cells of outer margin per side; content width remains the priority.
- Borders cost two columns and must be removed when they would reduce primary content below 76 cells at the supported minimum.
- Header state reserves at most 18 cells before project text is elided.
- Footer hints wrap to a second row only at standard/tall heights; otherwise omit lower-priority hints.

### Vertical budget

Use this formula:

```text
viewport height = terminal height
                - header height
                - composer height
                - footer height
                - separators/padding
```

Target budgets:

| Terminal | Header | Composer | Footer | Separators/padding | Viewport minimum |
| --- | ---: | ---: | ---: | ---: | ---: |
| 80x24 | 2 | 4 | 2 | 2 | 14 |
| 100x30 | 3 | 5 | 2 | 2 | 18 |
| 120x40 | 3 | 5 | 2 | 2 | 28 |

Consent views may borrow the composer rows, but must retain at least one footer row with denial/cancel guidance. Overlays never exceed the viewport's content box.

## Spacing

Use a four-step terminal-cell spacing scale:

- `0`: adjacent semantic content.
- `1`: default inset and row gap.
- `2`: section separation.
- `3`: maximum large break in tall layouts.

Never use more than two consecutive blank rows. At 80x24, reduce section gaps before reducing readable labels or viewport height.

## Borders

- Borders are optional grouping aids, not structure.
- Use one border style at a time.
- Default Unicode: light single-line borders.
- ASCII fallback: `+`, `-`, and `|`.
- Compact mode removes the outer frame first, then nonessential internal borders.
- Focus and error states must remain visible without a colored border.
- Do not use product-specific or copied decorative border motifs.

## Text hierarchy

| Level | Use | Treatment |
| --- | --- | --- |
| Product/context | Product and selected project | Strong weight where supported; no oversized art |
| View title | Current task or result title | Strong, one line, wraps only when necessary |
| Section heading | Evidence, next step, details | Strong or underlined; sentence case |
| Body | Explanations and evidence | Terminal-default foreground |
| Metadata | Paths, counts, timestamps | Muted but still contrast-safe; labels retained |
| Hint | Keyboard and secondary help | Muted; never required to understand state |
| Code/value | Paths, command arguments, IDs | Monospace is inherent; quote boundaries explicitly |

Do not communicate hierarchy through color alone. Avoid all-uppercase paragraphs; uppercase is reserved for short state labels.

## Semantic colors

Colors are DevDoctor-specific semantic tokens, not a copied palette. Exact terminal output may be adapted to ANSI capability.

| Token | Suggested truecolor | Meaning |
| --- | --- | --- |
| `text` | terminal default | Body content |
| `muted` | `#8E98A4` | Secondary metadata and hints |
| `accent` | `#58B7A8` | Focus, selected safe action, DevDoctor identity |
| `info` | `#6FA8DC` | Neutral notices and current context |
| `success` | `#72B889` | Completed successfully |
| `warning` | `#D6A84B` | Attention needed, nonfatal |
| `error` | `#D97777` | Failure or invalid input |
| `consent` | `#B091D8` | Permission/trust boundary |

Rules:

- Use terminal-default background.
- Never put ordinary body text in a low-contrast muted color.
- Pair every semantic color with a word and, where useful, a symbol.
- Do not assume dark mode or truecolor.
- Map gracefully to ANSI 16/256 colors based on terminal capability.
- Avoid animated color cycling.

## `NO_COLOR` behavior

When `NO_COLOR` is present, regardless of value:

- Emit no ANSI color sequences from the TUI's styles.
- Remove background fills, colored borders, and hue-based emphasis.
- Preserve bold/underline only if the selected flat/accessibility policy allows it; correctness must not depend on them.
- Use textual labels and ASCII/Unicode markers for selection and state.
- Keep layout, focus order, commands, and information unchanged.
- Ensure snapshots contain no escape sequences.

Non-TTY execution must never enter the TUI and must continue using stable text/JSON behavior.

## ASCII and flat fallbacks

| Semantic | Unicode/default | ASCII/flat |
| --- | --- | --- |
| Selected | `›` | `>` |
| Success | `✓ OK` | `[OK]` |
| Warning | `! WARNING` | `[WARN]` |
| Failure | `× FAILED` | `[FAIL]` |
| Running | `• RUNNING` with optional restrained animation | `[RUNNING]` |
| Pending | `○ PENDING` | `[PENDING]` |
| Cancelled | `– CANCELLED` | `[CANCELLED]` |
| Tree branch | `├─`, `└─` | `+-`, `\-` |
| Border | light line | `+ - |` |

A flat/screen-reader mode should remove decorative borders, animation, cursor-only cues, and unnecessary repeated chrome while retaining the same actions and state labels.

## Empty states

Every empty state answers:

1. What is empty?
2. Is that expected?
3. What safe action is available?

Examples:

- `No project selected. Choose a project to inspect safe metadata.`
- `No warnings found in the current discovery report.`
- `No command matches “/foo”. Press Esc to keep editing.`

Do not use blank panels, decorative mascots, or jokes in safety/error contexts.

## Running states

A running view contains:

- Explicit `RUNNING` label.
- Current operation in plain language.
- Project root or scope.
- Completed/total count only when total is known.
- Elapsed time only when accurate and useful.
- Clear cancellation key.
- In-place progress updates, not appended spinner lines.

Animation is optional, low frequency, and removed in flat mode. Final states have priority over progress messages so completion cannot be obscured by a late tick.

## Success, warning, and error states

### Success

- State `OK` or `COMPLETED`.
- One-sentence result.
- What was inspected.
- Next optional action.

### Warning

- State `WARNING`.
- Explain impact and evidence.
- Distinguish incomplete inspection from an actual detected problem.
- Provide a safe next step; do not imply failure when the run completed.

### Error

- State `FAILED` or a more precise terminal state such as `TIMED OUT` or `CANCELLED`.
- State what failed, what completed, and whether any result is partial.
- Preserve completed evidence.
- Offer retry/back/help only when valid.
- Sanitize terminal control characters and untrusted project text before rendering.
- Do not expose environment values, secrets, or unsafe raw output.

## Consent and trust screens

Consent is a primary view, not a small ambiguous popup.

Required content:

- Clear title: `Approval required` or `Trust this project?`.
- Exact operation and purpose.
- Canonical project root and working directory.
- Executable and each argument as separate boundaries when command execution is involved.
- Mutation, network, service, timeout, termination grace, output limit, environment **names**, and data descriptors as required by existing policy.
- What approval scope means.
- Consequence of denial.

Interaction rules:

- Default selection is the safest deny/cancel outcome.
- `Esc` denies or cancels.
- Approval choices use full text, not only single-letter shortcuts.
- High-consequence choices require deliberate selection and `Enter`; palette selection alone never approves.
- Non-interactive mode reports approval unavailable and fails closed without reading stdin.
- Any material request change invalidates prior approval.
- Never display environment values.

First-run trust must explain that project content is untrusted data and identify the selected root before proceeding. Trust UI must not weaken the existing consent manager or filesystem identity revalidation.

## Testing requirements

Future TUI implementation is incomplete without these tests.

### Pure state/update tests

- Every primary screen transition.
- Overlay-first key routing and focus restoration.
- Composer draft preservation.
- Registered slash command parsing and unknown-command rejection.
- Proof that arbitrary shell input cannot reach execution.
- Cancellation, completion, timeout, warning, failure, and skipped precedence.
- Late progress messages cannot overwrite terminal states.

### Responsive render tests

Golden/snapshot tests at minimum:

- 80x24.
- 100x30.
- 120x40.
- 79x23 limited mode.
- Long Windows and Unix paths.
- Long unbroken tokens and wide Unicode.
- Palette open, consent, running, results, warning, and error views.
- Color, `NO_COLOR`, ANSI-limited, and ASCII/flat modes.

Assertions must check rendered width and height with ANSI-aware helpers and fail on overflow.

### Viewport tests

- Bottom-follow while at bottom.
- Manual scroll preserved when updates arrive.
- Explicit jump to latest.
- Resize reflow preserves logical anchor.
- Empty and one-line content.
- Wrapped content and content revision cache invalidation.

### Keymap and accessibility tests

- No conflicting bindings in overlapping contexts.
- Footer hints match actual bindings.
- Every action has a keyboard path.
- Selection and semantic states remain detectable after ANSI removal.
- `NO_COLOR` output contains no ANSI color sequences.
- Flat mode contains no animation-dependent or border-dependent information.

### Integration and compatibility tests

- Interactive root enters and restores alternate screen on success, error, cancellation, and panic-equivalent failure paths where testable.
- Resize events do not corrupt the frame.
- Non-TTY root behavior remains fail-closed.
- `devdoctor diagnose --format text` remains stable.
- `devdoctor diagnose --format json` keeps report schema `1.0` unchanged.
- No implementation exposes arbitrary shell execution or begins Phase 3 rules.
- Existing consent, timeout, cancellation, bounded-output, identity-revalidation, process-tree cleanup, and fail-closed tests continue to pass.
- Run supported-platform tests and a PTY smoke test on Windows, macOS, and Linux when the implementation phase begins.

## Change control

Changes to this design system require an explicit UI design decision. Changes that affect report schema, diagnostic functionality, command execution, consent, process safety, or non-interactive behavior require separate approval and must not be bundled into visual TUI work.
