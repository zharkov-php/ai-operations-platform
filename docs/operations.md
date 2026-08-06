# Operations

The API exposes `/health/live`, dependency-aware `/health/ready`, and Prometheus metrics at `/metrics`. HTTP metrics use only method, normalized route template, and status labels; tenant, user, call, trace, and prompt identifiers are intentionally excluded.

Requests accept a W3C `traceparent` header and copy its trace ID into structured logs. When absent, the API creates a correlation trace ID. Logs must never contain authorization headers, cookies, prompts, responses, passwords, API keys, or push tokens.

PostgreSQL connections are bounded to 20 with two warm connections and a 30-minute lifetime. Redis operations use a 3-second connection timeout and 2-second read/write timeouts. HTTP read-header, write, idle, readiness, and graceful-shutdown deadlines prevent stalled dependencies from exhausting resources.

Notification delivery records expose attempts, next-attempt time, and terminal failure after five tries. The initial in-app channel is transactional. The development console adapter logs only event type and deep link. A production worker should claim pending deliveries with `FOR UPDATE SKIP LOCKED`, use the deterministic backoff policy, and retain failed records for operator inspection. Expo push remains disabled until credentials and device-token encryption are configured.

Operational response:

1. Check readiness to distinguish API health from PostgreSQL or Redis failure.
2. Correlate the request ID and trace ID; never request prompt content for diagnosis.
3. Inspect bounded error-rate and latency metrics by route.
4. Pause experiments before rollback when the guardrail state permits; rollback immediately for an active violation.
5. Preserve failed notification delivery records and retry only after the downstream channel recovers.
