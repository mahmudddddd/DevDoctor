# Coding-agent integrations

Coding-agent support is planned after deterministic diagnosis is reliable. DevDoctor will not require an agent, account, API key, or network connection for its core behavior.

## Adapter boundary

A provider-neutral adapter will expose discovery, capabilities, request preparation, invocation, and response parsing. Provider-specific behavior for Claude Code, Codex, Gemini, and generic CLIs remains behind that boundary.

## Consent flow

1. Complete deterministic local diagnosis.
2. Build a minimal redacted bundle.
3. Show the provider, included data, exclusions, redaction counts, and size.
4. Ask for exact bundle approval.
5. Invoke the selected agent in advice/proposal mode.
6. Treat the response as untrusted.
7. Show any proposed patch separately.
8. Require explicit diff approval before a later mutation phase.
9. Rerun the approved deterministic verification command.

Project files and logs may contain prompt-injection text. They are data, not instructions, and cannot alter DevDoctor policy.
