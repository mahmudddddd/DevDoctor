# DebugDoc

DebugDoc is a beginner-friendly command-line tool that explains why a software project fails to build or start.

It detects a project's stack, checks common setup problems, and turns evidence into clear explanations and actionable next steps. Its core diagnostics are deterministic, local, and usable without AI.

> **Status:** DebugDoc is in early development. Phase 1 provides safe project discovery for common Node.js and TypeScript projects on Windows, macOS, and Linux. Phase 2 adds an internal, consent-gated process-execution boundary for future checks. No current diagnostic executes project scripts.

## Why DebugDoc?

Build and startup errors often assume you already understand runtimes, package managers, ports, environment variables, Docker, and framework configuration. DebugDoc gathers that evidence in one place and explains it in beginner-friendly language.

A future diagnostic finding will tell you:

1. What failed
2. What evidence DebugDoc found
3. What the evidence means
4. What you can safely try next
5. How to verify the fix

## Installation

Prebuilt standalone binaries will be published for Windows, macOS, and Linux with the first release.

Go users can build the current development version with:

```bash
go install github.com/mahmudddddd/DevDoctor/cmd/debugdoc@latest
```

## Usage

Open the full-screen interactive shell:

```bash
debugdoc
```

Inspect a project without executing its scripts:

```bash
debugdoc diagnose --path ./my-project
```

Produce machine-readable output:

```bash
debugdoc diagnose --path ./my-project --format json
```

Show build information:

```bash
debugdoc version
```

When input or output is redirected, DebugDoc never opens an interactive prompt. Use the `diagnose` command explicitly in scripts and CI.

Set `DEBUGDOC_REDUCE_MOTION=1` to replace cosmetic terminal animation with static lifecycle labels. `TERM=dumb` also uses static ASCII presentation, while `NO_COLOR` disables color independently.

## Current project discovery

Phase 1 can report:

- JavaScript and TypeScript markers
- Node.js projects and declared Node engine requirements
- npm, pnpm, Yarn, and Bun package-manager evidence
- Common framework dependencies, including Next.js, Nuxt, Vite, React, Vue, Angular, Svelte, Remix, Astro, and NestJS
- npm and pnpm workspace layouts
- Relevant files that were safely inspected
- Warnings such as conflicting lockfiles or an unreadable manifest

Support is documented by tested detector behavior rather than broad language claims. See [docs/supported-checks.md](docs/supported-checks.md).

## Safety and privacy

DebugDoc is local-first. By default, it does not:

- Send project data anywhere
- Use AI or require an account
- Install or update packages
- Modify project files
- Start Docker containers or databases
- Execute build, test, or startup scripts during project discovery
- Read secret files such as `.env`, private keys, or credential stores

Phase 1 reads only a small allowlist of project metadata files and enforces project-root containment. Phase 2 implements the internal validation, exact consent, bounded-output, timeout, cancellation, and process-tree cleanup boundary required before a future check can run a command. It does not select commands or expose arbitrary execution, and `diagnose` remains discovery-only.

See [docs/privacy.md](docs/privacy.md) for the privacy and trust model.

## Optional coding-agent assistance

DebugDoc's core diagnostics will work without AI.

A later optional integration will be able to hand an approved, redacted diagnostic bundle to an installed Claude Code, Codex, Gemini, or compatible coding-agent CLI. Users will review the bundle and any proposed diff before data is shared or changes are applied.

## Project scope

The first milestone is a reliable deterministic diagnostic engine for common Node.js and TypeScript build/start failures. Automatic fixes, broad language support, cloud services, IDE integrations, and the future Guard module are intentionally outside the first MVP.

## Development

DebugDoc targets the current and previous stable Go releases.

```bash
go mod download
go test ./...
go vet ./...
go build ./cmd/debugdoc
```

If `golangci-lint` and GoReleaser are installed:

```bash
golangci-lint run
goreleaser release --snapshot --clean
```

See [CONTRIBUTING.md](CONTRIBUTING.md), [docs/rule-authoring.md](docs/rule-authoring.md), and [docs/ux-research.md](docs/ux-research.md).

## Security

Do not open public issues for suspected vulnerabilities or accidental secret exposure. Follow the private reporting process in [SECURITY.md](SECURITY.md).

## Roadmap

1. Phase 0: repository and release foundation
2. Phase 1: safe project discovery
3. Phase 2: consent-gated command execution foundation
4. Phase 3: deterministic rules and reports
5. Approved build/start diagnosis
6. Optional coding-agent handoff
7. Additional ecosystems and the future Guard module

## License

DebugDoc is released under the [MIT License](LICENSE).
