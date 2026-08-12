# Project

## Mission

Auth Haven aims to provide a reusable identity boundary for personal and organization accounts. Its design prioritizes tenant isolation, explicit authentication flows, revocable credentials, auditable mutations, and transport-independent business rules.

## Project status

The repository contains a substantial implementation, but it remains a development project rather than a production-ready identity provider. The implementation is strongest in PostgreSQL repositories and their integration tests. Service, transport, middleware, security, and end-to-end test coverage are still limited.

Current behavior is determined in this order:

1. Executable code and registered server wiring
2. Database migrations and generated protobuf contracts
3. Documentation under `docs/`, with planned behavior explicitly labeled

When these disagree, treat the disagreement as drift to resolve—not as permission to silently choose one side.

## Design principles

- **Tenant isolation:** every tenant-scoped read and mutation must prove its tenant boundary.
- **Transport independence:** business rules live in services, not Gin handlers or gRPC adapters.
- **Dependency inversion:** services use interfaces from `internal/domain/interfaces`.
- **Secure credential storage:** passwords are one-way hashed; opaque tokens are stored as hashes; MFA secrets are encrypted.
- **Anti-enumeration:** authentication and password-reset behavior must not reveal whether an account exists.
- **Auditable changes:** security-relevant mutations carry trace, actor, client, and tenant context.
- **Documentation discipline:** contract changes update code, generated artifacts, and documentation together.

## In scope

- Tenant and user lifecycle
- Authentication, refresh, logout, and sessions
- MFA and password recovery
- Roles and permissions
- Invitations and organization onboarding
- Security audit records
- HTTP and gRPC service interfaces

## Out of scope for the current repository

- A user-facing web or mobile client
- A production email delivery integration
- Hosted infrastructure definitions
- A complete OAuth 2.0 or OpenID Connect provider
- A published backwards-compatibility or release-support policy

## Known gaps

- Several declared gRPC AuthService methods fall through to `Unimplemented`; the refresh handler method is incorrectly named `Refresh` instead of `RefreshToken`.
- The gRPC validation interceptor is a pass-through.
- HTTP CORS currently permits every origin.
- Email delivery is constructed with empty SMTP configuration.
- Role services are not wired to HTTP routes.
- Token TTL configuration is not consistently consumed by token-generation code.
- Database pool settings are loaded but not applied to `database/sql`.
- The migration directory contains potentially overlapping initial/schema migrations and needs consolidation before a clean deployment workflow can be guaranteed.
- Automated tests do not yet adequately cover services, transports, middleware, security, or end-to-end workflows.

## Roadmap

Use the [Auth Haven Roadmap](https://github.com/orgs/MysteriousZero/projects/2)
for current priorities and the
[release milestones](https://github.com/MysteriousZero/auth-haven/milestones)
for planned version scope. Roadmap entries are planning targets, not release
commitments.
