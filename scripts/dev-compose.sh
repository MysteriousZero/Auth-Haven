#!/bin/sh
set -eu

if command -v docker >/dev/null 2>&1; then
  exec docker compose "$@"
fi

if command -v podman >/dev/null 2>&1; then
  exec podman compose "$@"
fi

echo "Docker Compose or Podman Compose is required" >&2
exit 1
