# Contributing to DebugDoc

Thank you for helping make software setup errors easier to understand.

## Before contributing

- Search existing issues before opening a new one.
- Use a discussion or proposal issue for architectural changes.
- Never include real credentials, private project files, or customer logs in an issue or fixture.
- Keep diagnostics deterministic and useful without an AI provider.

## Development setup

DebugDoc supports the current and previous stable Go releases.

```bash
go mod download
go test ./...
go vet ./...
go build ./cmd/debugdoc
```

Run `gofmt` on changed Go files. CI also runs `golangci-lint` and cross-platform tests. Changes to command execution must also pass `go test -race ./...` and platform-specific timeout, cancellation, output-flood, and descendant-cleanup tests.

## Design expectations

- Treat the selected project and its contents as untrusted.
- Do not execute a project command without a policy decision and user consent.
- Do not read secret-bearing files by default.
- Keep terminal rendering outside detectors and future diagnostic rules.
- Prefer standard-library functionality and small, well-maintained dependencies.
- Ensure redirected/non-interactive execution never waits for a prompt.
- Accept executable paths and argument vectors, never shell command strings.
- Bind every command to a canonical project root and revalidate approved filesystem identities before start.
- Keep stdout and stderr independently bounded and treat incomplete process-tree cleanup as an error.
- Never expose environment values in consent, logs, errors, or results.

## Adding detection support

Every advertised language, framework, runtime, or package-manager detector needs:

- Positive fixture coverage
- Negative fixture coverage
- Ambiguity/conflict coverage when applicable
- A clear evidence source
- No project-script execution

## Adding diagnostic rules

Rules arrive in Phase 3. Each rule will require a stable ID, evidence requirements, beginner-facing language, safe remediation, and regression fixtures. See [docs/rule-authoring.md](docs/rule-authoring.md).

## Pull requests

Keep pull requests focused. Explain the behavior change, safety impact, tests, and platform considerations. A change is not complete while supported-platform tests fail.

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
