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
| `DB_MAX_CONNECTIONS` | `25` | Configured maximum connection count |
| `DB_MAX_IDLE_CONNS` | `5` | Configured idle connection count |
| `DB_CONN_MAX_LIFETIME` | `5m` | Configured connection lifetime |

The current server opens the database using `database/sql`, but does not apply the three pool tuning values when constructing the connection.

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
