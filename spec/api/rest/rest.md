# REST Handler Sequence Diagrams

These diagrams show the full request lifecycle for each REST endpoint.
Handlers are responsible for: parsing the request, calling the validator, calling the service, and writing the response.
No business logic lives in handlers — they delegate entirely to the service layer.

---

## Auth

### POST /v1/auth/login

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/login
    Handler->>Validator: Validate(email, password, tenant_id)
    Validator-->>Handler: OK | ErrValidation

    Handler->>AuthService: Login(tenantID, email, password)

    AuthService->>TenantRepo: GetTenantByID
    TenantRepo-->>AuthService: Tenant | ErrTenantNotFound

    AuthService->>UserRepo: GetUserByEmail(tenantID, email)
    UserRepo-->>AuthService: User | ErrInvalidCredentials

    AuthService->>Hasher: CompareHash(password, hash)
    Hasher-->>AuthService: match | no match

    AuthService->>UserRepo: ListMFAMethods(userID)
    UserRepo-->>AuthService: []UserMFAMethod

    alt MFA required
        AuthService-->>Handler: LoginResult{MFARequired: true, TempToken}
        Handler-->>Client: 200 OK {mfa_required: true, temp_token}
    else No MFA
        AuthService->>UserRepo: UpdateLastLogin(userID)
        AuthService->>AuthRepo: CreateSession(Session)
        AuthService->>TokenService: GenerateTokenPair(userID, tenantID, roleID)
        TokenService-->>AuthService: AccessToken, RefreshToken
        AuthService->>AuthRepo: CreateRefreshToken(RefreshToken)
        AuthService->>AuditRepo: Record(user.login)
        AuthService-->>Handler: LoginResult{AccessToken, RefreshToken, SessionID}
        Handler-->>Client: 200 OK {access_token, refresh_token, session_id}
    end
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrTenantNotFound | 401 | ErrInvalidCredentials |
| ErrTenantSuspended | 403 | ErrTenantSuspended |
| ErrInvalidCredentials | 401 | ErrInvalidCredentials |
| ErrUserDisabled | 403 | ErrUserDisabled |

> Tenant not found maps to 401 not 404 — prevents tenant enumeration.

---

### POST /v1/auth/mfa/verify

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/mfa/verify
    Handler->>Validator: Validate(temp_token, code)
    Validator-->>Handler: OK | ErrValidation

    Handler->>AuthService: VerifyMFA(tempToken, code)

    AuthService->>TokenService: ValidateTempToken(tempToken)
    TokenService-->>AuthService: userID | ErrInvalidToken

    AuthService->>UserRepo: ListMFAMethods(userID)
    UserRepo-->>AuthService: []UserMFAMethod

    AuthService->>MFAVerifier: VerifyCode(method, code)
    MFAVerifier-->>AuthService: valid | ErrInvalidMFACode

    AuthService->>UserRepo: UpdateLastLogin(userID)
    AuthService->>AuthRepo: CreateSession(Session)
    AuthService->>TokenService: GenerateTokenPair(userID, tenantID, roleID)
    TokenService-->>AuthService: AccessToken, RefreshToken
    AuthService->>AuthRepo: CreateRefreshToken(RefreshToken)
    AuthService->>AuditRepo: Record(user.mfa_verified)
    AuthService-->>Handler: LoginResult{AccessToken, RefreshToken, SessionID}
    Handler-->>Client: 200 OK {access_token, refresh_token, session_id}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidToken | 401 | ErrInvalidToken |
| ErrInvalidMFACode | 401 | ErrInvalidMFACode |

---

### POST /v1/auth/logout

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/logout
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims | ErrInvalidToken

    Handler->>AuthService: Logout(sessionID)

    AuthService->>AuthRepo: GetSessionByID(sessionID)
    AuthRepo-->>AuthService: Session | ErrSessionNotFound

    AuthService->>AuthRepo: DeleteSession(sessionID)
    AuthService->>AuthRepo: RevokeRefreshToken(userID)
    AuthService->>AuditRepo: Record(user.logout)
    AuthService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidToken | 401 | ErrInvalidToken |
| ErrSessionNotFound | 404 | ErrSessionNotFound |

---

### POST /v1/auth/logout-all

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/logout-all
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims | ErrInvalidToken

    Handler->>AuthService: LogoutAll(userID)
    AuthService->>AuthRepo: DeleteAllUserSessions(userID)
    AuthService->>AuthRepo: RevokeAllUserTokens(userID)
    AuthService->>AuditRepo: Record(user.logout_all)
    AuthService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

