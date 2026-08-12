# Auth Haven

Auth Haven is a documentation-driven, multi-tenant authentication and authorization service written in Go. It offers HTTP/JSON and gRPC interfaces backed by PostgreSQL, with Redis used for caching and request rate limiting.

> Auth Haven is under active development. It is not currently documented as production-ready; review [known gaps](PROJECT.md#known-gaps) and the [testing guide](docs/testing.md) before deployment.

## Capabilities

- Individual, organization-domain, and invitation-based registration
- Password authentication with JWT access tokens and rotating refresh tokens
- TOTP multi-factor authentication
- Password reset and authenticated password changes
- Session and device management
- Tenant roles, invitations, and audit logging
- PostgreSQL persistence and Redis-backed caching/rate limits
- HTTP and gRPC transports over shared domain services

## Quick start

The recommended environment is the included VS Code Dev Container. For a local shell, install Go 1.25.1 or newer, Docker, PostgreSQL, and Redis.

```bash
# Start the PostgreSQL service declared by this repository.
docker compose up -d db

# Supply a stable 32-byte MFA encryption key.
export MFA_ENCRYPTION_KEY="$(openssl rand -base64 32)"

# Configure DB_* and REDIS_* variables, then:
go run ./cmd/migrate
go run ./cmd/server
```

The default endpoints are HTTP `http://localhost:8080`, gRPC `localhost:50051`, and health check `GET /health`. The root Compose file starts PostgreSQL only; see the [complete setup guide](docs/getting-started.md) before running the commands above.

## Documentation

| Document | Purpose |
|---|---|
| [Project](PROJECT.md) | Scope, status, principles, and roadmap |
| [Documentation index](docs/README.md) | All implementation and operations guides |
| [Architecture](docs/architecture.md) | Components, dependency flow, and source layout |
| [Data model](docs/data-model.md) | Tables, relationships, and migration constraints |
| [Security architecture](docs/security.md) | Authentication, isolation, validation, audit, and caching rules |
| [API guide](docs/api.md) | Currently registered HTTP routes and gRPC methods |
| [Configuration](docs/configuration.md) | Environment variables and security-sensitive settings |
| [Development](docs/development.md) | Local workflow and protobuf generation |
| [Testing](docs/testing.md) | Existing coverage and commands |
| [AI development](docs/ai-development.md) | Task workflow and context map for coding agents |
| [AI agent rules](AGENTS.md) | Canonical repository instructions for coding agents |
| [Contributing](CONTRIBUTING.md) | Contribution and review requirements |
| [Security](SECURITY.md) | Vulnerability reporting and security expectations |
| [Support](SUPPORT.md) | Help channels and bug-report guidance |

## Repository layout

```text
api/proto/       Protobuf source contracts
cmd/             Server and migration executables
docs/            Current implementation and operating guides
internal/        Private domain, service, repository, and transport code
migrations/      PostgreSQL migrations
pkg/proto/       Generated protobuf code
test/            Testcontainers, configuration, repository, and cache tests
```

## Development

```bash
gofmt -w path/to/changed.go
go vet ./...
go test ./...
```

Integration tests require a working Docker daemon. Changes to behavior should update documentation and tests together when applicable. See [CONTRIBUTING.md](CONTRIBUTING.md) for the full checklist.

## License

No license file is currently present. Unless a license is added, normal copyright restrictions apply.
