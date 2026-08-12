# Contributing

By participating, you agree to follow the
[Code of Conduct](CODE_OF_CONDUCT.md). Contributions are submitted under the
project's [Apache License 2.0](LICENSE).

Thank you for improving Auth Haven. Authentication changes can have broad security consequences, so contributions should be small, explicit, and supported by tests.

## Before starting

1. Read [PROJECT.md](PROJECT.md), [docs/architecture.md](docs/architecture.md), and the relevant guides under `docs/`.
2. Confirm current behavior in code. Documentation may be ahead of—or inconsistent with—the implementation.
3. Search existing issues or pull requests before beginning a large change.
4. Do not include credentials, `.env` contents, real user data, token values, or private keys in commits, tests, logs, or screenshots.

AI-assisted contributions must also follow [AGENTS.md](AGENTS.md) and [docs/ai-development.md](docs/ai-development.md). The contributor remains responsible for reviewing and understanding generated changes.

## Development setup

Use the VS Code Dev Container or follow [docs/getting-started.md](docs/getting-started.md). Integration tests use Testcontainers and require Docker.

Create a focused branch and keep unrelated changes out of the contribution. Do not rewrite existing migration history after it has been shared; add a new migration unless maintainers explicitly approve consolidation.

## Change workflow

1. Update domain models and interfaces when the contract changes.
2. Implement business behavior in services.
3. Implement repository behavior and migrations where required.
4. Add thin HTTP/gRPC adapters and wire them in `internal/server`.
5. Update `docs/`, clearly distinguishing current behavior from plans or requirements.
6. Add tests at the lowest useful layer and regression tests for bugs.
7. Regenerate `pkg/proto` whenever an `api/proto` contract changes.

## Required checks

Run the checks available for the affected scope:

```bash
gofmt -w path/to/changed.go
go vet ./...
go test ./...
go test -race ./...
```

If a check cannot run, state why in the pull request. Never claim a check passed when it was not executed.

## Pull requests

Describe:

- The problem and expected behavior
- The implementation approach
- Security and tenant-isolation impact
- Schema, API, configuration, or compatibility changes
- Tests executed and their results
- Known limitations or follow-up work

Keep generated files in the same commit as their source contract. Call out migrations and new environment variables prominently. Screenshots are usually unnecessary for this backend project; request/response examples or test output are more useful.

## Review expectations

Changes require special scrutiny when they affect authentication, authorization, cryptography, tokens, tenant filtering, audit data, rate limiting, migrations, or public API contracts. Reviewers should verify both success and failure paths, including cross-tenant attempts and information leakage.

## Commit style

Use short, imperative subjects such as `fix refresh token rotation` or `add session handler tests`. A strict conventional-commit format is not currently required.
