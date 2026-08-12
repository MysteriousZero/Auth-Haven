# Continuous integration

Pull requests to `main` run independent required gates so failures identify the affected concern without granting broad token permissions.

| Check | Purpose | Expected runtime |
|---|---|---:|
| `Quality` | Formatting, vetting, and non-container Go tests | under 10 minutes |
| `Contracts and docs` | Protobuf regeneration drift, local Markdown links, and issue-form YAML | under 10 minutes |
| `Race` | Race detector for application and command packages | under 15 minutes |
| `Testcontainers and migrations` | Race-enabled PostgreSQL/Redis repositories, bootstrap, idempotency, and legacy-state rejection | under 25 minutes |
| `Secret scan` | Full-history credential scanning with gitleaks v8.30.1 | under 10 minutes |
| `Dependency review` | Blocks newly introduced high or critical dependency vulnerabilities | under 10 minutes |
| `Analyze Go` | CodeQL security-and-quality analysis | under 15 minutes |
| `Validate pull request metadata` | Required PR sections, label, assignee, and milestone | under 5 minutes |

All third-party actions are pinned to immutable commit SHAs. Workflow permissions default to read-only. CodeQL alone receives `security-events: write`; no pull-request workflow executes repository secrets or uses `pull_request_target`.

## Local checks

Run the CI-equivalent checks before pushing:

```bash
test -z "$(gofmt -l $(git ls-files '*.go'))"
go vet ./...
go test ./internal/... ./cmd/...
go test -race ./internal/... ./cmd/...
go test -race -count=1 -timeout=20m ./test/...
./scripts/generate-proto.sh --check
go run ./tools/ci validate
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
```

The Testcontainers command requires a working Docker-compatible daemon. Protobuf drift checking requires `protoc`, `protoc-gen-go` v1.36.11, and `protoc-gen-go-grpc` v1.6.1.

## Failures and reruns

Open the failed job and use the first failing command rather than rerunning the whole workflow immediately. Generated-code failures require regenerating and committing `pkg/proto`; documentation failures identify the missing local target or invalid issue form. Testcontainers dependency-outage failures should report the Docker daemon or image-pull error without printing environment variables.

A test is flaky only after the same commit produces both a pass and failure without relevant external dependency failure. Record the failing seed/log, open an issue, and fix or quarantine it with an owner and expiry. Do not repeatedly rerun a required check until it passes, and do not weaken timeouts to hide deadlocks.

CI uploads no database contents, environment dumps, coverage profiles, or test logs as retained artifacts. GitHub job logs remain subject to repository retention settings and must never contain secrets, tokens, MFA material, private keys, or real identity data.

## Security triage and suppression

`CODEOWNERS` assigns workflow and security-policy review. Treat secret-scan and CodeQL alerts as security-sensitive: validate them privately, rotate exposed credentials before code cleanup, and use the process in [`SECURITY.md`](../SECURITY.md). Dependency-review failures should be remediated by avoiding, upgrading, or replacing the dependency.

Suppress an alert only when a maintainer documents the rule, affected path, evidence that it is a false positive or accepted risk, compensating control, owner, and review date. Inline broad exclusions and unbounded allowlists are not acceptable. Security gate changes require code-owner review.

## Required checks

Branch protection must require every check in the table above, use strict up-to-date branches, and apply to administrators. Required check names are contracts: rename a job only while updating branch protection in the same maintenance window.
