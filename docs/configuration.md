# Configuration

All configuration is read from environment variables by `internal/config`. Durations use Go syntax such as `15m`, `24h`, or `168h`. Invalid authentication durations fail startup; some legacy integer, boolean, and non-authentication duration settings still fall back to defaults.

## Database

| Variable | Default | Meaning |
|---|---:|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | empty | Database password |
| `DB_NAME` | `auth_haven` | Database name |
| `DB_SSLMODE` | `disable` | lib/pq SSL mode |
| `DB_MAX_CONNECTIONS` | `25` | Maximum open connection count |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle connection count |
| `DB_CONN_MAX_LIFETIME` | `5m` | Maximum connection lifetime |

The server applies all three values to `database/sql`. Current pool statistics are exposed in Prometheus text format at `GET /metrics`.

## Server

| Variable | Default | Meaning |
|---|---:|---|
| `SERVER_HOST` | `0.0.0.0` | HTTP bind host |
| `HTTP_PORT` | `:8080` | HTTP port, including leading colon |
| `GRPC_PORT` | `:50051` | gRPC listen address/port |

`HTTP_PORT` is concatenated with `SERVER_HOST`; keep its leading colon. The gRPC listener uses `GRPC_PORT` directly.

## Tokens and sessions

| Variable | Default | Meaning |
|---|---:|---|
| `APP_ENV` | `development` | Runtime environment; `production` enables production safety checks |
| `JWT_KEY_ID` | `development` | Identifier written to the JWT `kid` header for the active signing key |
| `JWT_PRIVATE_KEY` | empty | Base64 encoding of a 64-byte Ed25519 private key; required in production |
| `JWT_VERIFICATION_KEYS` | `{}` | JSON object mapping retained key IDs to base64 Ed25519 public keys |
| `ACCESS_TOKEN_TTL` | `15m` | Access-token lifetime |
| `REFRESH_TOKEN_TTL` | `168h` | Refresh-token lifetime |
| `TEMP_TOKEN_TTL` | `10m` | Temporary MFA-token lifetime |
| `SESSION_TTL` | `24h` | Session lifetime |
| `TOKEN_CLOCK_SKEW` | `30s` | Non-negative JWT validation leeway for replica clock differences |

Production startup fails unless `JWT_PRIVATE_KEY` is present and valid. Local development may omit it; the process then creates one ephemeral key shared by its HTTP and gRPC servers. Never log these variables or commit their values.

Generate a key pair, store the private key in a secret manager, and inject it at runtime. Every replica must receive the same active key configuration. The signing-key provider interface is the integration boundary for a KMS-backed implementation.

## Signing-key rotation

1. Generate a new Ed25519 key pair and a unique, immutable key ID.
2. Add the new public key to `JWT_VERIFICATION_KEYS` everywhere while the old key remains active, then deploy and verify propagation.
3. Add the old public key to `JWT_VERIFICATION_KEYS`, switch `JWT_KEY_ID` and `JWT_PRIVATE_KEY` to the new key, and deploy across all replicas.
4. Keep the old public key for at least `ACCESS_TOKEN_TTL + TOKEN_CLOCK_SKEW`.
5. Remove the old entry to revoke any remaining token signed by it.

For emergency invalidation, remove the compromised key from `JWT_VERIFICATION_KEYS`, replace the active key if necessary, and deploy atomically. This invalidates its JWTs immediately. Revoke affected refresh tokens and sessions in PostgreSQL separately because they are opaque credentials rather than JWTs.

## Security

| Variable | Default | Meaning |
|---|---:|---|
| `PASSWORD_MIN_LENGTH` | `12` | Minimum password length |
| `PASSWORD_REQUIRE_UPPER` | `true` | Require uppercase characters |
| `PASSWORD_REQUIRE_LOWER` | `true` | Require lowercase characters |
| `PASSWORD_REQUIRE_NUMBER` | `true` | Require digits |
| `PASSWORD_REQUIRE_SPECIAL` | `true` | Require special characters |
| `MAX_LOGIN_ATTEMPTS` | `5` | Login-attempt setting |
| `ACCOUNT_LOCKOUT_DURATION` | `15m` | Lockout-duration setting |
| `MFA_ENCRYPTION_KEY` | empty | Base64 encoding of exactly 32 bytes |

The server treats a missing or incorrectly sized MFA key as fatal. Generate one with `openssl rand -base64 32`, store it in a secret manager, and do not rotate it without a data migration.

## Redis

| Variable | Default | Meaning |
|---|---:|---|
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | empty | Redis password |
| `REDIS_DB` | `0` | Logical Redis database |
| `REDIS_RATE_LIMIT_FAIL_OPEN` | `false` | Allow protected endpoints to continue when Redis rate limiting is unavailable |

Cache access always fails open to PostgreSQL. Redis is a required startup and readiness dependency when rate limiting is fail-closed. Rate limiting defaults to fail-closed because silently bypassing brute-force controls is unsafe.

## Browser clients

| Variable | Default | Meaning |
|---|---:|---|
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated exact browser origins |
| `CORS_ALLOW_CREDENTIALS` | `false` | Permit credentialed cross-origin requests |

Production requires at least one explicit origin and rejects `*`. Credentialed CORS can never be combined with a wildcard.

## Email delivery

| Variable | Default | Meaning |
|---|---:|---|
| `SMTP_HOST` | empty | SMTP server host |
| `SMTP_PORT` | `587` | SMTP server port |
| `SMTP_USERNAME` | empty | Optional SMTP authentication username |
| `SMTP_PASSWORD` | empty | Optional SMTP authentication password |
| `EMAIL_FROM` | empty | Envelope and message sender |
| `PUBLIC_BASE_URL` | `http://localhost:3000` | Public application URL used in reset and invitation links |

Production requires `SMTP_HOST`, `EMAIL_FROM`, and `PUBLIC_BASE_URL`. Delivery failures never log raw reset or invitation tokens. Password-reset responses remain enumeration-resistant; operators observe delivery failures in server logs. Invitation delivery failures are returned to the authenticated caller after persistence so they can retry.
