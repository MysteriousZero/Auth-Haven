# Support

Auth Haven is an active development project and does not currently offer a service-level agreement or guaranteed support window.

## Where to ask

- Use GitHub issues for reproducible bugs, feature proposals, and documentation problems.
- Use GitHub discussions, if enabled, for general design and usage questions.
- Follow [SECURITY.md](SECURITY.md) for vulnerabilities. Never report a vulnerability in a public issue.

Before opening a report, read the [documentation index](docs/README.md), [known gaps](PROJECT.md#known-gaps), and existing issues.

## Useful bug reports

Include:

- Expected and actual behavior
- Minimal reproduction steps
- Relevant commit, Go version, operating system, and execution method
- Sanitized logs or error output
- Whether Docker, PostgreSQL, Redis, HTTP, or gRPC is involved

Remove credentials, personal data, tokens, MFA material, connection strings containing passwords, and `.env` contents.

## Compatibility questions

The project has no published stable release or compatibility guarantee yet. Pin deployments to a reviewed commit and evaluate migrations and API changes before upgrading.
