# Contributing

Use Go 1.24+ and Node 24. Install JavaScript dependencies with `npm ci`. Keep changes tenant-scoped, decimal-safe, UTC-based, and free of raw prompt or response persistence.

Before opening a pull request, run:

```bash
make fmt
make lint
make test
make test-integration
make test-e2e
make build
make check-api-client
docker compose config
```

Use Conventional Commits. Add migrations rather than editing an applied migration. Generated TypeScript contracts must come from `apps/api/openapi.yaml`. Report only checks that were actually run.
