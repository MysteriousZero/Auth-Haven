# Auth Haven documentation

## Documentation map

- [Project overview](overview.md) — purpose, capabilities, technology, and implementation status
- [Architecture](architecture.md) — layers, request flow, data stores, and source layout
- [Data model](data-model.md) — tables, relationships, invariants, and migration warning
- [Migration operations](migrations.md) — authoritative baseline, legacy rebuilds, and rollback safety
- [Security architecture](security.md) — authentication, isolation, validation, audit, and cache invariants
- [Getting started](getting-started.md) — prerequisites, local infrastructure, migrations, and startup
- [Configuration](configuration.md) — environment variables and production-sensitive settings
- [API guide](api.md) — implemented REST routes and gRPC services
- [Development guide](development.md) — tests, protobuf generation, and contribution workflow
- [Testing](testing.md) — current coverage, commands, and production-readiness gaps
- [Operations runbook](runbook.md) — dependency health, degraded modes, and incident checks
- [AI-assisted development](ai-development.md) — context map, safe change patterns, and agent handoff rules

Repository-level governance and contributor documents:

- [`PROJECT.md`](../PROJECT.md) — mission, scope, status, and roadmap
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — contribution workflow and review expectations
- [`SECURITY.md`](../SECURITY.md) — vulnerability reporting and secure-development rules
- [`SUPPORT.md`](../SUPPORT.md) — support channels and useful bug reports
- [`CODE_OF_CONDUCT.md`](../CODE_OF_CONDUCT.md) — community standards and enforcement
- [`LICENSE`](../LICENSE) — Apache License 2.0 terms
- [`AGENTS.md`](../AGENTS.md) — mandatory repository-wide instructions for coding agents

These documents describe current behavior and explicitly label plans or known
gaps. Executable code, migrations, and generated contracts are authoritative
when documentation drifts.
