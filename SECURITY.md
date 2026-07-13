# Security Policy

## Supported versions

DevDoctor has not published a stable release yet. Security fixes will be applied to the latest development branch until a version support policy is announced.

## Reporting a vulnerability

Do not open a public issue for suspected vulnerabilities, exposed secrets, path traversal, unsafe process execution, redaction failures, or supply-chain concerns.

Use GitHub's private vulnerability reporting feature for this repository. Include:

- A concise description and impact
- Reproduction steps using synthetic data
- Affected operating systems and DevDoctor version/commit
- Any suggested mitigation

Do not include real credentials, proprietary source code, or third-party personal data. Maintainers will acknowledge a complete report as soon as practical and coordinate disclosure after a fix is available.

## Security model

The selected project directory, project scripts, logs, configuration, symlinks, and coding-agent output are untrusted. DevDoctor is designed around project-root containment, allowlisted metadata reads, explicit consent for risky operations, bounded subprocesses, and redaction before persistence or sharing. See [docs/architecture.md](docs/architecture.md) and [docs/privacy.md](docs/privacy.md).
