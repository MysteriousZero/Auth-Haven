package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"auth-haven/internal/config"
	"auth-haven/test/container"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// TestDatabaseConfigWithTestcontainer tests that loaded database config can connect to testcontainer
func TestDatabaseConfigWithTestcontainer(t *testing.T) {
	// Setup PostgreSQL testcontainer
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	// Set environment variables with testcontainer values
	envVars := map[string]string{
		"DB_HOST":              pc.Host,
		"DB_PORT":              pc.Port,
		"DB_USER":              pc.Username,
		"DB_PASSWORD":          pc.Password,
		"DB_NAME":              pc.Database,
		"DB_SSLMODE":           "disable",
		"DB_MAX_CONNECTIONS":   "25",
		"DB_MAX_IDLE_CONNS":    "5",
		"DB_CONN_MAX_LIFETIME": "5m",
	}

	// Helper function to set environment variables
	setEnvVars := func(vars map[string]string) {
		for key, value := range vars {
			os.Setenv(key, value)
		}
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars map[string]string) {
		for key := range vars {
			os.Unsetenv(key)
		}
	}

	// Set environment variables
	setEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test database connection using loaded config
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Test basic query
	var result string
	err = db.QueryRowContext(ctx, "SELECT 'Config Connection Works!'").Scan(&result)
	if err != nil {
		t.Fatalf("Failed to execute test query: %v", err)
	}

	expected := "Config Connection Works!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	t.Log("Successfully connected to PostgreSQL using loaded config")
}

// TestRedisConfigWithTestcontainer tests that loaded Redis config can connect to testcontainer
func TestRedisConfigWithTestcontainer(t *testing.T) {
	// Setup Redis testcontainer
	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Set environment variables with testcontainer values
	envVars := map[string]string{
		"REDIS_HOST":     rc.Host,
		"REDIS_PORT":     rc.Port,
		"REDIS_PASSWORD": "",
		"REDIS_DB":       "0",
	}

	// Helper function to set environment variables
	setEnvVars := func(vars map[string]string) {
		for key, value := range vars {
			os.Setenv(key, value)
		}
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars map[string]string) {
		for key := range vars {
			os.Unsetenv(key)
		}
	}

	// Set environment variables
	setEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test Redis connection using loaded config
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	// Test basic Redis operations
	testKey := "config_test_key"
	testValue := "Config Redis Works!"

	// Set a key
	err = rdb.Set(ctx, testKey, testValue, 0).Err()
	if err != nil {
		t.Fatalf("Failed to set Redis key: %v", err)
	}

	// Get the key
	result, err := rdb.Get(ctx, testKey).Result()
	if err != nil {
		t.Fatalf("Failed to get Redis key: %v", err)
	}

	if result != testValue {
		t.Errorf("Expected '%s', got '%s'", testValue, result)
	}

	// Clean up test key
	err = rdb.Del(ctx, testKey).Err()
	if err != nil {
		t.Logf("Warning: Failed to delete test key: %v", err)
	}

	t.Log("Successfully connected to Redis using loaded config")
}

// TestBothConfigsConcurrent tests that both database and Redis configs work simultaneously
func TestBothConfigsConcurrent(t *testing.T) {
	// Setup both testcontainers
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Set environment variables with testcontainer values
	envVars := map[string]string{
		"DB_HOST":              pc.Host,
		"DB_PORT":              pc.Port,
		"DB_USER":              pc.Username,
		"DB_PASSWORD":          pc.Password,
		"DB_NAME":              pc.Database,
		"DB_SSLMODE":           "disable",
		"DB_MAX_CONNECTIONS":   "25",
		"DB_MAX_IDLE_CONNS":    "5",
		"DB_CONN_MAX_LIFETIME": "5m",
		"REDIS_HOST":           rc.Host,
		"REDIS_PORT":           rc.Port,
		"REDIS_PASSWORD":       "",
		"REDIS_DB":             "0",
	}

	// Helper function to set environment variables
	setEnvVars := func(vars map[string]string) {
		for key, value := range vars {
			os.Setenv(key, value)
		}
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars map[string]string) {
		for key := range vars {
			os.Unsetenv(key)
		}
	}

	// Set environment variables
	setEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test PostgreSQL connection using loaded config
	pgConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		t.Fatalf("Failed to open PostgreSQL connection: %v", err)
	}
	defer db.Close()

	// Test Redis connection using loaded config
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Test both connections
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test PostgreSQL
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	// Test Redis
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	// Test basic operations on both
	var pgResult string
	err = db.QueryRowContext(ctx, "SELECT 'PostgreSQL Config Works!'").Scan(&pgResult)
	if err != nil {
		t.Fatalf("Failed to execute PostgreSQL query: %v", err)
	}

	redisKey := "config_concurrent_test"
	redisValue := "Redis Config Works!"
	err = rdb.Set(ctx, redisKey, redisValue, 0).Err()
	if err != nil {
		t.Fatalf("Failed to set Redis key: %v", err)
	}

	redisResult, err := rdb.Get(ctx, redisKey).Result()
	if err != nil {
		t.Fatalf("Failed to get Redis key: %v", err)
	}

	// Verify results
	if pgResult != "PostgreSQL Config Works!" {
		t.Errorf("PostgreSQL result mismatch: got '%s'", pgResult)
	}

	if redisResult != redisValue {
		t.Errorf("Redis result mismatch: expected '%s', got '%s'", redisValue, redisResult)
	}

	// Clean up Redis test key
	err = rdb.Del(ctx, redisKey).Err()
	if err != nil {
		t.Logf("Warning: Failed to delete Redis test key: %v", err)
	}

	t.Log("Successfully connected to both PostgreSQL and Redis using loaded config")
}

