# Domain Services

Services contain all business logic. They are transport-agnostic — no HTTP or gRPC concepts here.
Services call repositories and emit audit logs. All mutating operations must record an audit log.
Personal tenant rules are enforced at this layer, not the repository layer.

---

## Rules

- Personal tenant users (`tenant.type == 2`) cannot be invited, assigned roles, or listed alongside org users
- Personal tenant data is always scoped by `user_id`, never by `tenant_id`
- Org tenant users are resolved by email domain match against `tenants.domain`
- Public email domains (gmail.com, yahoo.com, hotmail.com, outlook.com, etc.) are blocked for org registration
- The first user to register under an org tenant is automatically assigned the owner role
- Access tokens are short-lived JWTs; refresh tokens are long-lived and stored hashed
- All password comparison must use constant-time comparison to prevent timing attacks
- MFA challenge state is short-lived and must not be stored in the DB — use a signed temporary token

---

## 1. Auth Service

Handles login, logout, token lifecycle, and MFA verification.

### Methods

```
Login(tenantID, email, password string) (LoginResult, error)
VerifyMFA(tempToken, code string) (LoginResult, error)
Logout(sessionID string) error
LogoutAll(userID string) error
RefreshToken(refreshTokenRaw string) (TokenPair, error)
```

### LoginResult
```
type LoginResult struct {
    MFARequired bool
    TempToken   string    // populated only if MFARequired == true
    AccessToken  string   // populated only if MFARequired == false
    RefreshToken string   // populated only if MFARequired == false
    SessionID    string   // populated only if MFARequired == false
}
```

### TokenPair
```
type TokenPair struct {
    AccessToken  string
    RefreshToken string
}
```

### Claims
```
type Claims struct {
    UserID     string
    TenantID   string
    TenantType int16
    RoleID     *int64
    Email      string
}
```

---

### Login

```mermaid
flowchart TD
    A[Login] --> B[Resolve tenant by tenantID]
    B --> B1{Tenant exists and active?}
    B1 -- No --> E1[ErrTenantNotFound / ErrTenantSuspended]
    B1 -- Yes --> C[GetUserByEmail]
    C --> C1{User exists?}
    C1 -- No --> E2[ErrInvalidCredentials]
    C1 -- Yes --> C2{User active?}
    C2 -- No --> E3[ErrUserDisabled]
    C2 -- Yes --> D[Compare password hash]
    D --> D1{Password correct?}
    D1 -- No --> E2[ErrInvalidCredentials]
    D1 -- Yes --> E[ListMFAMethods for user]
    E --> E1{Any MFA method enabled?}
    E1 -- Yes --> F[Issue signed temp token with userID + expiry]
    F --> G[Return LoginResult with MFARequired=true, TempToken]
    E1 -- No --> H[UpdateLastLogin]
    H --> I[CreateSession]
    I --> J[Issue access token + refresh token]
    J --> K[CreateRefreshToken in DB]
    K --> L[Record audit log: user.login]
    L --> M[Return LoginResult with tokens]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrTenantNotFound | tenant does not exist |
| ErrTenantSuspended | tenant status is suspended or deleted |
| ErrInvalidCredentials | user not found or password wrong — same error to prevent enumeration |
| ErrUserDisabled | user status is disabled or pending |

---

### VerifyMFA

```mermaid
flowchart TD
    A[VerifyMFA] --> B[Validate and decode temp token]
    B --> B1{Token valid and not expired?}
    B1 -- No --> E1[ErrInvalidToken]
    B1 -- Yes --> C[ListMFAMethods for userID from token]
    C --> D[Find enabled MFA method matching submitted code]
    D --> D1{Code valid?}
    D1 -- No --> E2[ErrInvalidMFACode]
    D1 -- Yes --> E[UpdateLastLogin]
    E --> F[CreateSession]
    F --> G[Issue access token + refresh token]
    G --> H[CreateRefreshToken in DB]
    H --> I[Record audit log: user.mfa_verified]
    I --> J[Return LoginResult with tokens]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidToken | temp token missing, malformed, or expired |
| ErrInvalidMFACode | code does not match any enabled MFA method |

---

### Logout

```mermaid
flowchart TD
    A[Logout] --> B[GetSessionByID]
    B --> B1{Session exists?}
    B1 -- No --> E1[ErrSessionNotFound]
    B1 -- Yes --> C[DeleteSession]
    C --> D[RevokeRefreshToken for session user]
    D --> E[Record audit log: user.logout]
    E --> F[Return success]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrSessionNotFound | session does not exist or already deleted |

---

### LogoutAll

```mermaid
flowchart TD
    A[LogoutAll] --> B[GetUserByID]
    B --> B1{User exists?}
    B1 -- No --> E1[ErrUserNotFound]
    B1 -- Yes --> C[DeleteAllUserSessions]
    C --> D[RevokeAllUserTokens]
    D --> E[Record audit log: user.logout_all]
    E --> F[Return success]
