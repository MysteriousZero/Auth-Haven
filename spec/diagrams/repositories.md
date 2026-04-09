# Repository Definitions

Repositories are the only layer that touches the database.
They accept and return domain structs defined in structs.md.
No business logic lives here — only data access.

---

## 1. Tenant Repository

Manages tenant-scoped resources: tenants, roles, role permissions, and invitations.

```mermaid
classDiagram
    class TenantRepository {
        +CreateTenant(Tenant) error
        +GetTenantByID(id string) (Tenant, error)
        +GetTenantByDomain(domain string) (Tenant, error)
        +UpdateTenantStatus(id string, status int16) error

        +CreateRole(Role) error
        +ListRoles(tenantID string, limit int, offset int) ([]Role, error)
        +GetRoleByID(roleID int64) (Role, error)
        +DeleteRole(roleID int64) error

        +SetPermissions(roleID int64, keys []string) error
        +ListPermissions(roleID int64) ([]RolePermission, error)

        +CreateInvitation(Invitation) error
        +GetInvitationByHash(hash string) (Invitation, error)
        +ListInvitations(tenantID string, limit int, offset int) ([]Invitation, error)
        +UpdateInvitationStatus(id string, status int16) error
    }
```

---

## 2. User Repository

Manages identity-owned data: users, devices, MFA methods, password resets.

```mermaid
classDiagram
    class UserRepository {
        +CreateUser(User) error
        +GetUserByID(id string) (User, error)
        +GetUserByEmail(tenantID string, email string) (User, error)
        +ListUsers(tenantID string, limit int, offset int) ([]User, error)
        +UpdateUserStatus(id string, status int16) error
        +UpdateUserRole(id string, roleID int64) error
        +UpdateLastLogin(id string) error

        +RegisterDevice(Device) error
        +GetDeviceByID(deviceID string) (Device, error)
        +ListDevices(userID string, limit int, offset int) ([]Device, error)
        +UpdateDeviceLastSeen(deviceID string) error
        +DeleteDevice(deviceID string) error

        +CreateMFAMethod(UserMFAMethod) error
        +ListMFAMethods(userID string) ([]UserMFAMethod, error)
        +UpdateMFAMethod(mfaID string, enabled bool) error
        +DeleteMFAMethod(mfaID string) error

        +CreatePasswordReset(PasswordReset) error
        +GetPasswordResetByHash(hash string) (PasswordReset, error)
        +UpdatePasswordResetStatus(id string, status int16) error
    }
```

---

## 3. Auth Repository

Handles authentication lifecycle: refresh tokens and sessions.

```mermaid
classDiagram
    class AuthRepository {
        +CreateRefreshToken(RefreshToken) error
        +GetRefreshTokenByHash(hash string) (RefreshToken, error)
        +RevokeRefreshToken(id string) error
        +RevokeAllUserTokens(userID string) error

        +CreateSession(Session) error
        +GetSessionByID(sessionID string) (Session, error)
        +GetActiveSessions(userID string, limit int, offset int) ([]Session, error)
        +DeleteSession(sessionID string) error
        +DeleteAllUserSessions(userID string) error
    }
```

---

## 4. Audit Repository

Stores and queries audit event logs.

```mermaid
classDiagram
    class AuditRepository {
        +Record(AuditLog) error
        +ListUserLogs(userID string, cursor string, limit int) ([]AuditLog, string, error)
        +ListTenantLogs(tenantID string, cursor string, limit int) ([]AuditLog, string, error)
    }
```