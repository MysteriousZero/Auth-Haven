# Testing

## Current coverage

The checked-in tests concentrate on configuration, infrastructure connectivity, repositories, transactions, and Redis-backed repository caching.

| Area | Evidence in the repository | Current gap |
|---|---|---|
| Configuration | Environment loading, defaults, invalid-value fallback, PostgreSQL/Redis integration | No notable gap in the loader's current behavior |
| Test infrastructure | PostgreSQL and Redis Testcontainers, connection validation | CI execution is not documented or configured here |
| Tenant repository | Tenant and invitation CRUD, personal/organization cases | Service-level tenant rules are not covered |
| User repository | CRUD, status, role, password, last-login, duplicate email | Authorization rules are not covered |
| Auth repository | Sessions and refresh-token lifecycle | Full login/token workflows are not covered |
| Transactions | Commit and rollback | Complex multi-service failure paths remain untested |
| Redis cache wrappers | Hits, misses, population, invalidation, refresh tokens, basic performance comparison | Failure/degraded-mode behavior needs broader coverage |
| Services | No dedicated tests found | Authentication, registration, MFA, password, invitation, role, session, and audit rules |
| HTTP/gRPC handlers | No dedicated tests found | Request validation, status mapping, authentication, and contracts |
| Middleware/security | No dedicated tests found | JWT rejection, tenant isolation, rate limits, CORS, and attack cases |
| End-to-end workflows | No dedicated tests found | Registration through login, refresh, MFA, and logout |

This summary is based on test files present in the repository, not a measured coverage percentage. Run coverage tooling before making a quantitative claim.

## Running tests

Run the full suite from the repository root:

```bash
go test ./...
```

Repository and cache tests can be selected separately:

```bash
go test ./test/repositories/...
go test ./test/repositories/cache/...
```

To produce a coverage profile:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Many tests start PostgreSQL 15 Alpine and Redis 7 Alpine through Testcontainers. They require a working Docker daemon and may be slower on the first run while images are downloaded.

## Production-readiness priorities

1. Unit-test service-layer success and failure paths.
2. Exercise HTTP and gRPC contracts, including domain-error mapping.
3. Test access-token validation, tenant isolation, MFA encryption/verification, token rotation, and rate limits.
4. Add end-to-end registration, login, refresh, password reset, MFA, and logout scenarios.
5. Run the race detector and establish enforceable coverage thresholds in CI.

```bash
go test -race ./...
```

Historical test-result snapshots were removed because they described one machine's April 2026 run and conflicted with the current source. Test results should come from CI or a fresh local run.
