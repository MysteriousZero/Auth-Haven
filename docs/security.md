# Security architecture

This document combines the project's security, validation, caching, JWT, and audit invariants. It distinguishes current mechanisms from requirements that still need enforcement.

## Current mechanisms

- Password hashing uses Argon2id.
- Access and temporary tokens use Ed25519-signed JWTs.
- Refresh, reset, and invitation credentials are generated with secure randomness and stored as hashes.
- TOTP secrets are encrypted with AES-256-GCM using a configured 32-byte key.
- HTTP authentication validates bearer tokens and injects user, tenant, and role claims.
- gRPC has a bearer-token unary interceptor for protected methods.
- Redis supports selected cache-aside repositories and endpoint rate limits.
- Request IDs and client metadata are propagated for audit records.

Remaining gaps—including incomplete gRPC validation and broader security-test coverage—are tracked in [`PROJECT.md`](../PROJECT.md#known-gaps).

## Authentication and token rules

- Access tokens are short-lived JWTs containing user and tenant identity.
- Refresh tokens are opaque, long-lived, hashed in storage, checked for revocation/expiry, and rotated on use.
- MFA challenge tokens are short-lived and signed rather than stored as server-side challenge state.
- Protected HTTP requests send `Authorization: Bearer <token>`; logout also sends `X-Session-ID`.
- Missing, malformed, expired, revoked, or replayed credentials must fail closed.
- Production instances fail startup without a stable signing key. All replicas share the active key and identify it with a JWT `kid` header.
- Verification keys may overlap during planned rotation; removing a key ID immediately invalidates JWTs signed by it. See the [configuration guide](configuration.md#signing-key-rotation).
- Access, refresh, temporary-token, and session lifetimes come from validated configuration. JWT validation permits only the configured clock-skew window.

## Authorization and tenant isolation

- Never authorize a request solely from a path, body, metadata, or query tenant ID.
- Validate the actor from trusted token claims, then prove actor permission and target ownership in the service layer.
- Roles and permissions are tenant-scoped; a role from one tenant must never be assigned in another.
- Personal accounts must not be exposed through organization tenant listing or administration paths.
- Cross-tenant failures should avoid revealing whether the target resource exists.

## Input validation

Transport adapters perform structural validation; services perform validation requiring state or business context.

Common boundary rules include:

- UUID formatting for tenant, user, session, device, MFA, and invitation identifiers
- Email format and bounded string lengths
- Positive integer role IDs
- Six-digit TOTP codes
- Password length and complexity validation before hashing

HTTP handlers use `go-playground/validator`. The current gRPC validation interceptor is a placeholder, so handlers must not assume protobuf input has been validated.

Domain checks include tenant-scoped email uniqueness, public-domain restrictions for organization signup, actor/target tenant membership, resource ownership, invitation state/expiry, and MFA enrollment state.

## Sensitive data handling

Never log or include in audit metadata:

- Plaintext passwords or password hashes
- Raw access, refresh, reset, invitation, or MFA challenge tokens
- MFA codes, decrypted TOTP secrets, encryption keys, or signing keys
- Unnecessary email addresses or phone numbers
- Database/Redis passwords or complete `.env` contents

Use stable IDs in audit events when an authorized lookup can resolve identity later.

## Audit requirements

Security-relevant events should be recorded with tenant, actor when known, action, UTC timestamp, trace ID, client IP, and user agent. Relevant categories include authentication success/failure, logout, credential changes, MFA changes, registration, user/tenant status changes, role changes, and invitation lifecycle events.

Read-only operations are generally not audited unless policy requires it. Audit writes must not leak sensitive input. Retention/archival is not implemented by this repository and must be designed before making compliance claims.

## Cache and rate-limit rules

- Cache-aside reads must fall back to PostgreSQL when cached data is absent.
- Authentication-state mutations must invalidate related session/token entries.
- Tenant updates must invalidate ID and domain lookups.
- Every Redis key should use the `auth:` namespace.
- Cache failures default to fail-open database access. Rate-limit failures default to fail-closed `503` responses and may only fail open through explicit configuration; degraded responses include `X-RateLimit-Status: degraded`.
- Production CORS uses an exact configured origin allowlist. Wildcard origins and wildcard credential combinations fail configuration validation.
- Reset and invitation credentials are URL-escaped before delivery and are never written to application logs.

## Security review checklist

For changes to credentials, sessions, tenants, roles, MFA, invitations, or audit logs, test:

- Cross-tenant and wrong-owner requests
- Missing, malformed, expired, revoked, and replayed credentials
- Enumeration-resistant responses
- Transaction rollback and concurrency behavior
- Cache invalidation after persistent state changes
- Redaction in responses, logs, errors, and audit records
- HTTP and gRPC parity where both transports expose the operation
