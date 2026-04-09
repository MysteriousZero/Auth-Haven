# Database Constraints

This document lists all primary keys, foreign keys, unique constraints, and indexes.

---

## Primary Keys

- tenants: tenant_id
- users: user_id
- roles: role_id
- role_permissions: (role_id, permission_key)
- refresh_tokens: token_id
- sessions: session_id
- devices: device_id
- user_mfa_methods: mfa_id
- password_resets: reset_id
- invitations: invitation_id
- audit_logs: log_id

---

## Foreign Keys

- users.tenant_id → tenants.tenant_id
- users.role_id → roles.role_id (nullable)
- roles.tenant_id → tenants.tenant_id
- refresh_tokens.user_id → users.user_id
- sessions.user_id → users.user_id
- sessions.device_id → devices.device_id
- devices.user_id → users.user_id
- user_mfa_methods.user_id → users.user_id
- password_resets.user_id → users.user_id
- invitations.tenant_id → tenants.tenant_id
- invitations.role_id → roles.role_id
- audit_logs.user_id → users.user_id
- audit_logs.tenant_id → tenants.tenant_id

---

## Unique Constraints

- tenants.domain WHERE domain IS NOT NULL (partial unique; personal tenant has NULL domain)
- users.(tenant_id, email)
- refresh_tokens.token_hash
- password_resets.token_hash
- invitations.token_hash

---

## Nullable Columns

- tenants.domain — NULL for the system-owned personal tenant only
- users.role_id — NULL for personal tenant users
- users.last_login_at — NULL until first login
- user_mfa_methods.secret — NULL for SMS and Email MFA types
- user_mfa_methods.phone_number — NULL for TOTP and Email MFA types

---

## Indexes

- tenants.name
- tenants.type
- users.role_id
- users.status
- refresh_tokens.user_id
- refresh_tokens.expires_at
- refresh_tokens.revoked
- sessions.user_id
- sessions.device_id
- sessions.(user_id, expires_at)
- devices.user_id
- devices.last_seen_at
- user_mfa_methods.user_id
- user_mfa_methods.(user_id, type)
- user_mfa_methods.enabled
- password_resets.user_id
- password_resets.status
- password_resets.expires_at
- invitations.tenant_id
- invitations.role_id
- invitations.email
- invitations.status
- audit_logs.user_id
- audit_logs.tenant_id
- audit_logs.action
- audit_logs.target_id
- audit_logs.(tenant_id, created_at)
- audit_logs.(user_id, created_at)

---

## Seeded Data

- tenants: one system-owned row must be seeded at migration time
  - tenant_id: "personal"
  - name: "Personal"
  - type: 2 (personal)
  - status: 1 (active)
  - domain: NULL