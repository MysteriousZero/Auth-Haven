# Validation Layer Specification

This document defines the strict validation rules enforced at the boundary of the Auth Haven application. Validation is split into **Structural Validation** (performed at the handler/interceptor level) and **Domain Validation** (performed by the core services).

## 1. Structural Validation

Structural validation ensures that the incoming request payloads conform to basic data types, lengths, and formats. We use `buf.build/gen/go/envoyproxy/protoc-gen-validate` for gRPC and `go-playground/validator/v10` for REST.

Requests failing structural validation immediately return `HTTP 400 Bad Request` or `gRPC InvalidArgument (3)`, along with a detailed map of field errors.

### Common Field Rules

| Field | Rule/Tag | Description |
|---|---|---|
| **IDs** | `uuid` | All tenant, user, session, and device IDs must be valid UUIDv4 strings. |
| **Email** | `email,max=255` | Must be a valid standard email format (RFC 5322) and maximum 255 characters. |
| **Names / Strings** | `min=1,max=255` | General string fields (e.g., `full_name`) cannot be empty or excessively long to prevent buffer issues. |
| **Tokens / Codes** | `len=6` (MFA) / `min=32` | MFA codes are strictly 6 numeric characters. Generated tokens (temp, refresh) must meet minimum entropy lengths. |
| **Role IDs** | `gt=0` | Role IDs (if provided) must be positive integers. |

### Password Strength Requirements

All passwords must be evaluated structurally *before* hashing. Passwords must pass the following criteria:

*   **Minimum Length:** 12 characters.
*   **Maximum Length:** 128 characters (to prevent bcrypt/argon2 hashing DoS attacks).
*   **Complexity:**
    *   At least 1 uppercase letter `[A-Z]`
    *   At least 1 lowercase letter `[a-z]`
    *   At least 1 number `[0-9]`
    *   At least 1 special character (e.g., `!@#$%^&*()_+-=[]{}|;':",./<>?`)

## 2. Domain Validation

Domain validation requires database or state lookups and belongs strictly in the **Service Layer**.

If a domain validation fails, the service returns specific domain errors which map to HTTP `400/403/404/409` or gRPC `PermissionDenied / AlreadyExists / Unauthenticated`.

### Key Domain Rules

1.  **Email Uniqueness Constraints:** Emails must be strictly unique **per tenant**. Identical emails can exist in different organizational tenants.
2.  **Domain Allowance (Org Tenants):** When registering an organization, the email domain cannot be a public provider (e.g., `@gmail.com`, `@yahoo.com`). We use a deny-list of public domains.
3.  **Cross-Tenant Boundaries:** A user ID must actually belong to the `tenant_id` active in the request context. Attempting to manage a user ID belonging to a different tenant results in an `ErrForbidden` or `ErrTenantNotFound` to deter enumeration.
4.  **MFA State Checks:** Attempting to enroll a user in TOTP who is already enrolled will fail. Attempting to login without an MFA code when the user has `enabled=true` will pause the flow and return an MFA requirement error.