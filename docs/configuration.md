# Configuration

All configuration is read from environment variables by `internal/config`. Invalid integers, booleans, or durations silently fall back to defaults. Durations use Go syntax such as `15m`, `24h`, or `168h`.

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
| `JWT_PRIVATE_KEY` | empty | Base64/raw representation accepted by the token implementation; empty generates an ephemeral key |
| `ACCESS_TOKEN_TTL` | `15m` | Access-token lifetime |
| `REFRESH_TOKEN_TTL` | `168h` | Refresh-token lifetime |
| `TEMP_TOKEN_TTL` | `10m` | Temporary MFA-token lifetime |
| `SESSION_TTL` | `24h` | Session lifetime |

Use a persistent signing key in production. An ephemeral key invalidates existing JWTs whenever the process restarts and is inconsistent across replicas.

The current token service hard-codes token lifetimes rather than consuming these configured TTL values. Treat the table as the configuration surface intended by `internal/config`, and verify usage before relying on overrides.

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
