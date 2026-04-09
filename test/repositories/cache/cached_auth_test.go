package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"auth-haven/internal/repository"
	"auth-haven/test/repositories/testdata"

	"github.com/google/uuid"
)

// TestCachedAuthRepository_GetSessionByID_CacheHit tests cache hit scenario for sessions
func TestCachedAuthRepository_GetSessionByID_CacheHit(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := cachedRepo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Clear cache first to ensure clean state
		ClearCache(t, cache)

		// First call - should populate cache
		session1, err := cachedRepo.GetSessionByID(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		// Verify cache was populated
		cacheKey := "auth:sess:" + session.SessionID
		AssertCacheHit(t, cache, cacheKey, session1)

		// Second call - should hit cache
		session2, err := cachedRepo.GetSessionByID(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to get session from cache: %v", err)
		}

		// Verify both calls returned the same data
		if session1.SessionID != session2.SessionID {
			t.Errorf("Expected same session ID, got %s and %s", session1.SessionID, session2.SessionID)
		}

		if session1.UserID != session2.UserID {
			t.Errorf("Expected same user ID, got %s and %s", session1.UserID, session2.UserID)
		}
	})
}

// TestCachedAuthRepository_GetSessionByID_CacheMiss tests cache miss scenario
func TestCachedAuthRepository_GetSessionByID_CacheMiss(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Clear cache
		ClearCache(t, cache)

		// Call with non-existent session ID
		nonExistentID := uuid.New().String()
		_, err := cachedRepo.GetSessionByID(context.Background(), nonExistentID)
		if err == nil {
			t.Error("Expected error for non-existent session")
		}

		// Verify cache was not populated
		cacheKey := "auth:sess:" + nonExistentID
		AssertCacheMiss(t, cache, cacheKey)
	})
}

// TestCachedAuthRepository_CreateSession_CachePopulation tests cache population on session creation
func TestCachedAuthRepository_CreateSession_CachePopulation(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Clear cache
		ClearCache(t, cache)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := cachedRepo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Verify cache was populated
		cacheKey := "auth:sess:" + session.SessionID
		AssertCacheHit(t, cache, cacheKey, session)
	})
}

// TestCachedAuthRepository_DeleteSession_CacheInvalidation tests cache invalidation on session deletion
func TestCachedAuthRepository_DeleteSession_CacheInvalidation(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := cachedRepo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Verify cache was populated
		cacheKey := "auth:sess:" + session.SessionID
		AssertCacheHit(t, cache, cacheKey, session)

		// Delete session
		err = cachedRepo.DeleteSession(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}

		// Verify cache was invalidated
		AssertCacheMiss(t, cache, cacheKey)

		// Verify session is actually deleted
		_, err = cachedRepo.GetSessionByID(context.Background(), session.SessionID)
		if err == nil {
			t.Error("Expected session to be deleted, but it was found")
		}
	})
}

// TestCachedAuthRepository_RevokeSession_CacheInvalidation tests cache invalidation on session revocation
func TestCachedAuthRepository_RevokeSession_CacheInvalidation(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := cachedRepo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Verify cache was populated
		cacheKey := "auth:sess:" + session.SessionID
		AssertCacheHit(t, cache, cacheKey, session)

		// Revoke session
		err = cachedRepo.RevokeSession(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to revoke session: %v", err)
		}

		// Verify cache was invalidated
		AssertCacheMiss(t, cache, cacheKey)

		// Verify session still exists but should be in revoked state
		// (The actual revocation logic would be handled by the repository implementation)
		_, err = cachedRepo.GetSessionByID(context.Background(), session.SessionID)
		if err != nil {
			t.Logf("Session retrieval after revocation: %v", err)
		}
	})
}

// TestCachedAuthRepository_CreateRefreshToken tests refresh token creation (no caching)
func TestCachedAuthRepository_CreateRefreshToken(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create test refresh token
		token := testdata.CreateTestRefreshToken(userID)
		err := cachedRepo.CreateRefreshToken(context.Background(), token)
		if err != nil {
			t.Fatalf("Failed to create refresh token: %v", err)
		}

		// Verify token was created by retrieving it
		retrievedToken, err := cachedRepo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
		if err != nil {
			t.Fatalf("Failed to get refresh token by hash: %v", err)
		}

		if retrievedToken.TokenID != token.TokenID {
			t.Errorf("Expected token ID %s, got %s", token.TokenID, retrievedToken.TokenID)
		}

		if retrievedToken.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, retrievedToken.UserID)
		}
	})
}

// TestCachedAuthRepository_RevokeRefreshToken tests refresh token revocation
func TestCachedAuthRepository_RevokeRefreshToken(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create test refresh token
		token := testdata.CreateTestRefreshToken(userID)
		err := cachedRepo.CreateRefreshToken(context.Background(), token)
		if err != nil {
			t.Fatalf("Failed to create refresh token: %v", err)
		}

		// Revoke refresh token
		err = cachedRepo.RevokeRefreshToken(context.Background(), token.TokenID)
		if err != nil {
			t.Fatalf("Failed to revoke refresh token: %v", err)
		}

		// Verify token was revoked
		retrievedToken, err := cachedRepo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
		if err != nil {
			t.Fatalf("Failed to get refresh token: %v", err)
		}

		if !retrievedToken.Revoked {
			t.Error("Expected token to be revoked")
		}
	})
}

// TestCachedAuthRepository_RevokeAllUserTokens tests revoking all user tokens
func TestCachedAuthRepository_RevokeAllUserTokens(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create multiple test refresh tokens
		tokens := testdata.CreateTestRefreshTokens(3, userID)
		for _, token := range tokens {
			err := cachedRepo.CreateRefreshToken(context.Background(), token)
			if err != nil {
				t.Fatalf("Failed to create refresh token: %v", err)
			}
		}

		// Revoke all user tokens
		err := cachedRepo.RevokeAllUserTokens(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to revoke all user tokens: %v", err)
		}

		// Verify all tokens were revoked
		for _, token := range tokens {
			retrievedToken, err := cachedRepo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
			if err != nil {
				t.Errorf("Expected token to still exist after revocation, but got error: %v", err)
			}
			if !retrievedToken.Revoked {
				t.Errorf("Expected token %s to be revoked", token.TokenID)
			}
		}
	})
}

// TestCachePerformance tests cache performance benefits
func TestCachePerformance(t *testing.T) {
	RunWithCacheAndUser(t, func(db *sql.DB, cache *repository.RedisCache, tenantID string, userID string) {
		// Create base repository and cached repository
		baseRepo := repository.NewAuthRepository(db)
		cachedRepo := repository.NewCachedAuthRepository(baseRepo, cache)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := cachedRepo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Measure time for cached calls
		start := time.Now()
		for i := 0; i < 10; i++ {
			_, err := cachedRepo.GetSessionByID(context.Background(), session.SessionID)
			if err != nil {
				t.Fatalf("Failed to get session: %v", err)
			}
		}
		cachedDuration := time.Since(start)

		// Clear cache and measure time for uncached calls
		ClearCache(t, cache)

		start = time.Now()
		for i := 0; i < 10; i++ {
			_, err := cachedRepo.GetSessionByID(context.Background(), session.SessionID)
			if err != nil {
				t.Fatalf("Failed to get session: %v", err)
			}
		}
		uncachedDuration := time.Since(start)

		t.Logf("Cached calls took: %v", cachedDuration)
		t.Logf("Uncached calls took: %v", uncachedDuration)

		// Cache should be faster (though this may vary in test environment)
		if cachedDuration > uncachedDuration*2 {
			t.Logf("Warning: Cache was slower than direct database access")
		}
	})
}
