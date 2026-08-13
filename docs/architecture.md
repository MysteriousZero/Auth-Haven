# Architecture

## Layered design

```text
cmd/server
  ├── HTTP server (Gin)
  │     ├── middleware: request ID, logging, CORS, JWT, rate limits
  │     └── handlers: transport parsing and HTTP error responses
  └── gRPC server
        ├── interceptors: validation placeholder, JWT, logging
        └── shared gRPC handler
                 │
                 v
            service layer
      authentication, registration, passwords,
      MFA, sessions, invitations, roles, audit
                 │
                 v
          repository interfaces
                 │
          PostgreSQL + Redis cache
```

The dependency direction is transport → service → domain interfaces → repository implementation. Domain models, interfaces, and errors do not depend on Gin or gRPC.

## Request lifecycle

For HTTP, `internal/server/http.go` constructs dependencies and registers routes. Global middleware adds a trace ID, logs the request, and sets CORS headers. Protected groups additionally validate an access token and place `user_id`, `tenant_id`, and `role_id` in the Gin context. Handlers validate transport input, enrich the Go context with client metadata, call a service, and map domain errors to HTTP responses.

For gRPC, `internal/server/grpc.go` registers AuthService, UserService, and
SessionService. Auth login, MFA verification, refresh-token rotation, password
reset, registration, and session operations delegate to the same domain
services used by HTTP. Unary interceptors log calls and authenticate every
SessionService method; AuthService and UserService methods are public. The
validation interceptor remains a placeholder, so handlers validate equivalent
domain request models before invoking services. Authentication claims are
injected into `context.Context` for handlers and services.

## Persistence and caching

PostgreSQL is the system of record. Repository implementations use `database/sql`; transaction support is passed through context. Migrations define tenants, users, roles, permissions, sessions, devices, tokens, MFA methods, resets, invitations, and audit logs.

Redis is used by HTTP rate limiting and cached tenant/auth repositories. Cache outages degrade to PostgreSQL; rate-limit outages follow an independent configured policy. Redis is a required startup and readiness dependency when rate limiting is fail-closed.

The HTTP server exposes liveness at `/health`, dependency-aware readiness at `/ready`, and `database/sql` pool statistics at `/metrics`.

## Security model

- Passwords are hashed with Argon2id.
- Access tokens are signed JWTs; an Ed25519 key is generated at startup when no configured key is provided. That fallback is unsuitable for multiple instances or restarts because old tokens cannot be validated by a new key.
- Refresh, password-reset, and invitation tokens are persisted as hashes.
- MFA secrets require a 32-byte encryption key supplied as base64.
- Tenant identity is carried in tokens and request data; authorization decisions belong in the service layer.

See the [security architecture](security.md) for enforced mechanisms, invariants, and known gaps.

## Source layout

| Path | Responsibility |
|---|---|
| `cmd/server` | Starts HTTP and gRPC servers |
| `cmd/migrate` | Applies database migrations |
| `api/proto` | Source protobuf contracts |
| `pkg/proto` | Generated protobuf Go code |
| `internal/config` | Environment configuration |
| `internal/domain` | Models, interfaces, and domain errors |
| `internal/handlers` | HTTP and gRPC adapters |
| `internal/middleware` | JWT, validation, and rate limiting |
| `internal/service` | Authentication business rules |
| `internal/repository` | PostgreSQL and Redis access |
| `internal/server` | Dependency wiring and route registration |
| `migrations` | SQL schema changes |
| `test` | Unit/integration tests and containers |
| `docs` | Current behavior, architecture, operational guidance, and explicitly marked plans |
