# Contributing

Thank you for improving Auth Haven. Identity-system changes can affect
authentication, authorization, privacy, and every tenant using the service.
Contributions must therefore be focused, reviewable, and supported by evidence.

By participating, you agree to follow the
[Code of Conduct](CODE_OF_CONDUCT.md). Contributions are submitted under the
[Apache License 2.0](LICENSE).

## Start with project context

Before changing the repository:

1. Read [PROJECT.md](PROJECT.md), [the architecture guide](docs/architecture.md),
   and the documentation relevant to the change.
2. Verify current behavior in executable code, migrations, server wiring, and
   generated protobuf contracts. Documentation can drift.
3. Search [issues](https://github.com/MysteriousZero/auth-haven/issues) and
   [pull requests](https://github.com/MysteriousZero/auth-haven/pulls) for
   related work.
4. Discuss large, compatibility-breaking, security-sensitive, or architectural
   changes in an issue before implementation.
5. Never include credentials, environment files, real identity data, tokens,
   MFA material, or private keys in commits, tests, logs, or screenshots.

AI-assisted contributions must also follow [AGENTS.md](AGENTS.md) and the
[AI development guide](docs/ai-development.md). The human contributor remains
accountable for understanding, verifying, and reviewing generated changes.

## Development environment

Use the included VS Code Dev Container or follow the
[getting-started guide](docs/getting-started.md). The expected local tooling
includes Go 1.25.1 or newer, PostgreSQL, Redis, and Docker. Docker is required
for Testcontainers integration tests.

```bash
docker compose up -d db
export MFA_ENCRYPTION_KEY="$(openssl rand -base64 32)"
# Configure DB_* and REDIS_* variables.
go run ./cmd/migrate
go run ./cmd/server
```

Default listeners are HTTP `:8080` and gRPC `:50051`. The root Compose file
starts PostgreSQL only; provide Redis separately. See the
[configuration reference](docs/configuration.md) before relying on defaults.

## Branch and change scope

- Create a focused branch from the latest `main`.
- Keep unrelated formatting, refactors, generated output, and dependency
  updates out of the contribution.
- Prefer small pull requests that can be reviewed and reverted independently.
- Do not rewrite shared migration history. Add a forward migration unless
  maintainers explicitly approve consolidation.
- Do not edit generated files under `pkg/proto` directly.
- Preserve user changes in a dirty working tree and stage only intended files.

## Implementation rules

Follow the dependency direction described in the architecture guide:

1. Update domain models and interfaces when the contract changes.
2. Implement business rules and authorization in services.
3. Implement persistence and cache behavior behind domain interfaces.
4. Add migrations for durable schema changes.
5. Keep HTTP and gRPC adapters thin and wire dependencies in `internal/server`.
6. Regenerate `pkg/proto` whenever `api/proto` changes.
7. Update documentation and examples in the same pull request.
8. Add tests at the lowest useful layer and regression tests for defects.

Tenant-scoped operations must derive and validate tenant context rather than
trusting tenant identifiers supplied in paths, request bodies, query strings, or
metadata. Security-sensitive mutations must preserve authorization, audit
context, revocation, cache invalidation, and enumeration resistance.

## Validation

Run the narrowest useful checks while developing and the broad checks relevant
to the final change:

```bash
gofmt -w path/to/changed.go
go vet ./...
go test ./...
go test -race ./...
```

Integration tests require Docker. Also validate the affected contract:

| Change | Required evidence |
|---|---|
| Go behavior | Focused tests plus `go test ./...` |
| Concurrency, sessions, caches, or authentication state | `go test -race ./...` |
| Protobuf | Regenerated output and generated-code diff review |
| Migration | Forward and rollback/repair strategy; clean-database validation |
| HTTP or gRPC contract | Request/response or RPC tests and documentation |
| Configuration | Default, failure, and production-safety behavior |
| Documentation only | Link and Markdown checks; no unnecessary code-quality run |

If a check cannot run, explain why in the pull request. Never state that a check
passed when it was skipped or unavailable.

## Documentation and compatibility

Update the canonical files under `docs/` whenever behavior, architecture,
configuration, APIs, storage, testing, or security assumptions change. Clearly
separate implemented behavior from plans and requirements.

Call out all:

- API, protobuf, schema, and configuration changes
- New environment variables and changed defaults
- Migrations, data repair, and rollback considerations
- Breaking changes, deprecations, and client migration guidance
- Security, privacy, tenant-isolation, and audit implications

Release-specific behavior belongs in
[GitHub Releases](https://github.com/MysteriousZero/auth-haven/releases), not in
the root README.

## Commits

Use short, imperative subjects such as `fix refresh token rotation` or
`add session handler tests`. Keep commits coherent and avoid mixing mechanical
rewrites with behavioral changes. A strict conventional-commit format is not
currently required.

## Pull requests

Use the repository pull request template and complete every applicable section:

- Summary and linked work
- Behavior and compatibility
- Security and privacy impact
- Exact verification performed
- Documentation and generated-file impact
- Release notes or a reason none are needed
- Known limitations and follow-up work

Every pull request must also have suitable GitHub metadata:

- At least one appropriate default repository label
- An accountable assignee
- A version milestone
- Membership in the Auth Haven Roadmap
- Roadmap Status, Area, Risk, and Phase

Native Issue Priority applies to issues, not pull requests. Link an issue with
`Closes #123` only when merging the pull request will fully resolve it.

## Review expectations

Reviewers should verify the implementation rather than relying only on the pull
request description. Pay particular attention to:

- Authentication and recovery success and failure paths
- Authorization and cross-tenant attempts
- Token issuance, rotation, expiry, and revocation
- Cryptographic key and secret handling
- Cache consistency after security-state changes
- Audit completeness without sensitive payloads
- Migration safety and compatibility
- Public API and generated-contract drift
- Information leakage and enumeration behavior

Resolve substantive review conversations before merge. Maintainers may request
smaller scope, additional tests, independent security review, migration
evidence, or documentation before accepting a change.

## Reporting security concerns

Do not open a public issue for suspected vulnerabilities. Follow the private
reporting process in [SECURITY.md](SECURITY.md).
