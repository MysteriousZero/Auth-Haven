# Migration operations

`migrations/000001_init_schema.up.sql` is the only schema migration. It creates the complete schema represented by the ERD plus the `password_resets.updated_at` and `audit_logs.trace_id` fields consumed by current models and repositories. Application startup does not run migrations; operators run `go run ./cmd/migrate` separately.

## Supported upgrades

| Recorded version | Prior lineage | Upgrade behavior |
|---:|---|---|
| no version | Empty database | Applies the complete version-1 baseline |
| current version `1` | Current baseline | No change |

Databases created from the removed legacy `0001` or `001`–`007` files are not supported for in-place migration. Do not manually change `schema_migrations` to claim compatibility.

The migration command performs a schema preflight before applying files. It accepts an empty database or the authoritative version-1 schema, and rejects dirty, untracked, removed version-1, and incremental legacy states without modifying application tables. This prevents a legacy database that also reports version 1 from being mistaken for the current baseline.

## Production upgrade

1. Stop writes and take a tested PostgreSQL backup or snapshot.
2. Export required legacy data without migration bookkeeping or obsolete schema objects.
3. Apply the single baseline to a new empty database and verify version 1 with `dirty = false`.
4. Transform and import data under the authoritative constraints, explicitly assigning tenant types and resolving duplicates.
5. Run repository and integration tests before switching application traffic.

The automated migration tests verify clean bootstrap, a no-op rerun against the current baseline, and non-destructive rejection of representative removed version-1 and version-7 states. The data transformation itself remains deployment-specific and is not claimed as an automated in-place upgrade.

The repository does not contain destructive conversion SQL for legacy databases. Keep the original database available until row counts, tenant boundaries, credential state, audit evidence, and application behavior are verified in the rebuilt database.

## Rollback and failure handling

There is no automated down migration because dropping the baseline would destroy all application data. If bootstrap or import fails, discard the new database, correct the process, and retry from the verified legacy backup. Never force migration state or overwrite the source database.
