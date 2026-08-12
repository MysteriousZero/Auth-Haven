# AI agent instructions

These instructions apply to the entire repository.

## Start here

Before changing code, read:

1. `PROJECT.md`
2. `docs/architecture.md`
3. `docs/ai-development.md`
4. The relevant API, security, data-model, and implementation guides

Use executable code and migrations as the source of current behavior. Documentation captures architecture, constraints, and explicitly marked plans. If they disagree, identify the drift explicitly and update documentation only when the task authorizes it.

## Repository rules

- Keep the dependency direction: transport → service → domain interface → repository.
- Put business rules and authorization in services, not handlers.
- Keep HTTP and gRPC adapters thin and consistent.
- Enforce tenant isolation on every tenant-scoped operation.
- Reuse domain errors and central transport mappings.
- Preserve request trace, client IP, user-agent, actor, and tenant context for audit events.
- Never expose or log passwords, raw long-lived tokens, MFA codes/secrets, private keys, or `.env` values.
- Do not edit generated `pkg/proto/*.pb.go` files manually. Change `api/proto/*.proto` and regenerate.
- Do not modify an existing applied migration casually. Prefer an additive migration and document rollback/compatibility implications.
- Preserve user changes and avoid unrelated cleanup.

## Required task workflow

1. Inspect the affected call path and existing tests before editing.
2. State assumptions when documentation and implementation differ.
3. Make the smallest coherent change across model, interface, service, repository, transport, and wiring layers as applicable.
4. Add or update tests for behavior changes and regressions.
5. Update `docs/` for behavior, contract, architecture, or operational changes; label unimplemented plans clearly.
6. Run formatting and the narrowest relevant tests, then broader checks when available.
7. Report exactly what ran, what failed, and what could not run.

## Verification

For Go changes, prefer:

```bash
gofmt -w path/to/changed.go
go vet ./...
go test ./...
go test -race ./...
```

Tests under `test/` may require Docker because they use Testcontainers. Never report a test as passing unless its command completed successfully.

## High-risk review checklist

For authentication, authorization, token, MFA, password, invitation, or session changes, verify:

- Cross-tenant access is rejected.
- Missing, malformed, expired, revoked, and replayed credentials fail safely.
- Error responses do not enable user or tenant enumeration.
- Mutations are audited without sensitive values.
- Token/session revocation and cache invalidation stay consistent.
- Concurrent and transactional failure paths cannot leave partial state.
