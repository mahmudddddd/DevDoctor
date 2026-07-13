# Rule authoring

Deterministic diagnostic rules are planned for Phase 3. This document records the contract contributors should design toward.

Every rule must have:

- A stable, namespaced identifier such as `node.runtime.unsupported`
- Explicit applicability and required evidence
- Pure deterministic evaluation where practical
- Severity and confidence
- Evidence references
- A beginner-friendly explanation
- Safe and reversible remediation steps
- An optional structured verification command
- Positive, negative, and ambiguity fixtures where relevant

Rules must not print directly, invoke an AI provider, execute commands, mutate files, install dependencies, start services, or treat an exit code alone as proof of root cause. Collection and execution belong to policy-controlled application workflows.

Findings should state both what evidence proves and what remains uncertain. A connection refusal, for example, proves the endpoint was unreachable at that moment; it does not prove whether a service is stopped, on another port, container-isolated, or blocked by policy.
