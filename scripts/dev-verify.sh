#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_dir"
./scripts/dev-env.sh

compose() {
  ./scripts/dev-compose.sh --env-file .env.development "$@"
}

restore_redis() {
  compose start redis >/dev/null 2>&1 || true
}

echo "Checking HTTP readiness and gRPC reachability..."
compose exec -T app sh -c 'wget -q -O /dev/null http://127.0.0.1:8080/ready && nc -z 127.0.0.1 50051'

echo "Checking PostgreSQL volume persistence across a restart..."
compose exec -T db sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "CREATE TABLE IF NOT EXISTS dev_stack_persistence_probe (value integer PRIMARY KEY); INSERT INTO dev_stack_persistence_probe VALUES (1) ON CONFLICT DO NOTHING;"' >/dev/null
compose restart db >/dev/null
compose exec -T db sh -c 'until pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; do sleep 1; done; psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "SELECT value FROM dev_stack_persistence_probe WHERE value = 1"' | grep -q '^1$'
compose exec -T db sh -c 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "DROP TABLE dev_stack_persistence_probe"' >/dev/null

echo "Checking fail-closed readiness when Redis is unavailable..."
trap restore_redis EXIT HUP INT TERM
compose stop redis >/dev/null
if compose exec -T app wget -q -O /dev/null http://127.0.0.1:8080/ready; then
  echo "readiness remained healthy while Redis was stopped" >&2
  exit 1
fi
compose start redis >/dev/null
compose up --wait app >/dev/null
trap - EXIT HUP INT TERM

echo "Development stack verification passed."
