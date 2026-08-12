#!/usr/bin/env bash
set -euo pipefail

protoc \
  --go_out=pkg/proto --go_opt=paths=source_relative \
  --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
  -I api/proto \
  api/proto/AuthService.proto \
  api/proto/InviationService.proto \
  api/proto/UserService.proto \
  api/proto/SessionService.proto \
  api/proto/common/Tokens.proto

if [[ "${1:-}" == "--check" ]]; then
  git diff --exit-code -- pkg/proto
fi
