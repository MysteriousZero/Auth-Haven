# Auth Haven Copilot instructions

Follow the repository-wide instructions in [`AGENTS.md`](../AGENTS.md) for every task. Treat that file as canonical; do not duplicate or override its rules here.

Before suggesting or editing code:

1. Read `PROJECT.md`, `docs/architecture.md`, and `docs/ai-development.md`.
2. Trace the actual constructor wiring and registered route/RPC before assuming a feature exists.
3. Use code and migrations as the source of current behavior; label documentation-only plans explicitly.
4. Keep business rules and authorization in services, enforce tenant isolation, and never expose secrets or raw credentials.
5. Update tests and documentation with behavior changes and report only checks actually executed.

Do not hand-edit generated files under `pkg/proto`; edit `api/proto` and regenerate them.
