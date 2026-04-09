# Audit Logging Specification

Auth Haven maintains a strict, immutable audit log for all security-relevant and state-mutating actions to aid in compliance (SOC2/GDPR) and incident investigation.

## 1. Audit Scope

The following actions trigger an immediate audit log entry:
*   **Authentication Events:** Login success, login failure, logout, MFA verification success/failure.
*   **Credential Mutations:** Password change, password reset requested, password reset completed.
*   **MFA Mutations:** TOTP/SMS enrolled, activated, deactivated.
*   **User/Tenant Mutations:** User registration, user suspension, role changes, tenant creation, tenant suspension.
*   **Invitations:** Invitation sent, accepted, revoked.

Read-only actions (like listing sessions or listing users) are **not** audited to prevent database bloat, unless explicitly required for high-security environments.

## 2. Event Structure

Every audit event is recorded synchronously (or via a reliable background queue wrapper) and stored in the `audit_logs` table.

The required fields for every event include:
*   **`log_id`**: UUIDv4.
*   **`user_id`**: The user performing the action (actor). If the action is performed by the system (e.g., an automated expiry), this is `NULL` or marked `SYSTEM`.
*   **`tenant_id`**: The context in which the action occurred.
*   **`action`**: A dot-notated string classifying the event (e.g., `user.login.success`, `user.password.reset`).
*   **`ip_address`**: The raw IP of the client (extracted securely behind proxies using `X-Forwarded-For`).
*   **`user_agent`**: The client's User-Agent string.
*   **`created_at`**: UTC timestamp.

## 3. Sensitive Data Masking (Data Privacy)

Audit logs must NEVER contain raw sensitive information. The logging interceptors and repository methods must enforce the following filters:
*   **Passwords & Secrets:** Never logged in any format.
*   **Raw Tokens:** Access tokens, refresh tokens, and temp MFA tokens are never logged.
*   **PII:** Email addresses and Phone numbers are strictly omitted from the `action` or metadata payload strings. Instead, the `user_id` is logged, which serves as the foreign key to resolve PII strictly when requested by authorized admins.

## 4. Retention Policy

*   **Active DB:** Audit logs are kept in the hot operational database for **90 days** to allow tenant administrators to view their recent security history via the `/v1/tenants/{tenant_id}/audit-logs` endpoint.
*   **Cold Storage:** After 90 days, a background CRON job archives the logs to an external blob store (e.g., AWS S3 glacier) and deletes them from the Postgres `audit_logs` table to maintain query performance.