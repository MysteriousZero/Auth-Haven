package test

import (
	"context"
	"fmt"
	"testing"

	"auth-haven/test/container"
)

// TestPostgresContainerConnection tests PostgreSQL container setup and connection
func TestPostgresContainerConnection(t *testing.T) {
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	t.Log("Successfully connected to PostgreSQL using testcontainers")

	// Test basic query
	var result string
	err := pc.DB.QueryRow("SELECT 'Hello, PostgreSQL!'").Scan(&result)
	if err != nil {
		t.Fatalf("Failed to execute test query: %v", err)
	}

	expected := "Hello, PostgreSQL!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	t.Log("Successfully executed PostgreSQL test query")
}

// TestRedisContainerConnection tests Redis container setup and connection
func TestRedisContainerConnection(t *testing.T) {
	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	t.Log("Successfully connected to Redis using testcontainers")

	// Test basic Redis operations
	ctx := context.Background()

	// Set a key
	err := rc.Client.Set(ctx, "test_key", "Hello, Redis!", 0).Err()
	if err != nil {
		t.Fatalf("Failed to set Redis key: %v", err)
	}

	// Get the key
	result, err := rc.Client.Get(ctx, "test_key").Result()
	if err != nil {
		t.Fatalf("Failed to get Redis key: %v", err)
	}

	expected := "Hello, Redis!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Clean up test key
	err = rc.Client.Del(ctx, "test_key").Err()
	if err != nil {
		t.Logf("Warning: Failed to delete test key: %v", err)
	}

	t.Log("Successfully executed Redis test operations")
}

// TestBothContainersConnection tests both PostgreSQL and Redis containers together
func TestBothContainersConnection(t *testing.T) {
	// Setup PostgreSQL container
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	// Setup Redis container
	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	t.Log("Successfully connected to both PostgreSQL and Redis using testcontainers")

	// Test PostgreSQL connection
	var pgResult string
	err := pc.DB.QueryRow("SELECT 'PostgreSQL OK'").Scan(&pgResult)
	if err != nil {
		t.Fatalf("Failed to execute PostgreSQL test query: %v", err)
	}

	// Test Redis connection
	ctx := context.Background()
	err = rc.Client.Set(ctx, "container_test", "Redis OK", 0).Err()
	if err != nil {
		t.Fatalf("Failed to set Redis key: %v", err)
	}

	redisResult, err := rc.Client.Get(ctx, "container_test").Result()
	if err != nil {
		t.Fatalf("Failed to get Redis key: %v", err)
	}

	// Verify results
	if pgResult != "PostgreSQL OK" {
		t.Errorf("Expected 'PostgreSQL OK', got '%s'", pgResult)
	}

	if redisResult != "Redis OK" {
		t.Errorf("Expected 'Redis OK', got '%s'", redisResult)
	}

	// Clean up Redis test key
	err = rc.Client.Del(ctx, "container_test").Err()
	if err != nil {
		t.Logf("Warning: Failed to delete test key: %v", err)
	}

	t.Log("Both database connections are working correctly")
}

// TestContainerConnectionValidation tests connection validation functions
func TestContainerConnectionValidation(t *testing.T) {
	// Setup PostgreSQL container
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	// Setup Redis container
	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Test PostgreSQL connection validation
	err := container.ValidateConnection(pc.DB)
	if err != nil {
		t.Fatalf("PostgreSQL connection validation failed: %v", err)
	}

	// Test Redis connection validation
	err = container.ValidateRedisConnection(rc.Client)
	if err != nil {
		t.Fatalf("Redis connection validation failed: %v", err)
	}

	t.Log("Both connection validation functions work correctly")
}

// TestContainerInfoLogging tests container info logging functions
func TestContainerInfoLogging(t *testing.T) {
	// Setup PostgreSQL container
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	// Setup Redis container
	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Test PostgreSQL container info logging
	pc.LogContainerInfo()

	// Test Redis container info logging
	rc.LogRedisContainerInfo()

	t.Log("Container info logging functions executed successfully")
}

// TestGetConnectionStrings tests connection string generation
func TestGetConnectionStrings(t *testing.T) {
	// Setup PostgreSQL container
	pc := container.SetupPostgresContainer(t)
	defer container.TeardownPostgresContainer(t, pc)

	// Setup Redis container
	rc := container.SetupRedisContainer(t)
	defer container.TeardownRedisContainer(t, rc)

	// Test PostgreSQL connection string
	pgConnStr := pc.GetConnectionString()
	if pgConnStr == "" {
		t.Error("PostgreSQL connection string is empty")
	}

	expectedPGFormat := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=auth_haven_test sslmode=disable",
		pc.Host, pc.Port)
	if pgConnStr != expectedPGFormat {
		t.Errorf("PostgreSQL connection string format mismatch. Expected: %s, Got: %s", expectedPGFormat, pgConnStr)
	}

	// Test Redis connection string
	redisConnStr := rc.GetConnectionString()
	if redisConnStr == "" {
		t.Error("Redis connection string is empty")
	}

	expectedRedisFormat := fmt.Sprintf("%s:%s", rc.Host, rc.Port)
	if redisConnStr != expectedRedisFormat {
		t.Errorf("Redis connection string format mismatch. Expected: %s, Got: %s", expectedRedisFormat, redisConnStr)
	}

	t.Log("Connection string generation works correctly")
}
