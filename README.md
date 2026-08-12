# Auth Haven

Auth Haven is an open-source, multi-tenant authentication and authorization
service written in Go. It provides HTTP/JSON and gRPC interfaces over shared
identity services, uses PostgreSQL as its system of record, and uses Redis for
selected caches and request rate limiting.

The project is being developed toward an India-first enterprise identity
platform with regional privacy controls and standards-based federation for
broader international use.

> Auth Haven is under active development and is not yet production-ready.

## What Auth Haven provides

- Personal, organization-domain, and invitation-based registration
- Password authentication with signed access tokens and rotating refresh tokens
- TOTP multi-factor authentication and password recovery
- Session and device management
- Tenant-scoped roles, invitations, and audit records
- HTTP/JSON and gRPC transports over shared domain services
- Security-oriented tenant isolation and revocable credential design

## Project information

- [Wiki](https://github.com/MysteriousZero/auth-haven/wiki) — approachable
  guides to the project, architecture, setup, security, and development
- [Releases](https://github.com/MysteriousZero/auth-haven/releases) —
  version-specific changes, artifacts, and release notes
- [Roadmap](https://github.com/orgs/MysteriousZero/projects/2) — planned work,
  priorities, phases, risks, and release milestones
- [Milestones](https://github.com/MysteriousZero/auth-haven/milestones) —
  release scope and progress
- [API reference](docs/reference/OpenAPI3.yaml) — OpenAPI contract
- [Canonical documentation](docs/README.md) — detailed implementation and
  operating references maintained with the source
- [Security policy](SECURITY.md) — vulnerability reporting and security
  expectations
- [Support](SUPPORT.md) — help channels and useful diagnostic information
- [Contributing](CONTRIBUTING.md) — development setup, checks, review, and pull
  request requirements

For current limitations and implementation status, see
[PROJECT.md](PROJECT.md). Executable code, migrations, and generated contracts
remain authoritative when documentation drifts.

## Community and license

Participation is governed by the [Code of Conduct](CODE_OF_CONDUCT.md).
Auth Haven is licensed under the [Apache License 2.0](LICENSE).