---

### POST /v1/auth/token/refresh

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/token/refresh
    Handler->>Validator: Validate(refresh_token)
    Validator-->>Handler: OK | ErrValidation

    Handler->>AuthService: RefreshToken(rawToken)
    AuthService->>Hasher: Hash(rawToken)
    AuthService->>AuthRepo: GetRefreshTokenByHash(hash)
    AuthRepo-->>AuthService: RefreshToken | ErrInvalidToken

    AuthService->>UserRepo: GetUserByID(userID)
    UserRepo-->>AuthService: User | ErrUserDisabled

    AuthService->>AuthRepo: RevokeRefreshToken(tokenID)
    AuthService->>TokenService: GenerateTokenPair(userID, tenantID, roleID)
    TokenService-->>AuthService: AccessToken, RefreshToken
    AuthService->>AuthRepo: CreateRefreshToken(RefreshToken)
    AuthService-->>Handler: TokenPair
    Handler-->>Client: 200 OK {access_token, refresh_token}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidToken | 401 | ErrInvalidToken |
| ErrTokenExpired | 401 | ErrTokenExpired |
| ErrUserDisabled | 403 | ErrUserDisabled |

---

## Registration

### POST /v1/auth/register/individual

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/register/individual
    Handler->>Validator: Validate(email, password, full_name)
    Validator-->>Handler: OK | ErrValidation

    Handler->>RegistrationService: RegisterIndividual(email, password, fullName)
    RegistrationService->>UserRepo: GetUserByEmail("personal", email)
    UserRepo-->>RegistrationService: nil | ErrEmailTaken

    RegistrationService->>Hasher: HashPassword(password)
    Hasher-->>RegistrationService: hash

    RegistrationService->>UserRepo: CreateUser(User{tenantID: "personal"})
    RegistrationService->>AuditRepo: Record(user.registered)
    RegistrationService-->>Handler: User
    Handler-->>Client: 201 Created {user}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidEmail | 422 | ErrInvalidEmail |
| ErrEmailTaken | 409 | ErrEmailTaken |
| ErrWeakPassword | 422 | ErrWeakPassword |

---

### POST /v1/auth/register/org

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/register/org
    Handler->>Validator: Validate(email, password, full_name)
    Validator-->>Handler: OK | ErrValidation

    Handler->>RegistrationService: RegisterOrgUser(email, password, fullName)
    RegistrationService->>DomainChecker: IsPublicDomain(domain)
    DomainChecker-->>RegistrationService: blocked | OK

    RegistrationService->>TenantRepo: GetTenantByDomain(domain)
    TenantRepo-->>RegistrationService: Tenant | nil

    alt Tenant not found
        RegistrationService->>TenantRepo: CreateTenant(name, domain)
        TenantRepo-->>RegistrationService: Tenant
        RegistrationService->>TenantRepo: CreateRole("owner", tenantID)
    end

    RegistrationService->>UserRepo: GetUserByEmail(tenantID, email)
    UserRepo-->>RegistrationService: nil | ErrEmailTaken

    RegistrationService->>Hasher: HashPassword(password)
    RegistrationService->>UserRepo: CreateUser(User{tenantID, roleID})
    RegistrationService->>AuditRepo: Record(user.registered)
    RegistrationService-->>Handler: User
    Handler-->>Client: 201 Created {user}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidEmail | 422 | ErrInvalidEmail |
| ErrPublicDomainNotAllowed | 400 | ErrPublicDomainNotAllowed |
| ErrTenantSuspended | 403 | ErrTenantSuspended |
| ErrEmailTaken | 409 | ErrEmailTaken |
| ErrWeakPassword | 422 | ErrWeakPassword |

---

