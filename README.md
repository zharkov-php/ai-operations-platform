# AI Operations Platform

AI Operations Platform is an open-source portfolio project for observing AI workloads and making evidence-based execution recommendations. The product name used in the interface is **AI Execution Advisor**.

## Current status

Phases 0–21 are implemented: the platform securely ingests, prices, analyzes, classifies, and recommends execution changes for organization-scoped LLM workloads. The deterministic demo seeds five representative workloads, 750 usage records, expected recommendations, a budget alert, a rollback, and verified savings. The private platform supports audited review, evaluations, local-model economics, guarded experiments, budget governance, and notifications. GitHub Actions validates API, web, mobile, contracts, browser E2E, containers, dependencies, and secrets. The Expo companion provides secure operational actions and explicit permission and offline states.

## Product surfaces

- **Public website:** indexable product, documentation, comparison, blog, and educational optimization tools are implemented.
- **Private platform:** authentication, analytics, workload evidence, audited review, deterministic evaluations, local economics, guarded experiments, budgets, and alerts are implemented.
- **Mobile companion:** concise operational views and authorized actions for alerts, recommendations, and experiments.

The platform keeps observed cost, estimated savings, evaluated outcomes, and verified savings distinct. Recommendations use transparent deterministic rules and never directly switch production execution.

## Architecture

The monorepo contains a Go API and workers, a Next.js web application, an Expo mobile application, generated TypeScript API bindings, PostgreSQL, and Redis. See [architecture](docs/architecture.md) and the [roadmap](docs/roadmap.md).

## Documentation

- [Product requirements](docs/product-requirements.md)
- [Architecture](docs/architecture.md)
- [Domain model](docs/domain-model.md)
- [Security model](docs/security-model.md)
- [Privacy model](docs/privacy-model.md)
- [Cost methodology](docs/cost-methodology.md)
- [Recommendation engine](docs/recommendation-engine.md)
- [Evaluation engine](docs/evaluation-engine.md)
- [Experiment guardrails](docs/experiment-guardrails.md)
- [API authentication](docs/api-authentication.md)
- [Deployment](docs/deployment.md)
- [Operations](docs/operations.md)
- [Testing strategy](docs/testing-strategy.md)
- [Roadmap](docs/roadmap.md)

## Development

Run `make demo`, then sign in at `http://localhost:3000` with the printed illustrative credentials. Install dependencies with `npm ci`, then use `make build`, `make test`, `make lint`, and `make test-e2e`. Development entry points are `make dev-api`, `make dev-web`, and `make dev-mobile`.

## Limitations

Public pricing and model comparison content is illustrative and is not live provider data. Evaluation execution supports deterministic mock candidates and a documented JSON Schema subset. Local throughput is estimated, experiment routing is simulated, Expo push requires deployment credentials, and budget anomaly detection uses an explicit two-times prior-daily-average rule rather than statistical anomaly modeling. Estimated savings remain distinct from verified experiment savings.

## License

MIT. See [LICENSE](LICENSE).
