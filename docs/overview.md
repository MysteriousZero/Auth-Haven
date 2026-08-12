# Project overview

## What Auth Haven does

Auth Haven is an identity service designed around tenant isolation. A user belongs to a tenant and may have a role. The service manages registration, authentication, MFA, passwords, sessions, devices, invitations, and audit records.

The repository uses documentation-driven development: `docs/` records current behavior, architecture, constraints, and explicitly marked plans, while Go packages under `internal/` contain the implementation. Executable wiring remains authoritative when documentation drifts.

## Implemented HTTP capabilities

- Individual, organization-domain, and invitation-based registration
- Password login, optional TOTP verification, refresh-token rotation, and logout
- Password reset requests, resets, and authenticated password changes
- TOTP enrollment, activation, listing, and disabling
- Session and device listing/revocation
- Tenant invitation creation, listing, revocation, and resend
- User and tenant audit-log queries
- Redis-backed login, password-reset, and MFA rate limits
- Request IDs, request logging, CORS, and JWT authentication middleware

Role services exist in the service/repository layers, but no HTTP role routes are registered. SMS-related MFA methods also exist in service code but are not exposed by the current HTTP router.

## Technology

| Concern | Choice |
|---|---|
| Language | Go 1.25.1 (`go.mod`) |
| HTTP | Gin |
| RPC | gRPC and Protocol Buffers |
| Database | PostgreSQL |
| Cache/rate limiting | Redis |
| Migrations | golang-migrate |
| Tokens | JWT access/temp tokens and opaque refresh tokens |
| Password hashing | Argon2id |
| MFA | TOTP with AES-256-GCM storage encryption |
| Testing | Go test and Testcontainers |

## Core domain objects

- **Tenant**: isolation boundary for users, roles, invitations, and audit records.
- **User**: tenant member with credentials, status, and an optional role.
- **Role and permission**: tenant-scoped authorization metadata.
- **Session and device**: authenticated activity and client-device tracking.
- **Refresh token**: revocable, rotated credential stored as a hash.
- **MFA method**: encrypted TOTP secret or phone metadata.
- **Invitation**: tenant/role assignment offered to an email address.
- **Audit log**: security event associated with a user and tenant.

## Current maturity

This is an active implementation rather than a production-ready release. Repository tests are the strongest-covered area; service, handler, middleware, security, and end-to-end coverage remain incomplete. Review the [testing guide](testing.md), [security architecture](security.md), and [`PROJECT.md`](../PROJECT.md) before production use.
