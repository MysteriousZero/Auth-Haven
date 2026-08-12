## Summary

<!-- What changed, and why? -->

## Behavior and compatibility

<!-- Note API, protobuf, schema, configuration, or backwards-compatibility changes. -->

## Security impact

<!-- Cover authentication, authorization, tenant isolation, secrets, tokens, audit logs, and rate limits. Write "None" with a reason if not applicable. -->

## Verification

<!-- List exact commands run and their results. Do not mark checks that were not executed. -->

- [ ] Changed Go files were formatted with `gofmt`
- [ ] Relevant focused tests pass
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes, or the omission is explained
- [ ] Docker/Testcontainers checks pass when applicable

## Documentation and generated files

- [ ] Current behavior is reflected in `docs/`
- [ ] Contract and architecture changes are reflected in `docs/`
- [ ] Protobuf output was regenerated from `api/proto` when applicable
- [ ] Migration and environment-variable changes are documented
- [ ] No secrets, real identity data, or raw credentials are included

## Follow-up work

<!-- List known limitations or linked follow-ups. -->
