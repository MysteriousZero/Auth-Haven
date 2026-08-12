# AI-assisted development

This guide gives coding agents a compact map of the repository and a repeatable way to make safe changes. Repository-wide instructions live in [`AGENTS.md`](../AGENTS.md).

GitHub Copilot receives a small adapter in `.github/copilot-instructions.md`; it delegates to `AGENTS.md` so repository rules have one canonical source.

## Sources of truth

| Question | Inspect first |
|---|---|
| What runs today? | `internal/server/http.go`, `internal/server/grpc.go`, and `cmd/` |
| What does an endpoint accept? | Handler plus `internal/domain/models`; protobuf source for gRPC |
| Where is a business rule? | `internal/service` and `internal/domain/interfaces` |
| How is data stored? | `internal/repository` and `migrations/` |
| What constraints and plans are documented? | `docs/`, especially `PROJECT.md`, architecture, security, and data-model guides |
| What should developers be told? | `docs/` and root policy documents |

Do not infer implementation from filenames or documentation alone. Trace constructor wiring and registered methods, because some services exist without routes and some protobuf RPCs currently use embedded `Unimplemented` behavior.

## Context map by task

| Task | Minimum files to inspect |
|---|---|
| HTTP route | `internal/server/http.go`, handler, model, service interface/implementation, error mapper |
| gRPC method | `.proto`, generated server interface, `internal/handlers/grpc.go`, interceptors, gRPC wiring |
| Authentication/token | auth service, token service, auth repository, auth middleware/interceptor, security specs |
| Tenant authorization | actor/target service checks, tenant/user/role repositories, domain errors |
| Database change | current migrations, repository query/scan code, schema specs, integration fixtures |
| Cache change | base repository, cached wrapper, invalidation paths, Redis tests |
| Configuration | `internal/config/config.go`, all consumers, Compose/dev-container files, configuration docs |
| Documentation | current wiring and tests; distinguish shipped behavior from target design |

## Change patterns

### New HTTP capability

1. Define or reuse domain models and interfaces.
2. Implement authorization and business logic in a service.
3. Add repository behavior if needed.
4. Add a thin handler with structural validation and error mapping.
5. Register middleware and route wiring explicitly.
6. Test the service and handler, including unauthorized and cross-tenant cases.

### New gRPC capability

1. Update the protobuf source.
2. Regenerate Go protobuf files; never hand-edit them.
3. Implement the exact generated method name and signature.
4. Classify the method as public or protected in the interceptor.
5. Construct all non-nil dependencies in gRPC server wiring.
6. Test authentication metadata, validation, and status mapping.

### Schema change

1. Inspect every existing migration for overlapping objects.
2. Prefer an additive, ordered migration.
3. Update repository queries/scans and test factories atomically.
4. Describe data migration, compatibility, and rollback implications.
5. Update the data-model and operational documentation.

## Security invariants

- Never authorize solely from a client-supplied tenant or resource ID.
- Check actor membership/permission and target ownership in the service layer.
- Keep invalid-account and invalid-password behavior indistinguishable where anti-enumeration applies.
- Never log or return password hashes, raw persistent tokens, MFA secrets/codes, or private keys.
- Use existing hashing, encryption, validation, and randomness utilities.
- Audit security-sensitive mutations and propagate trace/client metadata.
- Invalidate related cache entries when persistent authentication state changes.

## Completion format

An AI-generated handoff should state:

- Outcome and files changed
- Behavior or contract decisions
- Tests/checks actually executed and results
- Checks not run and the reason
- Remaining risks or follow-ups

Avoid percentage-based completeness claims unless produced by a repeatable measurement command.
