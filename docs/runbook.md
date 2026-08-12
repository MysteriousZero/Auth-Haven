# Operations runbook

## Probes

- `GET /health` is a process liveness probe and does not contact dependencies.
- `GET /ready` checks PostgreSQL and Redis with a two-second timeout. PostgreSQL is always required. Redis is required when rate limiting is configured fail-closed.
- `GET /metrics` exposes current `database/sql` maximum, open, in-use, idle, and wait-count gauges in Prometheus text format.

## Redis outage

The default policy keeps cache-backed reads available through PostgreSQL and fails rate-limited endpoints closed with `503 Service Unavailable`. Do not enable `REDIS_RATE_LIMIT_FAIL_OPEN` in production without a documented temporary risk acceptance and an alternative edge rate limit. A deliberately fail-open response carries `X-RateLimit-Status: degraded`.

## Email delivery

Confirm SMTP reachability, sender authorization, and credentials without logging their values. Password-reset delivery failures appear in server logs while the public endpoint preserves its anti-enumeration response. Invitation operations return delivery errors to authenticated callers so they can retry. Raw reset and invitation tokens must never be copied into tickets or logs.

## Database pool pressure

Monitor `auth_haven_db_open_connections`, `auth_haven_db_in_use_connections`, and `auth_haven_db_wait_count`. A rising wait count with in-use connections near `auth_haven_db_max_open_connections` indicates pool saturation. Tune pool settings against the PostgreSQL connection budget across all replicas, leaving capacity for migrations and operator access.