### POST /v1/auth/register/invitation

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/register/invitation
    Handler->>Validator: Validate(invitation_token, password, full_name)
    Validator-->>Handler: OK | ErrValidation

    Handler->>RegistrationService: RegisterWithInvitation(token, password, fullName)
    RegistrationService->>Hasher: Hash(token)
    RegistrationService->>TenantRepo: GetInvitationByHash(hash)
    TenantRepo-->>RegistrationService: Invitation | ErrInvalidToken

    RegistrationService->>UserRepo: GetUserByEmail(tenantID, email)
    UserRepo-->>RegistrationService: nil | ErrEmailTaken

    RegistrationService->>Hasher: HashPassword(password)
    RegistrationService->>UserRepo: CreateUser(User{tenantID, roleID from invitation})
    RegistrationService->>TenantRepo: UpdateInvitationStatus(accepted)
    RegistrationService->>AuditRepo: Record(user.registered_via_invitation)
    RegistrationService-->>Handler: User
    Handler-->>Client: 201 Created {user}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidToken | 400 | ErrInvalidToken |
| ErrInvitationExpiredOrUsed | 400 | ErrInvitationExpiredOrUsed |
| ErrEmailTaken | 409 | ErrEmailTaken |
| ErrWeakPassword | 422 | ErrWeakPassword |

---

## Password

### POST /v1/auth/password/forgot

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/password/forgot
    Handler->>Validator: Validate(tenant_id, email)
    Validator-->>Handler: OK | ErrValidation

    Handler->>PasswordService: RequestPasswordReset(tenantID, email)
    PasswordService->>UserRepo: GetUserByEmail(tenantID, email)
    UserRepo-->>PasswordService: User | nil

    alt User exists and active
        PasswordService->>TokenService: GenerateResetToken()
        PasswordService->>UserRepo: CreatePasswordReset(PasswordReset)
        PasswordService->>EmailProvider: SendResetEmail(email, token)
        PasswordService->>AuditRepo: Record(user.password_reset_requested)
    end

    PasswordService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

> Always 204 regardless of whether email exists.

---

### POST /v1/auth/password/reset

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/password/reset
    Handler->>Validator: Validate(token, new_password)
    Validator-->>Handler: OK | ErrValidation

    Handler->>PasswordService: ResetPassword(token, newPassword)
    PasswordService->>Hasher: Hash(token)
    PasswordService->>UserRepo: GetPasswordResetByHash(hash)
    UserRepo-->>PasswordService: PasswordReset | ErrInvalidToken

    PasswordService->>Hasher: HashPassword(newPassword)
    PasswordService->>UserRepo: UpdatePasswordHash(userID, hash)
    PasswordService->>UserRepo: UpdatePasswordResetStatus(used)
    PasswordService->>AuthRepo: RevokeAllUserTokens(userID)
    PasswordService->>AuthRepo: DeleteAllUserSessions(userID)
    PasswordService->>AuditRepo: Record(user.password_reset)
    PasswordService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidToken | 400 | ErrInvalidToken |
| ErrTokenExpired | 400 | ErrTokenExpired |
| ErrWeakPassword | 422 | ErrWeakPassword |

---

### POST /v1/auth/password/change

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/auth/password/change
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims | ErrInvalidToken

    Handler->>Validator: Validate(current_password, new_password)
    Validator-->>Handler: OK | ErrValidation

    Handler->>PasswordService: ChangePassword(userID, currentPassword, newPassword)
    PasswordService->>UserRepo: GetUserByID(userID)
    PasswordService->>Hasher: CompareHash(currentPassword, hash)
    Hasher-->>PasswordService: match | ErrInvalidCredentials

    PasswordService->>Hasher: HashPassword(newPassword)
    PasswordService->>UserRepo: UpdatePasswordHash(userID, hash)
    PasswordService->>AuthRepo: RevokeAllUserTokens(userID)
    PasswordService->>AuthRepo: DeleteAllUserSessions(userID)
    PasswordService->>AuditRepo: Record(user.password_changed)
    PasswordService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrInvalidCredentials | 401 | ErrInvalidCredentials |
| ErrWeakPassword | 422 | ErrWeakPassword |

---

## MFA

### POST /v1/me/mfa/totp/enroll

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/me/mfa/totp/enroll
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims

    Handler->>MFAService: EnrollTOTP(userID)
    MFAService->>TOTPGenerator: GenerateSecret()
    TOTPGenerator-->>MFAService: secret
    MFAService->>UserRepo: CreateMFAMethod(enabled=false)
    MFAService->>QRGenerator: GenerateQRCodeURL(secret, email)
    MFAService-->>Handler: TOTPEnrollment{mfa_id, secret, qr_code_url}
    Handler-->>Client: 200 OK {mfa_id, secret, qr_code_url}
