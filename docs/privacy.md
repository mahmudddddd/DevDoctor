# Privacy and trust model

DevDoctor is local-first. Phase 1 performs no telemetry, uploads, AI calls, dependency installation, project-script execution, or network diagnostics.

## Data read during discovery

Discovery reads directory entry names and a small allowlist of metadata files, such as:

- `package.json`
- `tsconfig.json` and `jsconfig.json`
- Recognized package-manager lockfiles
- Recognized workspace and framework configuration filenames
- Docker and Compose configuration filenames

The current implementation only parses `package.json`; other recognized files are evidence by presence. Reads are bounded and project-root contained.

## Files denied by default

Sensitive or high-volume locations are denied even if requested through the discovery file policy:

- `.env` and `.env.*` except `.env.example`
- Private keys, certificates, keystores, and credential files
- `.git`, dependency directories, build output, caches, and editor metadata
- Files outside the selected project root
- Metadata symlinks of any kind; Phase 1 reads only direct regular files

Future user-selected logs will be size-limited and redacted before display or persistence.

## Consent boundaries

Later phases require specific approval before build/start commands, network checks, Docker or database actions, agent invocation, data sharing, or file changes. Approval must identify the operation, working directory, destination, mutation/network classification, and data bundle.

## Coding-agent integrations

Agents are optional and untrusted. The deterministic core remains available without them. Before any handoff, DevDoctor will show the exact included files, metadata, and logs; exclusions; redaction counts; bundle size; and selected provider. Agent text cannot bypass command, path, privacy, or mutation policy.

## Telemetry

There is no telemetry in the MVP. Any future proposal must be opt-in, documented, reviewable, and must not include project paths, filenames, source, logs, environment values, or diagnostic bundles.

## Phase 2 execution privacy

Phase 2 introduces an internal execution boundary but does not wire project commands into `diagnose`.

Before consent, DevDoctor canonicalizes the selected project root and working directory, rejects working directories outside the root, resolves the executable to an absolute regular file, and records filesystem identities. The same identities are checked immediately before process start; a changed path requires a new preparation and approval.

Consent displays the operation ID and purpose, exact executable and argument boundaries, canonical working directory, mutation/network/service classifications, timeout and termination grace, independent stream limit, environment variable names, and declared data descriptors. Environment values are never shown, logged, or returned. Grants are in-memory only and are scoped to once, the exact check, or the exact request for the current run. Non-interactive approval is unavailable and fails closed without reading stdin.

The runner does not read `.env` files, shell profiles, or startup scripts and does not inherit the full host environment. It supplies a small platform launch baseline and only explicitly declared additional names or overrides. Known secret-bearing additions are rejected until a later explicit data policy exists.

Stdout and stderr are retained independently up to the approved limit, then drained and counted without retaining excess bytes. Captures are valid UTF-8, carry explicit truncation metadata, stay in memory, and are not persisted or uploaded in Phase 2.
