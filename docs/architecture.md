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
