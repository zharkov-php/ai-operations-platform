# Testing strategy

## Principles

Tests follow risk and domain boundaries. Every phase must be formatted, linted, built, and tested at the narrowest useful level before commit. A failed required check blocks the phase. Test output is reported exactly; missing tools or skipped suites are not described as passing.

## Planned layers

- Go table-driven unit tests for pricing, recommendations, confidence, break-even, redaction, and state machines.
- Testcontainers-backed repository, migration, API, tenant-isolation, concurrency, and idempotency tests where Docker is available.
- Vitest and Testing Library for web logic and accessible components.
- Playwright for public tools, authentication, dashboard, review, experiment, alert, SEO/noindex, and demo journeys.
- Jest and React Native Testing Library for mobile navigation, API state, secure storage, permissions, and offline behavior.
- OpenAPI generation freshness checks to prevent duplicated or stale TypeScript contracts.

## End-to-end policy

E2E suites begin once an executable user flow exists. Documentation-only Phase 0 uses consistency, link, secret, and Git checks; it cannot truthfully claim application E2E coverage. Later phases extend the cumulative E2E journey only for behavior that exists and keep external provider credentials optional.

## Release gate

The release requires all unit, integration, web, mobile, contract, Docker, security, and deterministic demo checks to pass from a clean state. Flaky tests are defects and are not hidden through unconditional retries.
