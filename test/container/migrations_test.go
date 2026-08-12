package container

import (
	"database/sql"
	"testing"
)

func TestMigrationsCleanBootstrap(t *testing.T) {
	pc := SetupPostgresContainer(t)
	defer TeardownPostgresContainer(t, pc)

	RunMigrations(t, pc.Database, pc.DB)
	assertMigrationVersion(t, pc.DB, 1)
	assertAuthoritativeSchema(t, pc.DB)
}

func assertMigrationVersion(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var got int
	var dirty bool
	if err := db.QueryRow(`SELECT version, dirty FROM schema_migrations`).Scan(&got, &dirty); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if got != want || dirty {
		t.Fatalf("migration state = version %d dirty %t, want version %d dirty false", got, dirty, want)
	}
}

func assertAuthoritativeSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	for _, column := range []struct{ table, name string }{
		{"tenants", "type"},
		{"password_resets", "updated_at"},
		{"audit_logs", "target_id"},
		{"audit_logs", "metadata"},
		{"audit_logs", "trace_id"},
	} {
		var exists bool
		if err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
			)
		`, column.table, column.name).Scan(&exists); err != nil {
			t.Fatalf("inspect %s.%s: %v", column.table, column.name, err)
		}
		if !exists {
			t.Errorf("authoritative column %s.%s is missing", column.table, column.name)
		}
	}

	var tableCount int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE' AND table_name <> 'schema_migrations'
	`).Scan(&tableCount); err != nil {
		t.Fatalf("count authoritative tables: %v", err)
	}
	if tableCount != 11 {
		t.Errorf("authoritative table count = %d, want 11", tableCount)
	}

	var deleteRule string
	if err := db.QueryRow(`
		SELECT rc.delete_rule
		FROM information_schema.referential_constraints rc
		WHERE rc.constraint_schema = 'public' AND rc.constraint_name = 'sessions_device_id_fkey'
	`).Scan(&deleteRule); err != nil {
		t.Fatalf("inspect sessions device foreign key: %v", err)
	}
	if deleteRule != "SET NULL" {
		t.Errorf("sessions device delete rule = %s, want SET NULL", deleteRule)
	}
}