```

---

### RefreshToken

```mermaid
flowchart TD
    A[RefreshToken] --> B[Hash incoming raw token]
    B --> C[GetRefreshTokenByHash]
    C --> C1{Token exists?}
    C1 -- No --> E1[ErrInvalidToken]
    C1 -- Yes --> C2{Token revoked?}
    C2 -- Yes --> E1[ErrInvalidToken]
    C2 -- No --> C3{Token expired?}
    C3 -- Yes --> E2[ErrTokenExpired]
    C3 -- No --> D[GetUserByID]
    D --> D1{User still active?}
    D1 -- No --> E3[ErrUserDisabled]
    D1 -- Yes --> E[RevokeRefreshToken - rotate]
    E --> F[Issue new access token + refresh token]
    F --> G[CreateRefreshToken in DB]
    G --> H[Return new TokenPair]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidToken | token not found or revoked |
| ErrTokenExpired | token past expires_at |
| ErrUserDisabled | user no longer active |

---

## 2. Token Service

Handles JWT token generation, validation, and cryptographic operations.

### Methods

```
GenerateTokenPair(userID, tenantID string, roleID *int64) (TokenPair, error)
GenerateTempToken(userID string) (string, error)
ValidateTempToken(tempToken string) (string, error) // returns userID
GenerateResetToken() string
GenerateInvitationToken() string
ValidateAccessToken(token string) (Claims, error)
```

### ValidateAccessToken

Validates a JWT access token and returns its claims. No DB call unless token revocation list is implemented.

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidToken | token malformed or signature invalid |
| ErrTokenExpired | token past expiry |

---

## 3. Registration Service

Handles self-registration and invitation-based signup.

### Methods

```
RegisterIndividual(email, password, fullName string) (User, error)
RegisterOrgUser(email, password, fullName string) (User, error)
RegisterWithInvitation(invitationToken, password, fullName string) (User, error)
```

---

### RegisterIndividual

```mermaid
flowchart TD
    A[RegisterIndividual] --> B[Validate email format]
    B --> B1{Valid?}
    B1 -- No --> E1[ErrInvalidEmail]
    B1 -- Yes --> C[GetUserByEmail on personal tenantID]
    C --> C1{Email already registered?}
    C1 -- Yes --> E2[ErrEmailTaken]
    C1 -- No --> D[Hash password]
    D --> E[CreateUser with tenant_id=personal, role_id=nil]
    E --> F[Record audit log: user.registered]
    F --> G[Return User]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidEmail | email format invalid |
| ErrEmailTaken | email already registered under personal tenant |

---

### RegisterOrgUser

```mermaid
flowchart TD
    A[RegisterOrgUser] --> B[Validate email format]
    B --> B1{Valid?}
    B1 -- No --> E1[ErrInvalidEmail]
    B1 -- Yes --> C[Extract domain from email]
    C --> D{Domain on public blocklist?}
    D -- Yes --> E2[ErrPublicDomainNotAllowed]
    D -- No --> F[GetTenantByDomain]
    F --> F1{Tenant exists?}
    F1 -- Yes --> G{Tenant active?}
    G -- No --> E3[ErrTenantSuspended]
    G -- Yes --> H[GetUserByEmail on tenantID]
    H --> H1{Email already registered?}
    H1 -- Yes --> E4[ErrEmailTaken]
    H1 -- No --> I[Hash password]
    I --> J{First user in tenant?}
    J -- Yes --> K[Assign owner role]
    J -- No --> L[Assign default member role]
    K & L --> M[CreateUser]
    M --> N[Record audit log: user.registered]
    N --> O[Return User]
    F1 -- No --> P[Auto-create org tenant from domain]
    P --> K
