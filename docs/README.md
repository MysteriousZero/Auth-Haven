# Auth Haven documentation

Auth Haven is a documentation-driven, multi-tenant authentication and authorization service written in Go. It exposes HTTP/JSON and gRPC transports over shared domain services and persists identity data in PostgreSQL. Redis supports HTTP rate limiting and selected repository caches.

## Documentation map

- [Project overview](overview.md) — purpose, capabilities, technology, and implementation status
- [Architecture](architecture.md) — layers, request flow, data stores, and source layout
- [Data model](data-model.md) — tables, relationships, invariants, and migration warning
- [Security architecture](security.md) — authentication, isolation, validation, audit, and cache invariants
- [Getting started](getting-started.md) — prerequisites, local infrastructure, migrations, and startup
- [Configuration](configuration.md) — environment variables and production-sensitive settings
- [API guide](api.md) — implemented REST routes and gRPC services
- [Development guide](development.md) — tests, protobuf generation, and contribution workflow
- [Testing](testing.md) — current coverage, commands, and production-readiness gaps
- [AI-assisted development](ai-development.md) — context map, safe change patterns, and agent handoff rules

Repository-level governance and contributor documents:

- [`PROJECT.md`](../PROJECT.md) — mission, scope, status, and roadmap
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — contribution workflow and review expectations
- [`SECURITY.md`](../SECURITY.md) — vulnerability reporting and secure-development rules
- [`SUPPORT.md`](../SUPPORT.md) — support channels and useful bug reports
- [`AGENTS.md`](../AGENTS.md) — mandatory repository-wide instructions for coding agents

The `docs/` directory is the single documentation source. It describes current behavior and explicitly labels planned requirements or known gaps. Executable code and migrations remain authoritative when documentation drifts.

## At a glance

```text
HTTP client ──> Gin middleware ──> handlers ──> services ──> repositories ──> PostgreSQL
                    │                                          │
                    └── JWT auth and Redis rate limits          └── Redis cache

gRPC client ──> interceptors ──> gRPC handler ──> shared services/repositories
```

Default listeners are HTTP `:8080` and gRPC `:50051`.
