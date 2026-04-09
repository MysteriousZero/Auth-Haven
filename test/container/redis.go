package container

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// RedisContainer holds the test container and Redis client
type RedisContainer struct {
	Container *tcredis.RedisContainer
	Client    *redis.Client
	Host      string
	Port      string
}

// SetupRedisContainer creates a Redis container for testing
func SetupRedisContainer(t *testing.T) *RedisContainer {
	ctx := context.Background()

	// Create Redis container
	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("Failed to start Redis container: %s", err)
	}

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %s", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("Failed to get container port: %s", err)
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port.Port()),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %s", err)
	}

	return &RedisContainer{
		Container: container,
		Client:    rdb,
		Host:      host,
		Port:      port.Port(),
	}
}

// TeardownRedisContainer cleans up the test container
func TeardownRedisContainer(t *testing.T, rc *RedisContainer) {
	ctx := context.Background()

	if rc.Client != nil {
		if err := rc.Client.Close(); err != nil {
			t.Logf("Failed to close Redis client: %s", err)
		}
	}

	if rc.Container != nil {
		if err := rc.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %s", err)
		}
	}
}

// CleanRedisDatabase flushes all data for a clean test state
func CleanRedisDatabase(client *redis.Client) error {
	ctx := context.Background()
	return client.FlushDB(ctx).Err()
}

// ValidateRedisConnection checks if the Redis connection is still valid
func ValidateRedisConnection(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}

// GetRedisConnectionString returns the connection string for Redis
func (rc *RedisContainer) GetConnectionString() string {
	return fmt.Sprintf("%s:%s", rc.Host, rc.Port)
}

// WaitForRedisConnection waits for Redis to be ready
func (rc *RedisContainer) WaitForRedisConnection(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Redis connection")
		case <-ticker.C:
			if err := rc.Client.Ping(ctx).Err(); err == nil {
				return nil
			}
		}
	}
}

// LogRedisContainerInfo logs container information for debugging
func (rc *RedisContainer) LogRedisContainerInfo() {
	log.Printf("Redis Container Info:")
	log.Printf("  Host: %s", rc.Host)
	log.Printf("  Port: %s", rc.Port)
	log.Printf("  Connection String: %s", rc.GetConnectionString())
}
