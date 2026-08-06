# Cost methodology

Hosted cost is `(input tokens × input price + output tokens × output price + cached input tokens × cached price) ÷ unit size`. Values use decimal arithmetic and pricing effective at request time. Missing prices fail explicitly.

Observed cost comes from ingested executions. Candidate cost and savings are estimates. Savings become verified only after an experiment satisfies quality, latency, error, and cost guardrails. Local economics include amortization, electricity, maintenance, capacity, and utilization; deterministic replacements include implementation and maintenance rather than claiming zero cost.
