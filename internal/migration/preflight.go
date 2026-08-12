package migration

import (
	"context"
	"database/sql"
	"fmt"
)

// ValidateState rejects migration histories that reuse version numbers from the
// authoritative baseline but do not have its schema shape.
func ValidateState(ctx context.Context, db *sql.DB) error {
	var migrationsTableExists bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)
	`).Scan(&migrationsTableExists); err != nil {
		return fmt.Errorf("inspect migration state: %w", err)
	}
	if !migrationsTableExists {
		var tableCount int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`).Scan(&tableCount); err != nil {
			return fmt.Errorf("inspect untracked schema: %w", err)
		}
		if tableCount != 0 {
			return fmt.Errorf("database has an untracked legacy schema; rebuild using docs/migrations.md")
		}
		return nil
	}

	var version int
	var dirty bool
	if err := db.QueryRowContext(ctx, `SELECT version, dirty FROM schema_migrations`).Scan(&version, &dirty); err != nil {
		return fmt.Errorf("read migration state: %w", err)
	}
	if dirty {
		return fmt.Errorf("database migration state is dirty at version %d", version)
	}
	if version != 1 {
		return fmt.Errorf("unsupported legacy migration version %d; rebuild using docs/migrations.md", version)
	}

	for _, column := range []struct{ table, name string }{
		{"tenants", "type"},
		{"password_resets", "updated_at"},
		{"audit_logs", "trace_id"},
	} {
		var exists bool
		if err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
			)
		`, column.table, column.name).Scan(&exists); err != nil {
			return fmt.Errorf("inspect authoritative schema: %w", err)
		}
		if !exists {
			return fmt.Errorf("migration version 1 has legacy schema shape (missing %s.%s); rebuild using docs/migrations.md", column.table, column.name)
		}
	}
	return nil
}
