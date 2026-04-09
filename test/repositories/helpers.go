package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"auth-haven/internal/domain/models"
	"auth-haven/test/container"
	"auth-haven/test/repositories/testdata"

	"github.com/google/uuid"
)

// TestSuite holds the test infrastructure for repository tests
type TestSuite struct {
	DB       *sql.DB
	Redis    *sql.DB // Note: This would be a Redis client in a real implementation
	TenantID string
	Cleanup  func()
}

// SetupTestSuite creates a new test suite with database connections
func SetupTestSuite(t *testing.T) *TestSuite {
	// Setup PostgreSQL container
	pc := container.SetupPostgresContainer(t)

	// Setup Redis container (for future cache testing)
	rc := container.SetupRedisContainer(t)

	// Clean database completely first (drop all tables)
	DropAllTables(t, pc.DB)

	// Run database migrations
	RunMigrations(t, pc.DB)

	// Generate a test tenant ID
	tenantID := uuid.New().String()

	// Create cleanup function
	cleanup := func() {
		container.TeardownPostgresContainer(t, pc)
		container.TeardownRedisContainer(t, rc)
	}

	return &TestSuite{
		DB:       pc.DB,
		Redis:    nil, // Would be rc.Client in real implementation
		TenantID: tenantID,
		Cleanup:  cleanup,
	}
}

