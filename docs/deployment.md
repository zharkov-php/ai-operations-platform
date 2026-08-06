# Deployment

`make demo` builds and starts local web/API containers with PostgreSQL and Redis, applies migrations, and seeds deterministic illustrative data. `make demo-reset` removes only this Compose project and its named volume.

Production should supply secrets through a secret manager, terminate TLS, restrict data networks, persist PostgreSQL backups, configure retention, export telemetry to an approved collector, and run migrations as a controlled pre-deploy job. Kubernetes and registry publication are intentionally outside version 0.1.0.
