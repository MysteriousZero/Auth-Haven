package container

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	databasepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies the same golang-migrate lineage used by cmd/migrate.
func RunMigrations(t *testing.T, dbName string, db *sql.DB) {
	t.Helper()

	driver, err := databasepostgres.WithInstance(db, &databasepostgres.Config{})
	if err != nil {
		t.Fatalf("create migration database driver: %v", err)
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration helper path")
	}
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, dbName, driver)
	if err != nil {
		t.Fatalf("initialize migrations: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("apply migrations: %v", err)
	}
}
