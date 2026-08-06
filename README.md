# AI Operations Platform

AI Operations Platform is an open-source portfolio project for observing AI workloads and making evidence-based execution recommendations. The product name used in the interface is **AI Execution Advisor**.

## Current status

Phases 0–15 are implemented: the platform securely ingests, prices, analyzes, classifies, and recommends execution changes for organization-scoped LLM workloads. The public Next.js site and optimization tools are indexable, while the authenticated dashboard is protected and `noindex`. The private platform supports audited review, sanitized evaluations, local-model economics, and simulated controlled experiments. Experiments enforce quality, latency, error-rate, and cost guardrails with automatic rollback, concurrency-safe transitions, and a distinct verified-savings state.

## Planned product surfaces

- **Public website:** indexable product, documentation, comparison, blog, and educational optimization tools are implemented.
- **Private platform:** authentication, organization-scoped analytics, workload evidence, audited recommendation review, deterministic evaluations, local economics, and guarded simulated experiments are implemented; budgets and alerts follow in later phases.
- **Mobile companion:** concise operational views and authorized actions for alerts, recommendations, and experiments.

The platform will keep observed cost, estimated savings, evaluated outcomes, and verified savings distinct. Recommendations will begin with transparent deterministic rules and will never directly switch production execution.

## Planned architecture

The intended monorepo contains a Go API and workers, a Next.js web application, an Expo mobile application, generated TypeScript API bindings, PostgreSQL, and Redis. See [architecture](docs/architecture.md) and the [roadmap](docs/roadmap.md).

## Documentation

- [Product requirements](docs/product-requirements.md)
- [Architecture](docs/architecture.md)
- [Domain model](docs/domain-model.md)
- [Security model](docs/security-model.md)
- [Testing strategy](docs/testing-strategy.md)
- [Roadmap](docs/roadmap.md)

## Development

Install dependencies with `npm ci`, then use `make build`, `make test`, `make lint`, and `make test-e2e`. Development entry points are `make dev-api`, `make dev-web`, and `make dev-mobile`. PostgreSQL and Redis definitions can be validated with `make compose-config`.

## Limitations

The current public pricing and model comparison content is illustrative and is not live provider data. Evaluation execution currently supports deterministic mock candidates and a documented JSON Schema subset. Local throughput and hardware fields are user estimates, not measured benchmarks. Experiment routing is simulated and never proxies production traffic. Ollama execution, budgets, notifications, and the complete demo remain planned. Estimated savings remain distinct from verified experiment savings.

## License

MIT. See [LICENSE](LICENSE).
