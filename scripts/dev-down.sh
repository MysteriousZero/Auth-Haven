#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
env_file="$repo_dir/.env.development"

if [ ! -f "$env_file" ]; then
  echo "No local development environment exists."
  exit 0
fi

cd "$repo_dir"
exec docker compose --env-file .env.development down "$@"
