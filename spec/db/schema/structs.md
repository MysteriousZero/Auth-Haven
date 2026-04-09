# Go Struct Definitions

This document describes all Go structs that map directly to database tables.
Nullable DB columns are expressed as pointer types.
Smallint DB columns map to int16.

---

## 1. Core Entities

```mermaid
classDiagram

    class Tenant {
        +string TenantID
        +string Name
        +*string Domain
        +int16 Type
        +int16 Status
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class User {
        +string UserID
        +string TenantID
        +*int64 RoleID
        +string Email
        +string PasswordHash
        +string FullName
        +int16 Status
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +*time.Time LastLoginAt
    }

    class Role {
        +int64 RoleID
        +string TenantID
        +string Name
        +time.Time CreatedAt
    }

    class RolePermission {
        +int64 RoleID
        +string PermissionKey
    }
```

---

## 2. Authentication & Session

```mermaid
classDiagram

    class RefreshToken {
        +string TokenID
        +string UserID
        +string TokenHash
        +string UserAgent
        +string IPAddress
        +bool Revoked
        +time.Time ExpiresAt
        +time.Time CreatedAt
    }

    class Session {
        +string SessionID
        +string UserID
        +string DeviceID
        +string IPAddress
        +string UserAgent
        +time.Time CreatedAt
        +time.Time ExpiresAt
    }

    class Device {
        +string DeviceID
        +string UserID
        +string DeviceName
        +time.Time LastSeenAt
        +time.Time CreatedAt
    }
```

---

## 3. MFA & Security

```mermaid
classDiagram

    class UserMFAMethod {
        +string MFAID
        +string UserID
        +int16 Type
        +*string Secret "encrypted at rest"
        +*string PhoneNumber
        +bool Enabled
        +time.Time CreatedAt
    }

    class PasswordReset {
        +string ResetID
        +string UserID
        +string TokenHash
        +int16 Status
        +time.Time ExpiresAt
        +time.Time CreatedAt
    }

    class Invitation {
        +string InvitationID
        +string TenantID
        +int64 RoleID
        +string Email
        +string TokenHash
        +int16 Status
        +time.Time ExpiresAt
        +time.Time CreatedAt
    }
```

---

## 4. Audit Logging

```mermaid
classDiagram

    class AuditLog {
        +string LogID
        +string UserID
        +string TenantID
        +string Action
        +*string TargetID
        +*string Metadata
        +string IPAddress
        +string UserAgent
        +time.Time CreatedAt
    }
```

---

## Type Reference

| DB Type   | Go Type      | Notes                        |
|-----------|--------------|------------------------------|
| string    | string       |                              |
| int64     | int64        |                              |
| smallint  | int16        |                              |
| bool      | bool         |                              |
| datetime  | time.Time    |                              |
| nullable  | pointer (*)  | e.g. *string, *int64, *time.Time |