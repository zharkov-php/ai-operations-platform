# AI Operations Platform

AI Operations Platform is an open-source portfolio project for observing AI workloads and making evidence-based execution recommendations. The product name used in the interface is **AI Execution Advisor**.

## Current status

Phase 0 is implemented: product requirements, architecture, domain, security, testing, and roadmap documentation are defined. No runtime application, API, dashboard, mobile app, ingestion pipeline, recommendation engine, or demo is implemented yet.

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

Runtime setup commands will be added in Phase 1. The repository currently has no buildable application and no demo credentials.

## Limitations

Everything beyond the Phase 0 documentation is planned. Pricing, model quality, recommendations, savings, authentication, and production integrations are not yet available.

## License

MIT. See [LICENSE](LICENSE).
