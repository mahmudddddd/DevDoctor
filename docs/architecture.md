# Architecture

## Goals

DevDoctor is a standalone, local-first CLI that turns project evidence into deterministic, beginner-friendly diagnostics. The architecture must keep working when the target runtime is missing, avoid trusting project content, and support Windows, macOS, and Linux.

## Layers

```text
CLI and interactive presentation
        ↓
Application workflows
        ↓
Discovery / collection → deterministic rules → report
        ↓                         ↓
Privacy and platform adapters   bounded process runner
        ↓
Optional coding-agent adapters
```

Phase 1 implements CLI presentation, safe discovery, a file policy, structured project summaries, and text/JSON rendering. It does not execute project scripts.

## Package responsibilities

- `cmd/devdoctor`: executable entry point only.
- `internal/cli`: Cobra commands, Huh interaction, TTY policy, and dependency wiring.
- `internal/app`: use-case orchestration without terminal formatting.
- `internal/model`: serializable domain models and schema versions.
- `internal/detect`: evidence-based project, stack, workspace, runtime, and package-manager detection.
- `internal/privacy`: project-root containment and file-read policy.
- `internal/report`: text and JSON rendering.
- `internal/version`: build metadata populated by release ldflags.

Future packages add fact collectors, deterministic rules, bounded process execution, platform-specific process-tree handling, and provider-neutral coding-agent adapters.

## Dependency direction

Presentation depends on application workflows and models. Application workflows depend on detectors and policies. Detectors return structured data and never print. Domain models do not depend on Cobra, Huh, terminal styling, or external agents.

## Untrusted input

The selected project, manifests, filenames, symlinks, logs, project commands, and agent output are untrusted. Phase 1 controls reads through an allowlist and verifies canonical paths remain within the selected root. Later execution must use structured executable/argument specifications, explicit consent, timeouts, output limits, and process-tree cleanup.

## Determinism

Given the same readable project metadata, detector version, and options, discovery produces the same ordered summary. Detection evidence is included in the report so users can understand why a technology was identified.

## Report compatibility

JSON reports include a schema version. Fields may be added compatibly before v1, but consumers should reject unsupported major schema versions. Detector identifiers and future finding rule IDs become public compatibility surfaces once released.

## Phase 2 command boundary

Phase 2 adds reusable internal infrastructure; it does not add a command-selection feature or change the Phase 1 `ProjectReport` schema `1.0`.

The command path is:

```text
CommandSpec
    → project/cwd/executable preparation
    → immutable exact consent request
    → immediate identity revalidation
    → shell-free runner
    → bounded stdout/stderr + structured CommandResult
```

`internal/model` defines command classifications, environment declarations, stream captures, and lifecycle results. `internal/privacy.PathPolicy` canonicalizes the selected project root, requires working directories to remain within it, resolves the executable to an absolute regular file, and retains filesystem identities for pre-start revalidation.

`internal/consent` fingerprints the complete immutable request. Grants exist only in memory and are limited to once, the exact check, or the exact request for the current run. A changed path, executable, argument, classification, limit, environment declaration, or data descriptor requires another decision.

`internal/app.ExecutionService` owns the prepare → approve → revalidate → execute sequence. Denied or unavailable approval produces a structured skipped result and never reaches the runner.

## Runner and process lifecycle

`internal/runner` accepts only a prepared executable and argument vector. It never constructs a shell command and never implicitly invokes `sh`, `cmd.exe`, or PowerShell. A minimal platform environment is assembled from documented baseline names plus explicitly declared additions or overrides; `.env` files and shell profiles are not loaded.

Stdout and stderr are drained independently. Each stream retains only its approved prefix while continuing to discard and count excess bytes, preventing pipe blockage and unbounded memory use. Captures remain in memory and include explicit captured, discarded, and truncation metadata.

The runner records cancellation and timeout separately. On Unix it creates a dedicated process group, sends `SIGTERM`, waits the approved grace period, and escalates to `SIGKILL`. On Windows it starts the process suspended, assigns it to a kill-on-close Job Object before resuming its primary thread, and terminates the job on cancellation or timeout. Cleanup failure is reported separately without replacing the primary terminal state.