```

---

### POST /v1/me/mfa/totp/activate

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/me/mfa/totp/activate
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims

    Handler->>Validator: Validate(mfa_id, code)
    Validator-->>Handler: OK | ErrValidation

    Handler->>MFAService: ActivateTOTP(userID, mfaID, code)
    MFAService->>UserRepo: ListMFAMethods(userID)
    UserRepo-->>MFAService: []UserMFAMethod

    MFAService->>TOTPVerifier: VerifyCode(secret, code)
    TOTPVerifier-->>MFAService: valid | ErrInvalidMFACode

    MFAService->>UserRepo: UpdateMFAMethod(mfaID, enabled=true)
    MFAService->>AuditRepo: Record(user.mfa_enrolled)
    MFAService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrMFAMethodNotFound | 404 | ErrMFAMethodNotFound |
| ErrMFAAlreadyEnabled | 409 | ErrMFAAlreadyEnabled |
| ErrInvalidMFACode | 400 | ErrInvalidMFACode |

---

## Sessions

### GET /v1/me/sessions

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: GET /v1/me/sessions
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims

    Handler->>SessionService: ListSessions(userID)
    SessionService->>AuthRepo: GetActiveSessions(userID)
    AuthRepo-->>SessionService: []Session
    SessionService-->>Handler: []Session
    Handler-->>Client: 200 OK {sessions}
```

---

### DELETE /v1/me/sessions/{session_id}

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: DELETE /v1/me/sessions/{session_id}
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims

    Handler->>SessionService: RevokeSession(actorID, sessionID)
    SessionService->>AuthRepo: GetSessionByID(sessionID)
    AuthRepo-->>SessionService: Session | ErrSessionNotFound

    SessionService->>SessionService: session.UserID == actorID?
    SessionService->>AuthRepo: DeleteSession(sessionID)
    SessionService->>AuditRepo: Record(session.revoked)
    SessionService-->>Handler: nil
    Handler-->>Client: 204 No Content
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrSessionNotFound | 404 | ErrSessionNotFound |
| ErrForbidden | 403 | ErrForbidden |

---

## Invitations

### POST /v1/tenants/{tenant_id}/invitations

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: POST /v1/tenants/{tenant_id}/invitations
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims

    Handler->>Validator: Validate(email, role_id)
    Validator-->>Handler: OK | ErrValidation

    Handler->>InvitationService: SendInvitation(actorID, tenantID, email, roleID)
    InvitationService->>TenantRepo: GetTenantByID(tenantID)
    TenantRepo-->>InvitationService: Tenant | ErrForbidden (personal tenant)

    InvitationService->>UserRepo: GetUserByID(actorID)
    InvitationService->>TenantRepo: GetRoleByID(roleID)
    InvitationService->>UserRepo: GetUserByEmail(tenantID, email)
    UserRepo-->>InvitationService: nil | ErrUserAlreadyMember

    InvitationService->>TokenService: GenerateInvitationToken()
    InvitationService->>TenantRepo: CreateInvitation(Invitation)
    InvitationService->>EmailProvider: SendInvitationEmail(email, token)
    InvitationService->>AuditRepo: Record(invitation.sent)
    InvitationService-->>Handler: Invitation
    Handler-->>Client: 201 Created {invitation}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrForbidden | 403 | ErrForbidden |
| ErrRoleNotFound | 404 | ErrRoleNotFound |
| ErrUserAlreadyMember | 409 | ErrUserAlreadyMember |

---

## Audit Logs

### GET /v1/tenants/{tenant_id}/audit-logs

```mermaid
sequenceDiagram
    autonumber
    Client->>Handler: GET /v1/tenants/{tenant_id}/audit-logs
    Handler->>Middleware: ValidateAccessToken
    Middleware-->>Handler: Claims

    Handler->>AuditService: ListTenantLogs(actorID, tenantID, cursor, limit)
    AuditService->>UserRepo: GetUserByID(actorID)
    AuditService->>AuditService: actor has admin permission in tenant?
    AuditService->>AuditRepo: ListTenantLogs(tenantID, cursor, limit)
    AuditRepo-->>AuditService: []AuditLog, nextCursor
    AuditService-->>Handler: []AuditLog, nextCursor
    Handler-->>Client: 200 OK {logs, limit, cursor, next_cursor}
```

**Error responses**

| Service Error | HTTP Status | Response Code |
|---|---|---|
| ErrForbidden | 403 | ErrForbidden |