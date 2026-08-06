# Experiment guardrails

Experiments compare control and candidate execution with bounded candidate traffic. Quality, latency, error-rate, and cost thresholds are evaluated before completion. A violation records evidence and rolls back candidate traffic; successful completion can later become verified with measured savings.

```mermaid
stateDiagram-v2
  [*] --> draft
  draft --> approved
  approved --> running
  running --> paused
  paused --> running
  running --> completed
  completed --> verified
  running --> rolled_back
  paused --> rolled_back
```

Approval authorizes controlled testing, never an uncontrolled production switch.
