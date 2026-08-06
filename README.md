# AI Operations Platform

AI Operations Platform is an open-source portfolio project for observing AI workloads and making evidence-based execution recommendations. The product name used in the interface is **AI Execution Advisor**.

## Current status

Phases 0–8 are implemented: the platform securely ingests, prices, analyzes, classifies, and recommends execution changes for organization-scoped LLM workloads. Versioned deterministic rules expose reason codes, evidence, confidence inputs, estimated savings/cost/break-even, quality and operational risk, and required next action. Critical or insufficient-evidence workloads explicitly keep their current model. Recommendations never claim verified savings or apply production changes. Evaluations, experiments, and complete demo data are not implemented yet.

## Planned product surfaces

- **Public website:** indexable product, documentation, comparison, blog, and educational calculator routes.
- **Private platform:** organization-scoped cost analytics, workload evidence, recommendations, evaluations, experiments, budgets, and alerts.
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

Install dependencies with `npm ci`, then use `make build`, `make test`, and `make lint`. Development entry points are `make dev-api`, `make dev-web`, and `make dev-mobile`. PostgreSQL and Redis definitions can be validated with `make compose-config`; application connectivity is introduced in Phase 2.

## Limitations

Everything beyond the monorepo foundation is planned. Pricing, model quality, recommendations, savings, authentication, and production integrations are not yet available.

## License

MIT. See [LICENSE](LICENSE).
