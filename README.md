# AI Operations Platform

AI Operations Platform is an open-source portfolio project for observing AI workloads and making evidence-based execution recommendations. The product name used in the interface is **AI Execution Advisor**.

## Current status

Phases 0–12 are implemented: the platform securely ingests, prices, analyzes, classifies, and recommends execution changes for organization-scoped LLM workloads. The public Next.js site and optimization tools are indexable, while the authenticated dashboard is protected and `noindex`. The dashboard uses a generated OpenAPI client for analytics, drill-down, redacted call exploration, and recommendation review. Authorized users may accept a recommendation for evaluation or reject it with a required reason; atomic transitions, duplicate-action protection, tenant isolation, permissions, and audit history are enforced. Acceptance never changes production execution and savings remain estimated. Evaluation and experiment workflows remain planned.

## Planned product surfaces

- **Public website:** indexable product, documentation, comparison, blog, and educational optimization tools are implemented.
- **Private platform:** authentication, organization-scoped analytics, workload evidence, and audited recommendation review are implemented; evaluations, experiments, budgets, and alerts follow in later phases.
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

The current public pricing and model comparison content is illustrative and is not live provider data. The prompt estimator uses a documented approximation rather than a provider tokenizer. Evaluation execution, experiments, budgets, notifications, and complete demo remain planned. No recommendation automatically changes production execution, and no estimated savings are presented as verified.

## License

MIT. See [LICENSE](LICENSE).
