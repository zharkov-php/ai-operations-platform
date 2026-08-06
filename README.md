# AI Operations Platform

AI Operations Platform is an open-source portfolio project for observing AI workloads and making evidence-based execution recommendations. The product name used in the interface is **AI Execution Advisor**.

## Current status

Phases 0–3 are implemented: the architecture, monorepo, Go service foundation, and organization-scoped authentication exist. Authentication includes bcrypt password hashes, account lockout, secure web cookies, short-lived signed mobile access tokens, rotating hashed refresh tokens, roles, and authentication audits. Ingestion, analytics, recommendations, and complete demo data are not implemented yet.

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
