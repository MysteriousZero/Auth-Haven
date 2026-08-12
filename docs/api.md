# API guide

This page lists routes and RPCs wired by the current server. The machine-readable [OpenAPI contract](reference/OpenAPI3.yaml) is retained as a reference artifact and may contain endpoints that are not currently wired.

## HTTP conventions

- Base path: `/v1`
- Protected routes require `Authorization: Bearer <access-token>`.
- Logout additionally requires `X-Session-ID`.
- `X-Request-ID` is accepted or generated and echoed in the response.
- Error bodies generally contain `error` and `message`; validation errors may include `details`.
- List endpoints use `limit` and `offset`, except audit logs, which use `limit` and `cursor`.

## Public HTTP routes

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Process health response |
| `POST` | `/v1/auth/login` | Password login; may return an MFA challenge |
| `POST` | `/v1/auth/mfa/verify` | Complete an MFA login challenge |
| `POST` | `/v1/auth/token/refresh` | Rotate a refresh token and issue a token pair |
| `POST` | `/v1/auth/register/individual` | Create an individual account |
| `POST` | `/v1/auth/register/org` | Create/join an organization by email domain |
| `POST` | `/v1/auth/register/invitation` | Register using an invitation token |
| `POST` | `/v1/auth/password/forgot` | Request a reset; always responds `204` to prevent enumeration |
| `POST` | `/v1/auth/password/reset` | Reset a password using a token |

Common JSON fields are defined in `internal/domain/models`. Login uses `tenant_id`, `email`, and `password`; refresh uses `refresh_token`; registration uses `email`, `password`, and `full_name` (invitation registration uses `invitation_token`).

## Protected self-service routes

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/me/logout` | Revoke the session identified by `X-Session-ID` |
| `POST` | `/v1/me/logout-all` | Revoke all current-user sessions |
| `POST` | `/v1/me/password/change` | Change password with current/new password |
| `GET` | `/v1/me/mfa` | List MFA methods |
| `POST` | `/v1/me/mfa/totp/enroll` | Create a pending TOTP enrollment |
| `POST` | `/v1/me/mfa/totp/activate` | Activate with `mfa_id` and six-digit `code` |
| `DELETE` | `/v1/me/mfa/:mfa_id` | Disable an MFA method |
| `GET` | `/v1/me/sessions` | List sessions |
| `DELETE` | `/v1/me/sessions/:session_id` | Revoke a session |
| `GET` | `/v1/me/devices` | List devices |
| `DELETE` | `/v1/me/devices/:device_id` | Remove a device |
| `GET` | `/v1/me/audit-logs` | List the current user's audit records |

## Protected tenant routes

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/tenants/:tenant_id/invitations` | Invite an email with `role_id` |
| `GET` | `/v1/tenants/:tenant_id/invitations` | List invitations |
| `DELETE` | `/v1/tenants/:tenant_id/invitations/:invitation_id` | Revoke an invitation |
| `POST` | `/v1/tenants/:tenant_id/invitations/:invitation_id/resend` | Reissue an invitation |
| `GET` | `/v1/tenants/:tenant_id/audit-logs` | List tenant audit records |

## gRPC

The server registers `auth.AuthService`, `auth.UserService`, and `auth.SessionService` on port `50051` by default. Send access tokens as `authorization: Bearer <token>` metadata for protected methods. Protobuf sources are under `api/proto/`.

Handler methods that correctly override the generated contracts are:

- `AuthService.Login`
- `AuthService.RefreshToken`
- `UserService.CreatePersonalUser`
- `UserService.CreateCompanyAndOwner`
- `SessionService.ListSessions`
- `SessionService.RevokeSession`
- `SessionService.RevokeAllSessions`

The other declared AuthService methods are not implemented. Calls to those methods return gRPC `Unimplemented` rather than invoking domain behavior.
