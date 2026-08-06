# ADR 0001: Modular monolith

Status: accepted.

The platform uses one Go service with explicit domain packages, PostgreSQL, and Redis. This keeps transactions, tenant scoping, migrations, and local operation reviewable. Microservices, Kafka, GraphQL, and additional databases are deferred until measured constraints require them.
