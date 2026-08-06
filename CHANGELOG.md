# Changelog

## 0.1.0 - 2026-08-06

### Added

- Organization-scoped Go API for authentication, projects, workloads, redacted LLM ingestion, pricing, analytics, recommendations, evaluations, experiments, budgets, notifications, and local-model economics.
- Indexable Next.js product site, local prompt estimator, cost calculator, routing advisor, and authenticated operational dashboard.
- Expo operational companion with secure token storage and authorized alert, recommendation, and experiment actions.
- Deterministic five-workload demo with observed costs, estimated opportunities, guarded rollback, and verified savings.
- PostgreSQL/Redis Compose environment, generated OpenAPI client, Prometheus telemetry, trace correlation, full-stack CI, security automation, and operational documentation.

### Security

- Raw prompt and response storage is disabled; previews are redacted before persistence.
- Sessions, refresh tokens, API keys, tenant isolation, permissions, rate limits, and state transitions have automated coverage.

### Known limitations

- Provider prices and model names are illustrative.
- Provider-backed evaluation adapters and Expo push require deployment-specific credentials.
- Experiment traffic routing and local-model throughput are simulated or estimated.
