package container

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer holds the test container and database connection
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DB        *sql.DB
	Host      string
	Port      string
	Database  string
	Username  string
	Password  string
}

// SetupPostgresContainer creates a PostgreSQL container for testing
func SetupPostgresContainer(t *testing.T) *PostgresContainer {
	ctx := context.Background()

	// Create PostgreSQL container
	container, err := postgres.Run(ctx, "postgres:15-alpine",
		postgres.WithDatabase("auth_haven_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2). // Wait for the second occurrence (after restart)
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %s", err)
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %s", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("Failed to get container port: %s", err)
	}

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=postgres password=postgres dbname=auth_haven_test sslmode=disable",
		host, port.Port(),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %s", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %s", err)
	}

	return &PostgresContainer{
		Container: container,
		DB:        db,
		Host:      host,
		Port:      port.Port(),
		Database:  "auth_haven_test",
		Username:  "postgres",
		Password:  "postgres",
	}
}

// TeardownPostgresContainer cleans up the test container
func TeardownPostgresContainer(t *testing.T, pc *PostgresContainer) {
	ctx := context.Background()

	if pc.DB != nil {
		pc.DB.Close()
	}

	if pc.Container != nil {
		if err := pc.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %s", err)
		}
	}
}

// CleanDatabase truncates all tables for a clean test state
func CleanDatabase(db *sql.DB) error {
	tables := []string{
		"audit_logs",
		"devices",
		"password_resets",
		"invitations",
		"user_mfa_methods",
		"refresh_tokens",
		"sessions",
		"users",
		"role_permissions",
		"roles",
		"tenants",
	}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// ValidateConnection checks if the database connection is still valid
func ValidateConnection(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// GetConnectionString returns the connection string for the database
func (pc *PostgresContainer) GetConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pc.Host, pc.Port, pc.Username, pc.Password, pc.Database,
	)
}

// WaitForConnection waits for the database to be ready
func (pc *PostgresContainer) WaitForConnection(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for database connection")
		case <-ticker.C:
			if err := pc.DB.Ping(); err == nil {
				return nil
			}
		}
	}
}

// LogContainerInfo logs container information for debugging
func (pc *PostgresContainer) LogContainerInfo() {
	log.Printf("PostgreSQL Container Info:")
	log.Printf("  Host: %s", pc.Host)
	log.Printf("  Port: %s", pc.Port)
	log.Printf("  Database: %s", pc.Database)
	log.Printf("  Username: %s", pc.Username)
	log.Printf("  Connection String: %s", pc.GetConnectionString())
}