```

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidEmail | email format invalid |
| ErrPublicDomainNotAllowed | domain is on public provider blocklist |
| ErrTenantSuspended | tenant exists but is suspended or deleted |
| ErrEmailTaken | email already registered under this tenant |

---

### RegisterWithInvitation

```mermaid
flowchart TD
    A[RegisterWithInvitation] --> B[Hash token]
    B --> C[GetInvitationByHash]
    C --> C1{Invitation exists?}
    C1 -- No --> E1[ErrInvalidToken]
    C1 -- Yes --> C2{Status pending?}
    C2 -- No --> E2[ErrInvitationExpiredOrUsed]
    C2 -- Yes --> C3{Past expires_at?}
    C3 -- Yes --> E2
    C3 -- No --> D[GetUserByEmail on invitation tenantID]
    D --> D1{Email already registered?}
    D1 -- Yes --> E3[ErrEmailTaken]
    D1 -- No --> E[Hash password]
    E --> F[CreateUser with tenantID and roleID from invitation]
    F --> G[UpdateInvitationStatus to accepted]
    G --> H[Record audit log: user.registered_via_invitation]
    H --> I[Return User]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidToken | invitation token not found |
| ErrInvitationExpiredOrUsed | invitation status is not pending or past expires_at |
| ErrEmailTaken | email already registered under this tenant |

---

## 3. Password Service

Handles password reset and change.

### Methods

```
RequestPasswordReset(tenantID, email string) error
ResetPassword(token, newPassword string) error
ChangePassword(userID, currentPassword, newPassword string) error
```

---

### RequestPasswordReset

```mermaid
flowchart TD
    A[RequestPasswordReset] --> B[GetUserByEmail]
    B --> B1{User exists and active?}
    B1 -- No --> C[Return success silently]
    B1 -- Yes --> D[Generate reset token]
    D --> E[CreatePasswordReset with 1hr expiry]
    E --> F[Send reset email - via email provider]
    F --> G[Record audit log: user.password_reset_requested]
    G --> H[Return success]
```

> Always returns success even if email not found — prevents user enumeration.

---

### ResetPassword

```mermaid
flowchart TD
    A[ResetPassword] --> B[Hash token]
    B --> C[GetPasswordResetByHash]
    C --> C1{Record exists?}
    C1 -- No --> E1[ErrInvalidToken]
    C1 -- Yes --> C2{Status pending?}
    C2 -- No --> E2[ErrTokenExpired]
    C2 -- Yes --> C3{Past expires_at?}
    C3 -- Yes --> E2
    C3 -- No --> D[Validate new password strength]
    D --> D1{Valid?}
    D1 -- No --> E3[ErrWeakPassword]
    D1 -- Yes --> E[Hash new password]
    E --> F[Update user password_hash]
    F --> G[UpdatePasswordResetStatus to used]
    G --> H[RevokeAllUserTokens]
    H --> I[DeleteAllUserSessions]
    I --> J[Record audit log: user.password_reset]
    J --> K[Return success]
```

**Error cases**

| Error | Condition |
|---|---|
| ErrInvalidToken | token not found |
| ErrTokenExpired | status not pending or past expires_at |
| ErrWeakPassword | password fails strength rules |

---

### ChangePassword

```mermaid
flowchart TD
    A[ChangePassword] --> B[GetUserByID]
    B --> C[Compare currentPassword with hash]
    C --> C1{Correct?}
    C1 -- No --> E1[ErrInvalidCredentials]
    C1 -- Yes --> D[Validate new password strength]
    D --> D1{Valid?}
    D1 -- No --> E2[ErrWeakPassword]
    D1 -- Yes --> E[Hash new password]
    E --> F[Update user password_hash]
    F --> G[RevokeAllUserTokens]
    G --> H[DeleteAllUserSessions]
    H --> I[Record audit log: user.password_changed]
    I --> J[Return success]
```

---

## 4. MFA Service

Handles MFA enrollment and management.

### Methods

```
EnrollTOTP(userID string) (TOTPEnrollment, error)
ActivateTOTP(userID, mfaID, code string) error
EnrollSMS(userID, phoneNumber string) error
ActivateSMS(userID, mfaID, code string) error
DisableMFAMethod(userID, mfaID string) error
ListMFAMethods(userID string) ([]UserMFAMethod, error)
```

### TOTPEnrollment
```
type TOTPEnrollment struct {
    MFAID     string
    Secret    string
    QRCodeURL string
}
```

---

### EnrollTOTP

```mermaid
flowchart TD
    A[EnrollTOTP] --> B[Generate TOTP secret]
    B --> C[CreateMFAMethod with enabled=false]
    C --> D[Generate QR code URL from secret]
    D --> E[Return TOTPEnrollment]
```

> Method is created with `enabled=false` until user verifies with ActivateTOTP.

---

### ActivateTOTP

