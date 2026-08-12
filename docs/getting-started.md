# Getting started

## Prerequisites

- Go 1.25.1 or a compatible newer toolchain
- PostgreSQL 17 (the Compose configuration uses `postgres:17-alpine`)
- Redis
- Docker, if using Compose or the Testcontainers-based tests
- `protoc` and Go protobuf plugins only when changing `.proto` files

The VS Code dev container includes Go tooling, protobuf tools, Buf, Delve, and golangci-lint.

## Containerized development stack

From the repository root, one command generates local-only credentials, builds
the image, applies migrations, and starts PostgreSQL, Redis, HTTP, and gRPC:

```bash
./scripts/dev-up.sh
```

The generated `.env.development` is ignored by Git and has owner-only
permissions. Its passwords, MFA key, and ephemeral JWT signing key behavior are
for local development only; they are not production-safe defaults. Recreating
the application container invalidates access tokens issued by its previous
ephemeral signing key.

The stack waits for PostgreSQL and Redis health, runs `cmd/migrate` as a
one-shot service, and starts the application only after migrations succeed.
It exposes:

- HTTP: `http://localhost:8080`
- gRPC: `localhost:50051`
- Liveness: `GET http://localhost:8080/health`
- Readiness: `GET http://localhost:8080/ready`

Verify transport reachability, PostgreSQL persistence across restart, and
fail-closed readiness during a Redis outage with:

```bash
./scripts/dev-verify.sh
```

Stop containers while retaining database and Redis volumes with
`./scripts/dev-down.sh`. Use `./scripts/dev-down.sh --volumes` for a clean reset.
After adding migrations, rerun `./scripts/dev-up.sh`; Compose reruns the one-shot
migration service before the application starts.

The VS Code dev container composes the same PostgreSQL, Redis, and migration
services, then opens the repository in its development-tooling container. Run
the server from the devcontainer with `go run ./cmd/server`; the root Compose
`app` service is intentionally not started in that workflow.

## Manual development

Use the following steps when running the service directly instead of using the
containerized stack.

### 1. Configure the environment

The application reads environment variables directly; it does not load `.env` itself. Export them in your shell, use the dev-container environment, or invoke the process with an env loader.

At minimum, configure PostgreSQL and the MFA key:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=auth_haven
export DB_SSLMODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export MFA_ENCRYPTION_KEY="$(openssl rand -base64 32)"
export CORS_ALLOWED_ORIGINS="http://localhost:3000"
export SMTP_HOST="localhost"
export SMTP_PORT=1025
export EMAIL_FROM="auth-haven@localhost"
export PUBLIC_BASE_URL="http://localhost:3000"
```

Keep the generated MFA key stable if encrypted MFA records must remain readable. See [Configuration](configuration.md) for all settings.

### 2. Start dependencies

Generate the development environment and start PostgreSQL and Redis:

```bash
./scripts/dev-env.sh
docker compose --env-file .env.development up -d db redis
```

When running Go processes on the host, export equivalent `DB_HOST=localhost`
and `REDIS_HOST=localhost` values rather than the Compose service names in the
generated file. Existing PostgreSQL and Redis installations work as well.

### 3. Apply migrations

```bash
go run ./cmd/migrate
```

The migration command reads the same database environment variables and expects to run from the repository root so it can resolve `migrations/`.

The command applies the authoritative baseline and later forward migrations. Existing installations should follow the [migration operations guide](migrations.md) before upgrading.

### 4. Run the service

```bash
go run ./cmd/server
```

The defaults are:

- HTTP: `http://localhost:8080`
- gRPC: `localhost:50051`
- Health check: `GET http://localhost:8080/health`
- Readiness check: `GET http://localhost:8080/ready`
- Database pool metrics: `GET http://localhost:8080/metrics`

Both transports run in one process. Failure of either server terminates the process.

## Container image

The root `Dockerfile` builds both the server and migration binaries and exposes
HTTP port 8080 and gRPC port 50051. The image still defaults to the server;
Compose explicitly selects the migration binary for its one-shot service.