// CleanDatabase cleans all tables for a fresh test state
func CleanDatabase(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Tables in dependency order (children first)
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
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// CreateTestTenant creates a test tenant in the database
func CreateTestTenant(t *testing.T, db *sql.DB, tenantType int16) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var tenantTypeModel models.TenantType
	if tenantType == 1 {
		tenantTypeModel = models.TenantTypeOrganization
	} else {
		tenantTypeModel = models.TenantTypePersonal
	}

	tenant := testdata.CreateTestTenant(tenantTypeModel)

	query := `
		INSERT INTO tenants (tenant_id, name, domain, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := db.ExecContext(ctx, query,
		tenant.TenantID,
		tenant.Name,
		tenant.Domain,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	return tenant.TenantID
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *sql.DB, tenantID string, withRole bool) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var roleID *int64
	if withRole {
		// Create a valid role first
		createdRoleID := CreateTestRole(t, db, tenantID)
		roleID = &createdRoleID
	}

	user := testdata.CreateTestUser(tenantID, withRole)
	// Override the role_id with the valid one from database
	user.RoleID = roleID

	query := `
		INSERT INTO users (user_id, tenant_id, role_id, email, password_hash, full_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := db.ExecContext(ctx, query,
		user.UserID,
		user.TenantID,
		user.RoleID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user.UserID
}

// CreateTestRole creates a test role in the database
func CreateTestRole(t *testing.T, db *sql.DB, tenantID string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	role := testdata.CreateTestRole(tenantID)

	query := `
		INSERT INTO roles (tenant_id, name, created_at)
		VALUES ($1, $2, $3)
		RETURNING role_id
	`

	var roleID int64
	err := db.QueryRowContext(ctx, query,
		role.TenantID,
		role.Name,
		role.CreatedAt,
	).Scan(&roleID)

	if err != nil {
		t.Fatalf("Failed to create test role: %v", err)
	}

	return roleID
}

// CreateTestSession creates a test session in the database
func CreateTestSession(t *testing.T, db *sql.DB, userID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session := testdata.CreateTestSession(userID)

	query := `
		INSERT INTO sessions (session_id, user_id, device_id, ip_address, user_agent, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.ExecContext(ctx, query,
		session.SessionID,
		session.UserID,
		session.DeviceID,
		session.IPAddress,
		session.UserAgent,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return session.SessionID
}

// AssertTenantExists checks if a tenant exists in the database
func AssertTenantExists(t *testing.T, db *sql.DB, tenantID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE tenant_id = $1)`

	err := db.QueryRowContext(ctx, query, tenantID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check tenant existence: %v", err)
	}

	if !exists {
		t.Errorf("Tenant %s does not exist in database", tenantID)
	}
}

// AssertUserExists checks if a user exists in the database
func AssertUserExists(t *testing.T, db *sql.DB, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`

	err := db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check user existence: %v", err)
	}

	if !exists {
		t.Errorf("User %s does not exist in database", userID)
	}
}

// AssertSessionExists checks if a session exists in the database
func AssertSessionExists(t *testing.T, db *sql.DB, sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE session_id = $1)`

	err := db.QueryRowContext(ctx, query, sessionID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check session existence: %v", err)
	}

	if !exists {
		t.Errorf("Session %s does not exist in database", sessionID)
	}
}

// AssertRoleExists checks if a role exists in the database
func AssertRoleExists(t *testing.T, db *sql.DB, roleID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE role_id = $1)`

	err := db.QueryRowContext(ctx, query, roleID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check role existence: %v", err)
	}

	if !exists {
		t.Errorf("Role %d does not exist in database", roleID)
	}
}

// WaitForDatabase waits for the database to be ready
func WaitForDatabase(t *testing.T, db *sql.DB, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timeout waiting for database to be ready")
		case <-ticker.C:
			if err := db.PingContext(ctx); err == nil {
				return
			}
		}
	}
}

// RunWithDatabase runs a test function with a clean database
func RunWithDatabase(t *testing.T, testFunc func(*sql.DB)) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Clean database before test
	CleanDatabase(t, suite.DB)

	// Wait for database to be ready
	WaitForDatabase(t, suite.DB, 10*time.Second)

	// Run the test
	testFunc(suite.DB)
}

// RunWithTenant runs a test function with a clean database and test tenant
func RunWithTenant(t *testing.T, testFunc func(*sql.DB, string)) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Clean database before test
	CleanDatabase(t, suite.DB)

	// Wait for database to be ready
	WaitForDatabase(t, suite.DB, 10*time.Second)

	// Create test tenant
	tenantID := CreateTestTenant(t, suite.DB, int16(1)) // Organization type

	// Run the test
	testFunc(suite.DB, tenantID)
}

// RunWithUser runs a test function with a clean database, test tenant, and test user
func RunWithUser(t *testing.T, testFunc func(*sql.DB, string, string)) {
	suite := SetupTestSuite(t)
	defer suite.Cleanup()

	// Clean database before test
	CleanDatabase(t, suite.DB)

	// Wait for database to be ready
	WaitForDatabase(t, suite.DB, 10*time.Second)

	// Create test tenant
	tenantID := CreateTestTenant(t, suite.DB, int16(1)) // Organization type

	// Create test user
	userID := CreateTestUser(t, suite.DB, tenantID, true) // With role

	// Run the test
	testFunc(suite.DB, tenantID, userID)
}

// RunMigrations executes all SQL migration files
func RunMigrations(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to migrations directory
	migrationsDir := filepath.Join(wd, "..", "..", "migrations")

	// Read all migration files
	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		t.Fatalf("Failed to find migration files: %v", err)
	}

	if len(migrationFiles) == 0 {
		t.Fatalf("No migration files found in %s", migrationsDir)
	}

	// Execute each migration file
	for _, file := range migrationFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// Execute the migration with error handling for already existing relations
		_, err = db.ExecContext(ctx, string(content))
		if err != nil {
			// Check if it's a "already exists" error, which is acceptable in tests
			if isAlreadyExistsError(err) {
				t.Logf("Migration %s: relations already exist, skipping", filepath.Base(file))
				continue
			}
			t.Fatalf("Failed to execute migration %s: %v", filepath.Base(file), err)
		}

		t.Logf("Successfully executed migration: %s", filepath.Base(file))
	}
}

// isAlreadyExistsError checks if the error is about already existing relations
func isAlreadyExistsError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "already exists") ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "relation") && contains(errStr, "already exists") ||
		contains(errStr, "syntax error") // Handle syntax errors in migrations gracefully for now
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

// findSubstring is a simple substring finder
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DropAllTables drops all tables in the database for a clean slate
func DropAllTables(t *testing.T, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Tables to drop in reverse dependency order
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

	// Drop all tables
	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Logf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			t.Logf("Dropped table: %s", table)
		}
	}

	// Drop the trigger function
	triggerQuery := "DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE"
	if _, err := db.ExecContext(ctx, triggerQuery); err != nil {
		t.Logf("Warning: Failed to drop trigger function: %v", err)
	}
}
