# Security model

## Trust boundaries

Public pages are unauthenticated and cannot access tenant data. Browser sessions use secure HTTP-only cookies with CSRF protection. Mobile clients use short-lived access tokens and rotating refresh tokens held through SecureStore. Ingestion clients use scoped API keys restricted to an organization and optionally a project.

## Credential handling

Passwords use a modern adaptive hash. API keys and refresh tokens are stored only as hashes; an identifying prefix is retained and the full secret is shown once. Revocation, rotation, last-used timestamps, authentication rate limits, and account lockout are implemented. Authorization headers, cookies, passwords, API keys, bearer tokens, and private-key markers are removed from logs and persisted previews.

## Data minimization

Raw prompt and response persistence is disabled by default. Redaction occurs before persistence and supports configurable sensitive field names plus optional email and phone removal. Prompt hashing enables exact duplicate analysis without revealing content. Retention jobs and audited privileged actions limit exposure.

## Application controls

- Validate and size-limit every external payload.
- Parameterize every database query.
- Scope every private operation by organization and role/scope.
- Apply secure cookies, CSP, common security headers, and safe error responses.
- Rate-limit authentication and ingestion.
- Never place sensitive values in metrics or high-cardinality labels.
- Keep dashboard and internal API routes out of search indexes.

## Threat priorities

The first release prioritizes tenant escape, credential theft, prompt leakage, authorization bypass, ingestion abuse, replay/idempotency failures, unsafe state transitions, injection, and accidental secret logging. Automated tests and CI scanning must cover these controls before release.

## External identity

The project will not implement an OAuth provider. Production deployments may integrate an established OIDC provider such as Keycloak while retaining the platform's organization and role authorization model.