```mermaid
flowchart TD
    A[ActivateTOTP] --> B[ListMFAMethods for userID]
    B --> C{mfaID found and type=TOTP?}
    C -- No --> E1[ErrMFAMethodNotFound]
    C -- Yes --> D{Already enabled?}
    D -- Yes --> E2[ErrMFAAlreadyEnabled]
    D -- No --> E[Verify TOTP code against secret]
    E --> E1{Code valid?}
    E1 -- No --> E3[ErrInvalidMFACode]
    E1 -- Yes --> F[UpdateMFAMethod enabled=true]
    F --> G[Record audit log: user.mfa_enrolled]
    G --> H[Return success]
```

---

### DisableMFAMethod

```mermaid
flowchart TD
    A[DisableMFAMethod] --> B[ListMFAMethods for userID]
    B --> C{mfaID belongs to user?}
    C -- No --> E1[ErrMFAMethodNotFound]
    C -- Yes --> D[UpdateMFAMethod enabled=false]
    D --> E[Record audit log: user.mfa_disabled]
    E --> F[Return success]
```

---

## 5. User Service

Handles profile and user management within a tenant.

### Methods

```
GetProfile(userID string) (User, error)
UpdateProfile(userID, fullName string) error
ListUsers(tenantID string, limit, offset int) ([]User, error)
GetUser(userID string) (User, error)
UpdateUserStatus(actorID, targetUserID string, status int16) error
AssignRole(actorID, targetUserID string, roleID int64) error
```

> ListUsers must never be called with the personal tenantID — enforce at service layer.

---

### UpdateUserStatus

```mermaid
flowchart TD
    A[UpdateUserStatus] --> B[GetUserByID actor]
    B --> C{Actor is owner or admin?}
    C -- No --> E1[ErrForbidden]
    C -- Yes --> D[GetUserByID target]
    D --> D1{Target in same tenant?}
    D1 -- No --> E2[ErrForbidden]
    D1 -- Yes --> E[UpdateUserStatus in repo]
    E --> F[Record audit log: user.status_changed]
    F --> G[Return success]
```

---

### AssignRole

```mermaid
flowchart TD
    A[AssignRole] --> B[GetUserByID actor]
    B --> C{Actor is owner or admin?}
    C -- No --> E1[ErrForbidden]
    C -- Yes --> D[GetRoleByID]
    D --> D1{Role belongs to same tenant?}
    D1 -- No --> E2[ErrRoleNotFound]
    D1 -- Yes --> E[GetUserByID target]
    E --> E1{Target in same tenant?}
    E1 -- No --> E3[ErrForbidden]
    E1 -- Yes --> F[UpdateUserRole in repo]
    F --> G[Record audit log: user.role_assigned]
    G --> H[Return success]
```

---

## 6. Invitation Service

Handles sending, listing, and revoking invitations. Org tenants only.

### Methods

```
SendInvitation(actorID, tenantID, email string, roleID int64) error
ListInvitations(tenantID string, limit, offset int) ([]Invitation, error)
RevokeInvitation(actorID, invitationID string) error
ResendInvitation(actorID, invitationID string) error
```

---

### SendInvitation

```mermaid
flowchart TD
    A[SendInvitation] --> B[GetTenantByID]
    B --> B1{Tenant is org type?}
    B1 -- No --> E1[ErrForbidden]
    B1 -- Yes --> C[GetUserByID actor]
    C --> C1{Actor is owner or admin?}
    C1 -- No --> E2[ErrForbidden]
    C1 -- Yes --> D[GetUserByEmail on tenantID]
    D --> D1{Already a member?}
    D1 -- Yes --> E3[ErrUserAlreadyMember]
    D1 -- No --> E[GetRoleByID]
    E --> E1{Role belongs to tenant?}
    E1 -- No --> E4[ErrRoleNotFound]
    E1 -- Yes --> F[Generate invitation token]
    F --> G[CreateInvitation with 72hr expiry]
    G --> H[Send invitation email]
    H --> I[Record audit log: invitation.sent]
    I --> J[Return success]
```

---

### RevokeInvitation

```mermaid
flowchart TD
    A[RevokeInvitation] --> B[GetInvitationByID]
    B --> B1{Invitation exists?}
    B1 -- No --> E1[ErrInvitationNotFound]
    B1 -- Yes --> C[GetUserByID actor]
    C --> C1{Actor is owner or admin in same tenant?}
    C1 -- No --> E2[ErrForbidden]
    C1 -- Yes --> D{Invitation status pending?}
    D -- No --> E3[ErrInvitationExpiredOrUsed]
    D -- Yes --> E[UpdateInvitationStatus to expired]
    E --> F[Record audit log: invitation.revoked]
    F --> G[Return success]
```

---

## 7. Role Service

Handles role and permission management. Org tenants only.

### Methods

