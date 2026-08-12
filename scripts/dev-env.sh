#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
env_file="$repo_dir/.env.development"

if [ -f "$env_file" ]; then
  exit 0
fi

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required to generate development credentials" >&2
  exit 1
fi

umask 077
postgres_password=$(openssl rand -hex 24)
redis_password=$(openssl rand -hex 24)
mfa_key=$(openssl rand -base64 32 | tr -d '\n')
temporary_file="$env_file.tmp.$$"
trap 'rm -f "$temporary_file"' EXIT HUP INT TERM

cat >"$temporary_file" <<EOF
# Generated local-development credentials. Never use this file in production.
APP_ENV=development
POSTGRES_DB=auth_haven
POSTGRES_USER=auth_haven
POSTGRES_PASSWORD=$postgres_password
POSTGRES_PORT=5432
HOST_POSTGRES_PORT=5432
DB_HOST=db
DB_PORT=5432
DB_NAME=auth_haven
DB_USER=auth_haven
DB_PASSWORD=$postgres_password
DB_SSLMODE=disable
REDIS_HOST=redis
REDIS_PORT=6379
HOST_REDIS_PORT=6379
REDIS_PASSWORD=$redis_password
REDIS_DB=0
MFA_ENCRYPTION_KEY=$mfa_key
JWT_KEY_ID=development-ephemeral
SERVER_HOST=0.0.0.0
HTTP_PORT=:8080
GRPC_PORT=:50051
HOST_HTTP_PORT=8080
HOST_GRPC_PORT=50051
CORS_ALLOWED_ORIGINS=http://localhost:3000
EOF

mv "$temporary_file" "$env_file"
trap - EXIT HUP INT TERM
echo "Created $env_file with development-only credentials."
