# Domain model

## Ownership hierarchy

An organization owns users through memberships and owns projects. A project owns workloads; a workload groups LLM calls, classifications, recommendations, evaluation datasets, and experiments. Budget alerts belong to projects. Model catalog and pricing entries describe execution economics without containing provider credentials.

Every tenant-owned query includes organization scope. Cross-tenant identifiers must return a safe not-found or forbidden response according to the endpoint's disclosure policy.

## Core lifecycle concepts

- **LLM call:** immutable, redacted execution observation priced using the effective catalog entry.
- **Task classification:** evidence-backed characterization at call or workload level.
- **Recommendation:** versioned rule output with proposed execution, risks, confidence inputs, and estimated—not verified—savings.
- **Evaluation:** sanitized cases and repeatable candidate scoring.
- **Experiment:** guarded comparison of control and candidate execution.
- **Verified savings:** recorded only after an experiment completes and satisfies verification rules.

## State models

Recommendation status progresses through `new`, `under_review`, `accepted` or `rejected`, then optionally `testing`, `applied`, and `verified`; `dismissed` is terminal. Transitions will be explicit and audited.

Experiments use `draft -> approved -> running -> completed -> verified`, with `running -> paused`, `paused -> running`, and rollback from running or paused. Invalid and concurrent transitions must fail safely.

## Value rules

Money is stored as decimal amounts with ISO currency codes, never binary floating-point. Timestamps are stored in UTC. Pricing periods cannot overlap for the same provider/model/currency. Enumerations for roles, workload properties, recommendation types, and statuses are validated at boundaries and constrained in storage where appropriate.

## Privacy

LLM calls store hashes and redacted previews, not raw prompts or full responses by default. Metadata is treated as untrusted input and passes through field-name and value-pattern redaction before storage.
