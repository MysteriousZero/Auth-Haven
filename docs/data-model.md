# Data model

PostgreSQL is the system of record. [`000001_init_schema.up.sql`](../migrations/000001_init_schema.up.sql) is the authoritative baseline. All later changes are forward-only, numerically ordered migrations.

## Relationships

```text
Tenant
  ├── Roles ──> Role permissions
  ├── Users ──> MFA methods
  │     ├── Sessions ──> Device
  │     ├── Refresh tokens
  │     └── Password resets
  ├── Invitations ──> Role
  └── Audit logs ──> optional User actor
```

## Tables

| Table | Purpose | Important relationships/invariants |
|---|---|---|
| `tenants` | Organization or personal identity boundary | Domain is tenant lookup metadata; status controls availability |
| `users` | Tenant-scoped identities and credential hashes | Email is unique within a tenant; optional role belongs to the same tenant |
| `roles` | Tenant-scoped role names | Unique by tenant and name |
| `role_permissions` | Permission keys assigned to roles | Composite identity of role and permission key |
| `sessions` | Active login sessions | Belongs to a user; may refer to a device; has an expiry |
| `devices` | Known user devices | Belongs to a user |
| `refresh_tokens` | Rotatable long-lived credentials | Only the token hash is stored; token can be revoked and expires |
| `user_mfa_methods` | TOTP/SMS MFA configuration | Secret or phone data is encrypted before persistence |
| `password_resets` | Single-use password reset records | Only the token hash is stored; status and expiry gate use |
| `invitations` | Tenant membership offers | Bound to a tenant, role, email, status, and expiry; token is hashed |
| `audit_logs` | Security-relevant events | Bound to a tenant and optionally an actor; includes trace/client context where available |

## Enumerated values

The Go constants in `internal/domain/models` are the application-facing source for enum values:

| Type | Values |
|---|---|
| Tenant type | `1` organization, `2` personal |
| Tenant status | `1` active, `2` suspended, `3` deleted |
| User status | `1` active, `2` pending, `3` disabled |
| MFA method | `1` TOTP, `2` SMS |
| Password reset | `1` pending, `2` used, `3` expired |
| Invitation | `1` pending, `2` accepted, `3` expired |

## Integrity rules

- Tenant-scoped email uniqueness is enforced by `(tenant_id, email)`.
- Tenant and user identifiers are UUIDs; role identifiers are integers.
- Foreign keys generally cascade when the owning tenant/user is deleted, while optional actor/role references may be set to null depending on the migration lineage.
- Token hashes must be unique so one opaque token resolves to at most one record.
- Repository methods must scope user, role, invitation, and audit access to the correct tenant even when the database foreign key alone cannot express the authorization rule.
- Timestamps are generated in UTC by application/database conventions and expiry comparisons must use current time safely.

## Migration baseline

The former overlapping `0001` and `001`–`007` histories were replaced by the single authoritative `000001_init_schema.up.sql` baseline. It creates all 11 application tables in dependency order with the fields, foreign keys, enum checks, uniqueness rules, timestamps, and indexes consumed by the application.

Existing databases from either removed lineage are not upgraded in place. Export required data, validate it against the authoritative model, and import it into a freshly migrated database. See the [migration operations guide](migrations.md).

## Reference artifacts

- [OpenAPI contract](reference/OpenAPI3.yaml)
- [Entity relationship diagram (SVG)](reference/data-model/erd.svg)
- [Entity relationship diagram (PNG)](reference/data-model/erd.png)
- [Entity relationship diagram source](reference/data-model/erd.mmd)

The ERD originated as a design artifact and may be ahead of the active migration lineage. Validate it against migrations before using it for implementation decisions.
