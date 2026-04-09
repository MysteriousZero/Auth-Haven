# Caching Specification

This document defines the caching strategy for Auth Haven to improve performance and reduce database load using Redis.

## 1. Technology Stack
- **Provider**: Redis (v7.0+)
- **Storage Strategy**: 
    - **Opaque tokens/IDs**: Stored as simple key-value pairs.
    - **Complex objects**: Serialized as JSON strings.

## 2. Cache Targets

### 2.1 Session Data
To avoid hitting the PostgreSQL database on every authenticated request, active sessions are cached.
- **Key**: `auth:sess:{session_id}`
- **TTL**: 1 hour (sliding window - reset on each access).
- **Scope**: Includes `UserID`, `TenantID`, and `RoleID`.
- **Invalidation**: 
    - Explicit `Logout` or `RevokeSession`.
    - User password change (invalidate all user sessions).

### 2.2 Tenant Metadata
Tenant configuration and domain mapping are frequently accessed but rarely changed.
- **Key**: `auth:ten:{tenant_id}`
- **Key (Domain mapping)**: `auth:dom:{domain_name}`
- **TTL**: 24 hours.
- **Invalidation**: 
    - `UpdateTenant` operation.

### 2.3 User Basic Info
- **Key**: `auth:user:{user_id}:basic`
- **TTL**: 10 minutes.
- **Invalidation**: 
    - `UpdateUserStatus` or `UpdateUserRole`.

## 3. Caching Patterns
- **Cache-Aside**: The application first checks Redis. On miss, it fetches from SQL and populates Redis.
- **Fail-Safe**: If Redis is unavailable, the system MUST fail-over to direct database access (except for Rate Limiting, which is Fail-Closed).

## 4. Key Naming Convention
All keys must be prefixed with `auth:` to avoid collisions if the Redis instance is shared.
- Format: `auth:{component}:{sub-component}:{identifier}`
