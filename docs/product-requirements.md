# Product requirements

## Purpose

AI Execution Advisor helps engineering and finance teams understand AI execution cost and decide whether workloads should retain their current implementation or be tested with deterministic code, exact caching, retrieval, smaller hosted models, local models, reduced context/output, batching, or structured output.

## Product truths

The system must separate four stages:

1. Observed usage and cost.
2. Estimated optimization opportunity.
3. Evaluated or experimentally measured outcome.
4. Verified savings from a completed controlled experiment.

An approval permits an evaluation or experiment; it never performs an uncontrolled production switch. Demonstration prices are illustrative and carry an effective date. Quality percentages are reported only when supported by evaluation data.

## Primary users

- Owners and admins govern the organization and team.
- Engineers inspect workloads, evaluations, and experiments.
- Finance users monitor cost, budgets, and savings evidence.
- Viewers receive read-only access.

## Functional scope

Planned scope includes organization-scoped authentication, projects, workloads, usage ingestion, a versioned model pricing catalog, analytics, deterministic recommendations, evaluations, controlled experiments, budgets and alerts, public educational tools, and a mobile operations companion.

## Success criteria

- Every recommendation exposes reason codes, evidence, confidence inputs, financial assumptions, risk, and required next action.
- Tenant boundaries and role permissions are tested.
- Monetary values use decimal-safe arithmetic and UTC timestamps at rest.
- Raw prompts and responses are not stored by default.
- A deterministic demo can exercise the documented end-to-end journey without external model credentials.

## Out of scope for v0.1.0

Automatic replacement-code deployment, semantic caching, a custom OAuth provider, real production traffic proxying by default, Kubernetes, Kafka, Elasticsearch, GraphQL, and unsupported model-quality claims are out of scope.
