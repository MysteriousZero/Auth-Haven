package main

import (
	"auth-haven/internal/config"
	appmigration "auth-haven/internal/migration"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Build connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open database for migration preflight: %v", err)
	}
	defer db.Close()
	if err := appmigration.ValidateState(context.Background(), db); err != nil {
		log.Fatalf("migration preflight failed: %v", err)
	}

	m, err := migrate.New("file://migrations", connStr)
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("✅ database migrated successfully")
}
