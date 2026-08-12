## Summary

<!-- What changed, and why? -->

## Behavior and compatibility

<!-- Note API, protobuf, schema, configuration, or backwards-compatibility changes. -->

## Security and privacy impact

<!-- Cover authentication, authorization, tenant isolation, secrets, tokens, audit logs, and rate limits. Write "None" with a reason if not applicable. -->

## Verification

<!-- List exact commands run and their results. Do not mark checks that were not executed. -->

- [ ] Changed Go files were formatted with `gofmt`
- [ ] Relevant focused tests pass
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes, or the omission is explained
- [ ] Docker/Testcontainers checks pass when applicable

## Documentation

- [ ] Current behavior is reflected in `docs/`
- [ ] Contract and architecture changes are reflected in `docs/`
- [ ] Protobuf output was regenerated from `api/proto` when applicable
- [ ] Migration and environment-variable changes are documented
- [ ] No secrets, real identity data, or raw credentials are included

## Release notes

<!-- Describe the user-visible change, compatibility impact, migration, or write "None" with a reason. -->

## GitHub metadata

<!-- Complete these in the PR sidebar and Auth Haven Roadmap. Native Issue Priority does not apply to pull requests. -->

- [ ] At least one suitable default label is applied
- [ ] An accountable assignee is set
- [ ] A release milestone is selected
- [ ] Related issues are linked through GitHub's Development section
- [ ] The PR is added to Auth Haven Roadmap
- [ ] Roadmap Status, Area, Risk, and Phase are set

## Follow-up work

<!-- List known limitations or linked follow-ups. -->
