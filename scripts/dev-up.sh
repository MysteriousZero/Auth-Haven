#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
"$repo_dir/scripts/dev-env.sh"
cd "$repo_dir"
exec ./scripts/dev-compose.sh --env-file .env.development up --build --wait