```
CreateRole(actorID, tenantID, name string) (Role, error)
DeleteRole(actorID string, roleID int64) error
ListRoles(tenantID string, limit, offset int) ([]Role, error)
GetRolePermissions(roleID int64) ([]RolePermission, error)
SetRolePermissions(actorID string, roleID int64, keys []string) error
```

---

### CreateRole

```mermaid
flowchart TD
    A[CreateRole] --> B[GetUserByID actor]
    B --> C{Actor is owner?}
    C -- No --> E1[ErrForbidden]
    C -- Yes --> D[GetTenantByID]
    D --> D1{Tenant is org type?}
    D1 -- No --> E2[ErrForbidden]
    D1 -- Yes --> E[CreateRole in repo]
    E --> F[Record audit log: role.created]
    F --> G[Return Role]
```

---

### DeleteRole

```mermaid
flowchart TD
    A[DeleteRole] --> B[GetUserByID actor]
    B --> C{Actor is owner?}
    C -- No --> E1[ErrForbidden]
    C -- Yes --> D[GetRoleByID]
    D --> D1{Role belongs to actor tenant?}
    D1 -- No --> E2[ErrRoleNotFound]
    D1 -- Yes --> D2{Any users assigned to role?}
    D2 -- Yes --> E3[ErrRoleInUse]
    D2 -- No --> E[DeleteRole in repo]
    E --> F[Record audit log: role.deleted]
    F --> G[Return success]
```

---

## 8. Session & Device Service

### Methods

```
ListSessions(userID string, limit, offset int) ([]Session, error)
RevokeSession(actorID, sessionID string) error
ListDevices(userID string, limit, offset int) ([]Device, error)
RemoveDevice(actorID, deviceID string) error
```

---

### RevokeSession

```mermaid
flowchart TD
    A[RevokeSession] --> B[GetSessionByID]
    B --> B1{Session exists?}
    B1 -- No --> E1[ErrSessionNotFound]
    B1 -- Yes --> C{Session belongs to actorID?}
    C -- No --> E2[ErrForbidden]
    C -- Yes --> D[DeleteSession]
    D --> E[Record audit log: session.revoked]
    E --> F[Return success]
```

---

### RemoveDevice

```mermaid
flowchart TD
    A[RemoveDevice] --> B[GetDeviceByID]
    B --> B1{Device exists?}
    B1 -- No --> E1[ErrDeviceNotFound]
    B1 -- Yes --> C{Device belongs to actorID?}
    C -- No --> E2[ErrForbidden]
    C -- Yes --> D[DeleteDevice]
    D --> E[Record audit log: device.removed]
    E --> F[Return success]
```

---

## 9. Tenant Service

### Methods

```
CreateTenant(name, domain string) (Tenant, error)
GetTenant(tenantID string) (Tenant, error)
UpdateTenantStatus(actorID, tenantID string, status int16) error
```

> CreateTenant is a superadmin or system-level operation, not user-facing.

---

## 10. Audit Service

### Methods

```
ListUserLogs(actorID, userID, cursor string, limit int) ([]AuditLog, string, error)
ListTenantLogs(actorID, tenantID, cursor string, limit int) ([]AuditLog, string, error)
```

Both methods verify the actor has permission to view the requested logs before calling the repository.

---

## Error Catalog

| Error | Description |
|---|---|
| ErrInvalidCredentials | wrong email or password |
| ErrUserNotFound | user does not exist |
| ErrUserDisabled | user is disabled or pending |
| ErrUserAlreadyMember | user already belongs to tenant |
| ErrEmailTaken | email already registered in this tenant |
| ErrInvalidEmail | email format invalid |
| ErrPublicDomainNotAllowed | email domain is a public provider |
| ErrWeakPassword | password fails strength requirements |
| ErrTenantNotFound | tenant does not exist |
| ErrTenantSuspended | tenant is suspended or deleted |
| ErrInvalidToken | token missing, malformed, revoked, or not found |
| ErrTokenExpired | token past expiry |
| ErrSessionNotFound | session does not exist |
| ErrDeviceNotFound | device does not exist |
| ErrRoleNotFound | role does not exist or belongs to different tenant |
| ErrRoleInUse | role has users assigned, cannot delete |
| ErrMFAMethodNotFound | MFA method not found for user |
| ErrMFAAlreadyEnabled | MFA method already active |
| ErrInvalidMFACode | submitted MFA code is incorrect |
| ErrInvitationNotFound | invitation does not exist |
| ErrInvitationExpiredOrUsed | invitation is not in pending status |
| ErrForbidden | actor lacks permission for this operation |