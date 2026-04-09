package repositories

import (
	"context"
	"testing"
	"time"

	"auth-haven/internal/repository"

	"github.com/google/uuid"
)

// clearCache clears all keys from Redis cache
func clearCache(t *testing.T, cache *repository.RedisCache) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Flush all Redis data
	err := cache.DeletePrefix(ctx, "*")
	if err != nil {
		t.Logf("Warning: Failed to clear cache: %v", err)
	}
}

// assertCacheHit verifies that a value exists in cache
func assertCacheHit(t *testing.T, cache *repository.RedisCache, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result interface{}
	err := cache.Get(ctx, key, &result)
	if err != nil {
		t.Errorf("Expected cache hit for key %s, got error: %v", key, err)
	}
}

// assertCacheMiss verifies that a value does not exist in cache
func assertCacheMiss(t *testing.T, cache *repository.RedisCache, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result interface{}
	err := cache.Get(ctx, key, &result)
	if err == nil {
		t.Errorf("Expected cache miss for key %s, but got value: %v", key, result)
	}
}

// TestCachedTenantRepository_GetTenantByID_CacheHit tests cache hit scenario
func TestCachedTenantRepository_GetTenantByID_CacheHit(t *testing.T) {
	suite := SetupCacheTestSuite(t)
	defer suite.Cleanup()

	// Create base repository and cached repository
	baseRepo := repository.NewTenantRepository(suite.DB)
	cachedRepo := repository.NewCachedTenantRepository(baseRepo, suite.Cache)

	// Create test tenant
	tenantID := createTestTenant(t, suite.DB)

	// First call - should populate cache
	tenant1, err := cachedRepo.GetTenantByID(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	// Verify cache was populated
	cacheKey := "auth:ten:" + tenantID
	assertCacheHit(t, suite.Cache, cacheKey)

	// Second call - should hit cache
	tenant2, err := cachedRepo.GetTenantByID(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Failed to get tenant from cache: %v", err)
	}

	// Verify both calls returned the same data
	if tenant1.TenantID != tenant2.TenantID {
		t.Errorf("Expected same tenant ID, got %s and %s", tenant1.TenantID, tenant2.TenantID)
	}

	if tenant1.Name != tenant2.Name {
		t.Errorf("Expected same tenant name, got %s and %s", tenant1.Name, tenant2.Name)
	}
}

// TestCachedTenantRepository_GetTenantByID_CacheMiss tests cache miss scenario
func TestCachedTenantRepository_GetTenantByID_CacheMiss(t *testing.T) {
	suite := SetupCacheTestSuite(t)
	defer suite.Cleanup()

	// Create base repository and cached repository
	baseRepo := repository.NewTenantRepository(suite.DB)
	cachedRepo := repository.NewCachedTenantRepository(baseRepo, suite.Cache)

	// Clear cache
	clearCache(t, suite.Cache)

	// Call with non-existent tenant ID
	nonExistentID := uuid.New().String()
	_, err := cachedRepo.GetTenantByID(context.Background(), nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent tenant")
	}

	// Verify cache was not populated
	cacheKey := "auth:ten:" + nonExistentID
	assertCacheMiss(t, suite.Cache, cacheKey)
}

// TestCachedTenantRepository_GetTenantByDomain_CacheHit tests cache hit for domain lookup
func TestCachedTenantRepository_GetTenantByDomain_CacheHit(t *testing.T) {
	suite := SetupCacheTestSuite(t)
	defer suite.Cleanup()

	// Create base repository and cached repository
	baseRepo := repository.NewTenantRepository(suite.DB)
	cachedRepo := repository.NewCachedTenantRepository(baseRepo, suite.Cache)

	// Create test tenant with domain
	tenantID := createTestTenant(t, suite.DB)

	// Get tenant to retrieve domain
	tenant, err := baseRepo.GetTenantByID(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	// First call - should populate cache
	tenant1, err := cachedRepo.GetTenantByDomain(context.Background(), *tenant.Domain)
	if err != nil {
		t.Fatalf("Failed to get tenant by domain: %v", err)
	}

	// Verify cache was populated
	cacheKey := "auth:dom:" + *tenant.Domain
	assertCacheHit(t, suite.Cache, cacheKey)

	// Second call - should hit cache
	tenant2, err := cachedRepo.GetTenantByDomain(context.Background(), *tenant.Domain)
	if err != nil {
		t.Fatalf("Failed to get tenant by domain from cache: %v", err)
	}

	// Verify both calls returned the same data
	if tenant1.TenantID != tenant2.TenantID {
		t.Errorf("Expected same tenant ID, got %s and %s", tenant1.TenantID, tenant2.TenantID)
	}
}

// TestCachedTenantRepository_UpdateTenant_CacheInvalidation tests cache invalidation on update
func TestCachedTenantRepository_UpdateTenant_CacheInvalidation(t *testing.T) {
	suite := SetupCacheTestSuite(t)
	defer suite.Cleanup()

	// Create base repository and cached repository
	baseRepo := repository.NewTenantRepository(suite.DB)
	cachedRepo := repository.NewCachedTenantRepository(baseRepo, suite.Cache)

	// Create test tenant
	tenantID := createTestTenant(t, suite.DB)

	// Get tenant to populate cache
	_, err := cachedRepo.GetTenantByID(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	// Verify cache was populated
	cacheKey := "auth:ten:" + tenantID
	assertCacheHit(t, suite.Cache, cacheKey)

	// Update tenant
	tenant, err := baseRepo.GetTenantByID(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Failed to get tenant for update: %v", err)
	}

	tenant.Name = "Updated Organization"
	err = cachedRepo.UpdateTenant(context.Background(), tenant)
	if err != nil {
		t.Fatalf("Failed to update tenant: %v", err)
	}

	// Verify cache was invalidated
	assertCacheMiss(t, suite.Cache, cacheKey)

	// Next call should repopulate cache
	updatedTenant, err := cachedRepo.GetTenantByID(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Failed to get updated tenant: %v", err)
	}

	if updatedTenant.Name != "Updated Organization" {
		t.Errorf("Expected updated name 'Updated Organization', got '%s'", updatedTenant.Name)
	}
}
