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
