# Database Tables

This document defines all DB tables, columns, types, and indexes.

---

## 1. Tenants

| Column       | Type     | Notes                              |
|--------------|----------|------------------------------------|
| tenant_id PK | string   | UUID                               |
| name         | string   |                                    |
| domain       | string   | nullable; NULL for personal tenant |
| type         | smallint | 1=organization, 2=personal         |
| status       | smallint | 1=active, 2=suspended, 3=deleted   |
| created_at   | datetime |                                    |
| updated_at   | datetime |                                    |

**Indexes**
- UNIQUE(domain) WHERE domain IS NOT NULL
- INDEX(name)
- INDEX(type)

---

## 2. Users

| Column        | Type     | Notes                                          |
|---------------|----------|------------------------------------------------|
| user_id PK    | string   | UUID                                           |
| tenant_id FK  | string   | → tenants; always populated                    |
| role_id FK    | int64    | → roles; nullable; NULL for personal tenant users |
| email         | string   | unique per tenant                              |
| password_hash | string   |                                                |
| full_name     | string   |                                                |
| status        | smallint | 1=active, 2=pending, 3=disabled                |
| created_at    | datetime |                                                |
| updated_at    | datetime |                                                |
| last_login_at | datetime | nullable                                       |

**Indexes**
- UNIQUE(tenant_id, email)
- INDEX(role_id)
- INDEX(status)

---

## 3. Roles

| Column       | Type     | Notes      |
|--------------|----------|------------|
| role_id PK   | int64    |            |
| tenant_id FK | string   | → tenants  |
| name         | string   |            |
| created_at   | datetime |            |

**Indexes**
- UNIQUE(tenant_id, name)

---

## 4. Role Permissions

| Column         | Type   | Notes                  |
|----------------|--------|------------------------|
| role_id FK     | int64  | → roles                |
| permission_key | string | permission identifier  |

**Indexes**
- PRIMARY KEY(role_id, permission_key)

---

## 5. Refresh Tokens

| Column      | Type     | Notes    |
|-------------|----------|----------|
| token_id PK | string   | UUID     |
| user_id FK  | string   | → users  |
| token_hash  | string   |          |
| user_agent  | string   |          |
| ip_address  | string   |          |
| revoked     | bool     |          |
| expires_at  | datetime |          |
| created_at  | datetime |          |

**Indexes**
- UNIQUE(token_hash)
- INDEX(user_id)
- INDEX(expires_at)
- INDEX(revoked)

---

## 6. Sessions

| Column        | Type     | Notes     |
|---------------|----------|-----------|
| session_id PK | string   | UUID      |
| user_id FK    | string   | → users   |
| device_id FK  | string   | → devices |
| ip_address    | string   |           |
| user_agent    | string   |           |
| created_at    | datetime |           |
| expires_at    | datetime |           |

**Indexes**
- INDEX(user_id)
- INDEX(device_id)
- INDEX(user_id, expires_at)

---

## 7. Devices

| Column       | Type     | Notes   |
|--------------|----------|---------|
| device_id PK | string   | UUID    |
| user_id FK   | string   | → users |
| device_name  | string   |         |
| last_seen_at | datetime |         |
| created_at   | datetime |         |

**Indexes**
- INDEX(user_id)
- INDEX(last_seen_at)

---

## 8. User MFA Methods

| Column       | Type     | Notes                          |
|--------------|----------|--------------------------------|
| mfa_id PK    | string   | UUID                           |
| user_id FK   | string   | → users                        |
| type         | smallint | 1=TOTP, 2=SMS, 3=Email         |
| secret       | string   | nullable; TOTP only; encrypted at rest |
| phone_number | string   | nullable; SMS only             |
| enabled      | bool     |                                |
| created_at   | datetime |                                |

**Indexes**
- INDEX(user_id)
- INDEX(user_id, type)
- INDEX(enabled)

---

## 9. Password Resets

| Column      | Type     | Notes                        |
|-------------|----------|------------------------------|
| reset_id PK | string   | UUID                         |
| user_id FK  | string   | → users                      |
| token_hash  | string   |                              |
| status      | smallint | 1=pending, 2=used, 3=expired |
| expires_at  | datetime |                              |
| created_at  | datetime |                              |

**Indexes**
- UNIQUE(token_hash)
- INDEX(user_id)
- INDEX(status)
- INDEX(expires_at)

---

## 10. Invitations

| Column           | Type     | Notes                            |
|------------------|----------|----------------------------------|
| invitation_id PK | string   | UUID                             |
| tenant_id FK     | string   | → tenants; org tenants only      |
| role_id FK       | int64    | → roles                          |
| email            | string   |                                  |
| token_hash       | string   |                                  |
| status           | smallint | 1=pending, 2=accepted, 3=expired |
| expires_at       | datetime |                                  |
| created_at       | datetime |                                  |

**Indexes**
- UNIQUE(token_hash)
- INDEX(tenant_id)
- INDEX(role_id)
- INDEX(email)
- INDEX(status)

---

## 11. Audit Logs

| Column       | Type     | Notes    |
|--------------|----------|----------|
| log_id PK    | string   | UUID     |
| user_id FK   | string   | → users  |
| tenant_id FK | string   | → tenants|
| action       | string   |          |
| target_id    | string   | nullable |
| metadata     | string   | jsonb/string; nullable |
| ip_address   | string   |          |
| user_agent   | string   |          |
| created_at   | datetime |          |

**Indexes**
- INDEX(user_id)
- INDEX(tenant_id)
- INDEX(action)
- INDEX(target_id)
- INDEX(tenant_id, created_at)
- INDEX(user_id, created_at)