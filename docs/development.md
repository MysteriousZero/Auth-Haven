# Development guide

Use [`CONTRIBUTING.md`](../CONTRIBUTING.md) for setup, implementation order,
validation, commits, pull requests, and review requirements. Coding agents must
also follow [`AGENTS.md`](../AGENTS.md) and the
[AI-assisted development guide](ai-development.md). Test commands and current
coverage are maintained in the [testing guide](testing.md).
Required automation and troubleshooting are maintained in the [CI guide](ci.md).

## Local stack

Run `./scripts/dev-up.sh` for the shared PostgreSQL, Redis, migration, and
application stack. The VS Code devcontainer uses those same dependency and
migration services. Lifecycle, persistence, failure-mode verification, and
development-secret limitations are documented in the
[getting-started guide](getting-started.md).

## Generate protobuf code

Run the pinned repository script from the repository root:

```bash
./scripts/generate-proto.sh
```

Confirm generated package paths and compile the repository after regeneration. `InviationService.proto` is currently misspelled and is not registered by the server.

## HTTP changes

- Define transport-independent data in `internal/domain/models` and behavior in `internal/domain/interfaces`.
- Put business rules in `internal/service`, not handlers.
- Implement persistence in `internal/repository`.
- Add a thin handler under `internal/handlers`.
- Wire dependencies and routes in `internal/server/http.go`.
- Map new domain errors centrally in `internal/domain/errors`.

## gRPC changes

- Update the protobuf contract and regenerate code.
- Implement the generated server method in `internal/handlers/grpc.go`.
- Decide explicitly whether the method is public in `internal/server/interceptors.go`.
- Wire any required service in `internal/server/grpc.go`.

The registered gRPC contract intentionally includes login, MFA verification,
refresh-token rotation, password reset, individual/organization registration,
and authenticated session management. Token revocation by raw refresh token and
token introspection were removed before a stable release because no matching
domain operation or authorization policy exists; use session revocation and
refresh-token rotation instead. Any future RPC must be implemented and assigned
an explicit public or protected classification in the same change.

## Known development hazards

- Documentation and implementation can drift; verify registered routes and constructed dependencies.
- HTTP and gRPC currently build separate repository/service graphs, and gRPC does not use the Redis caching wrappers.
- The gRPC validation interceptor is currently a pass-through.
- CORS uses an environment-configured exact-origin allowlist and rejects wildcard origins in production.
- SMTP delivery is configured through environment variables; production startup requires the host, sender, and public link base URL.
- Role business logic exists, but the HTTP server deliberately leaves `roleService` unwired to a handler.
