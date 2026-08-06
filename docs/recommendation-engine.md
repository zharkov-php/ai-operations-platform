# Recommendation engine

Rule version `v1.0.0` is deterministic and does not call an LLM. It analyzes at most 500 active workloads per organization per run. A snapshot records call volume, hashed-input repetition, template stability, task metadata coverage, token averages, configured relevant/required token counts, structured-output failures, burst size, current 30-day cost, pricing completeness, and workload risk classifications.

The engine can propose deterministic code, exact caching, a smaller hosted model, a local-model candidate, context/output reduction, structured output, batching, or keeping the current model. Critical quality/failure impact and insufficient volume produce `keep_current_model` before cost-saving rules are considered. Exact cache uses prompt-hash equality only; semantic caching is not implemented.

Confidence is descriptive evidence strength, not a statistically validated quality probability. Its inspectable inputs are analyzed call count, template stability, repeated-input rate, task-classification coverage, and pricing completeness. Ground-truth and evaluation availability are stored as false until those systems provide evidence.

Savings and implementation costs are estimates. Conservative v1 factors are documented in code and persisted with each recommendation. Local-model estimates explicitly separate illustrative hardware amortization, electricity, and maintenance, require measured throughput/latency and evaluation quality, enforce a context limit, and calculate break-even. No recommendation creates replacement code, changes routing, or claims verified savings.

Re-running the same rule version refreshes evidence on the existing `(workload, recommendation type, rule version)` row and preserves its workflow status. A new methodology must use a new rule version.
