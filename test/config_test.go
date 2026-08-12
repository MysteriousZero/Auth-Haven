package test

import (
	"os"
	"testing"
	"time"

	"auth-haven/internal/config"
)

// TestLoadDatabaseConfigFromEnv tests loading database configuration from environment variables
func TestLoadDatabaseConfigFromEnv(t *testing.T) {
	// Set up environment variables
	envVars := map[string]string{
		"DB_HOST":              "localhost",
		"DB_PORT":              "5432",
		"DB_USER":              "postgres",
		"DB_PASSWORD":          "postgres",
		"DB_NAME":              "auth_haven_test",
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

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test DatabaseConfig values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", cfg.Database.Host, "localhost"},
		{"Port", cfg.Database.Port, 5432},
		{"User", cfg.Database.User, "postgres"},
		{"Password", cfg.Database.Password, "postgres"},
		{"DBName", cfg.Database.DBName, "auth_haven_test"},
		{"SSLMode", cfg.Database.SSLMode, "disable"},
		{"MaxConnections", cfg.Database.MaxConnections, 25},
		{"MaxIdleConns", cfg.Database.MaxIdleConns, 5},
		{"ConnMaxLifetime", cfg.Database.ConnMaxLifetime, 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Database.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestLoadRedisConfigFromEnv tests loading Redis configuration from environment variables
func TestLoadRedisConfigFromEnv(t *testing.T) {
	// Set up environment variables
	envVars := map[string]string{
		"REDIS_HOST":     "localhost",
		"REDIS_PORT":     "6379",
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

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test RedisConfig values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", cfg.Redis.Host, "localhost"},
		{"Port", cfg.Redis.Port, 6379},
		{"Password", cfg.Redis.Password, ""},
		{"DB", cfg.Redis.DB, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Redis.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestConfigFallbackValues tests that fallback values are used when environment variables are not set
func TestConfigFallbackValues(t *testing.T) {
	// Clean environment variables
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_CONNECTIONS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars []string) {
		for _, key := range vars {
			os.Unsetenv(key)
		}
	}

	// Clean environment variables
	cleanupEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config (should use fallback values)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test fallback values for DatabaseConfig
	dbTests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", cfg.Database.Host, "localhost"},
		{"Port", cfg.Database.Port, 5432},
		{"User", cfg.Database.User, "postgres"},
		{"Password", cfg.Database.Password, ""},
		{"DBName", cfg.Database.DBName, "auth_haven"},
		{"SSLMode", cfg.Database.SSLMode, "disable"},
		{"MaxConnections", cfg.Database.MaxConnections, 25},
		{"MaxIdleConns", cfg.Database.MaxIdleConns, 5},
		{"ConnMaxLifetime", cfg.Database.ConnMaxLifetime, 5 * time.Minute},
	}

	for _, tt := range dbTests {
		t.Run("DB_"+tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Database.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	// Test fallback values for RedisConfig
	redisTests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", cfg.Redis.Host, "localhost"},
		{"Port", cfg.Redis.Port, 6379},
		{"Password", cfg.Redis.Password, ""},
		{"DB", cfg.Redis.DB, 0},
	}

	for _, tt := range redisTests {
		t.Run("Redis_"+tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Redis.%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestInvalidEnvironmentValues tests handling of invalid environment variable values
func TestInvalidEnvironmentValues(t *testing.T) {
	// Set up invalid environment variables
	envVars := map[string]string{
		"DB_PORT":              "invalid_port",
		"DB_MAX_CONNECTIONS":   "invalid_connections",
		"DB_MAX_IDLE_CONNS":    "invalid_idle",
		"DB_CONN_MAX_LIFETIME": "invalid_duration",
		"REDIS_PORT":           "invalid_redis_port",
		"REDIS_DB":             "invalid_redis_db",
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

	// Load config (should use fallback values for invalid entries)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that fallback values are used for invalid entries
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"DB_Port", cfg.Database.Port, 5432},                                  // fallback
		{"DB_MaxConnections", cfg.Database.MaxConnections, 25},                // fallback
		{"DB_MaxIdleConns", cfg.Database.MaxIdleConns, 5},                     // fallback
		{"DB_ConnMaxLifetime", cfg.Database.ConnMaxLifetime, 5 * time.Minute}, // fallback
		{"Redis_Port", cfg.Redis.Port, 6379},                                  // fallback
		{"Redis_DB", cfg.Redis.DB, 0},                                         // fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestConfigLoadFunction tests the basic config.Load() function
func TestConfigLoadFunction(t *testing.T) {
	// Clean environment to test basic functionality
	envVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_CONNECTIONS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
	}

	// Helper function to clean environment variables
	cleanupEnvVars := func(vars []string) {
		for _, key := range vars {
			os.Unsetenv(key)
		}
	}

	// Clean environment variables
	cleanupEnvVars(envVars)
	defer cleanupEnvVars(envVars)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() failed: %v", err)
	}

	// Test that config is not nil
	if cfg == nil {
		t.Fatal("config.Load() returned nil config")
	}

	// Test that all sections are initialized
	if cfg.Database.Host == "" {
		t.Error("Database.Host is empty")
	}
	if cfg.Redis.Host == "" {
		t.Error("Redis.Host is empty")
	}
	if cfg.Server.HTTPPort == "" {
		t.Error("Server.HTTPPort is empty")
	}
	if cfg.Auth.JWTPrivateKey == "" {
		// This is expected to be empty by default
		t.Log("Auth.JWTPrivateKey is empty (expected)")
	}
}