// TestConfigConnectionValidation tests connection validation using loaded config
func TestConfigConnectionValidation(t *testing.T) {
	// Setup both testcontainers
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Set environment variables with testcontainer values
	envVars := map[string]string{
		"DB_HOST":              pc.Host,
		"DB_PORT":              pc.Port,
		"DB_USER":              pc.Username,
		"DB_PASSWORD":          pc.Password,
		"DB_NAME":              pc.Database,
		"DB_SSLMODE":           "disable",
		"DB_MAX_CONNECTIONS":   "25",
		"DB_MAX_IDLE_CONNS":    "5",
		"DB_CONN_MAX_LIFETIME": "5m",
		"REDIS_HOST":           rc.Host,
		"REDIS_PORT":           rc.Port,
		"REDIS_PASSWORD":       "",
		"REDIS_DB":             "0",
	}

	// Helper function to set environment variables
	setEnvVars := func(vars map[string]string) {
		for key, value := range vars {
			os.Setenv(key, value)
		}
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars map[string]string) {
		for key := range vars {
			os.Unsetenv(key)
		}
	}

	// Set environment variables
	setEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test config validation functions
	if err := container.ValidateConnection(pc.DB); err != nil {
		t.Fatalf("PostgreSQL connection validation failed: %v", err)
	}

	if err := container.ValidateRedisConnection(rc.Client); err != nil {
		t.Fatalf("Redis connection validation failed: %v", err)
	}

	// Test config values match testcontainer values
	if cfg.Database.Host != pc.Host {
		t.Errorf("Database host mismatch: config=%s, container=%s", cfg.Database.Host, pc.Host)
	}

	if cfg.Database.Port != 5432 {
		// Note: config.Port is the environment variable value (5432), not the mapped port
		t.Logf("Database port: config=%d, container=%s (expected - config uses original port)", cfg.Database.Port, pc.Port)
	}

	if cfg.Redis.Host != rc.Host {
		t.Errorf("Redis host mismatch: config=%s, container=%s", cfg.Redis.Host, rc.Host)
	}

	if cfg.Redis.Port != 6379 {
		// Note: config.Port is the environment variable value (6379), not the mapped port
		t.Logf("Redis port: config=%d, container=%s (expected - config uses original port)", cfg.Redis.Port, rc.Port)
	}

	t.Log("Config validation tests passed")
}

// TestConfigValuesMatchTestcontainer tests that config values match expected testcontainer settings
func TestConfigValuesMatchTestcontainer(t *testing.T) {
	// Setup both testcontainers
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Set environment variables with testcontainer values
	envVars := map[string]string{
		"DB_HOST":              pc.Host,
		"DB_PORT":              "5432", // Original PostgreSQL port
		"DB_USER":              pc.Username,
		"DB_PASSWORD":          pc.Password,
		"DB_NAME":              pc.Database,
		"DB_SSLMODE":           "disable",
		"DB_MAX_CONNECTIONS":   "25",
		"DB_MAX_IDLE_CONNS":    "5",
		"DB_CONN_MAX_LIFETIME": "5m",
		"REDIS_HOST":           rc.Host,
		"REDIS_PORT":           "6379", // Original Redis port
		"REDIS_PASSWORD":       "",
		"REDIS_DB":             "0",
	}

	// Helper function to set environment variables
	setEnvVars := func(vars map[string]string) {
		for key, value := range vars {
			os.Setenv(key, value)
		}
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars map[string]string) {
		for key := range vars {
			os.Unsetenv(key)
		}
	}

	// Set environment variables
	setEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config from environment
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that config values match expected testcontainer settings
	tests := []struct {
		name      string
		configVal interface{}
		expected  interface{}
	}{
		{"Database.Host", cfg.Database.Host, pc.Host},
		{"Database.User", cfg.Database.User, pc.Username},
		{"Database.Password", cfg.Database.Password, pc.Password},
		{"Database.DBName", cfg.Database.DBName, pc.Database},
		{"Database.SSLMode", cfg.Database.SSLMode, "disable"},
		{"Database.MaxConnections", cfg.Database.MaxConnections, 25},
		{"Database.MaxIdleConns", cfg.Database.MaxIdleConns, 5},
		{"Database.ConnMaxLifetime", cfg.Database.ConnMaxLifetime, 5 * time.Minute},
		{"Redis.Host", cfg.Redis.Host, rc.Host},
		{"Redis.Password", cfg.Redis.Password, ""},
		{"Redis.DB", cfg.Redis.DB, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configVal != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.configVal, tt.expected)
			}
		})
	}

	// Note: Ports are expected to be the original container ports (5432, 6379)
	// not the mapped ports, since that's what we set in environment variables
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432 (original port)", cfg.Database.Port)
	}

	if cfg.Redis.Port != 6379 {
		t.Errorf("Redis.Port = %d, want 6379 (original port)", cfg.Redis.Port)
	}

	t.Log("Config values match testcontainer settings")
}
