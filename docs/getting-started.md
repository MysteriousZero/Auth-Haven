# Getting started

## Prerequisites

- Go 1.25.1 or a compatible newer toolchain
- PostgreSQL 17 (the Compose configuration uses `postgres:17-alpine`)
- Redis
- Docker, if using Compose or the Testcontainers-based tests
- `protoc` and Go protobuf plugins only when changing `.proto` files

The VS Code dev container includes Go tooling, protobuf tools, Buf, Delve, and golangci-lint.

## 1. Configure the environment

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
```

Keep the generated MFA key stable if encrypted MFA records must remain readable. See [Configuration](configuration.md) for all settings.

## 2. Start dependencies

The root Compose file starts PostgreSQL only and obtains its `POSTGRES_*` settings from `.env`:

```bash
docker compose up -d db
```

Start Redis separately, for example:

```bash
docker run --name auth-haven-redis --rm -p 6379:6379 redis:7-alpine
```

The second command runs in the foreground. Existing PostgreSQL and Redis installations work as well.

## 3. Apply migrations

```bash
go run ./cmd/migrate
```

The migration command reads the same database environment variables and expects to run from the repository root so it can resolve `migrations/`.

> The migration directory currently contains both `0001_init_schema.up.sql` and later numbered schema migrations. Review the migration history for the target database before applying it; overlapping schema definitions may need consolidation.

## 4. Run the service

```bash
go run ./cmd/server
```

The defaults are:

- HTTP: `http://localhost:8080`
- gRPC: `localhost:50051`
- Health check: `GET http://localhost:8080/health`

Both transports run in one process. Failure of either server terminates the process.

## Container image

The root `Dockerfile` builds only the server binary. It does not run migrations and its runtime image exposes HTTP port 8080; map the gRPC port separately if needed. The root Compose file does not currently define an application service.
