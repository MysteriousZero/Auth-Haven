package repositories

import (
	"context"
	"database/sql"
	"testing"

	"auth-haven/internal/repository"
	"auth-haven/test/repositories/testdata"
)

// TestAuthRepository_CreateSession tests creating a new session
func TestAuthRepository_CreateSession(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test session
		session := testdata.CreateTestSession(userID)

		// Test creating session
		err := repo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Verify session was created
		AssertSessionExists(t, db, session.SessionID)
	})
}

// TestAuthRepository_GetSessionByID tests retrieving a session by ID
func TestAuthRepository_GetSessionByID(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := repo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Test getting session by ID
		retrievedSession, err := repo.GetSessionByID(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to get session by ID: %v", err)
		}

		// Verify session data
		if retrievedSession.SessionID != session.SessionID {
			t.Errorf("Expected session ID %s, got %s", session.SessionID, retrievedSession.SessionID)
		}

		if retrievedSession.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, retrievedSession.UserID)
		}

		if retrievedSession.IPAddress != session.IPAddress {
			t.Errorf("Expected IP address %s, got %s", session.IPAddress, retrievedSession.IPAddress)
		}

		if retrievedSession.UserAgent != session.UserAgent {
			t.Errorf("Expected user agent %s, got %s", session.UserAgent, retrievedSession.UserAgent)
		}
	})
}

// TestAuthRepository_ListSessions tests listing sessions for a user
func TestAuthRepository_ListSessions(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create multiple test sessions
		sessions := testdata.CreateTestSessions(3, userID)
		for _, session := range sessions {
			err := repo.CreateSession(context.Background(), session)
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}
		}

		// Test listing sessions
		retrievedSessions, err := repo.ListSessions(context.Background(), userID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}

		// Verify sessions were retrieved
		if len(retrievedSessions) < 3 {
			t.Errorf("Expected at least 3 sessions, got %d", len(retrievedSessions))
		}

		// Verify all sessions belong to the correct user
		for _, session := range retrievedSessions {
			if session.UserID != userID {
				t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
			}
		}
	})
}

// TestAuthRepository_RevokeSession tests revoking a session
func TestAuthRepository_RevokeSession(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := repo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Test revoking session
		err = repo.RevokeSession(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to revoke session: %v", err)
		}

		// Verify session was revoked by trying to get it (should be deleted)
		_, err = repo.GetSessionByID(context.Background(), session.SessionID)
		if err == nil {
			t.Error("Expected session to be deleted, but it was found")
		}
	})
}

// TestAuthRepository_DeleteSession tests deleting a session
func TestAuthRepository_DeleteSession(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test session
		session := testdata.CreateTestSession(userID)
		err := repo.CreateSession(context.Background(), session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Test deleting session
		err = repo.DeleteSession(context.Background(), session.SessionID)
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}

		// Verify session was deleted
		_, err = repo.GetSessionByID(context.Background(), session.SessionID)
		if err == nil {
			t.Error("Expected session to be deleted, but it was found")
		}
	})
}

// TestAuthRepository_DeleteAllUserSessions tests deleting all sessions for a user
func TestAuthRepository_DeleteAllUserSessions(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create multiple test sessions
		sessions := testdata.CreateTestSessions(3, userID)
		for _, session := range sessions {
			err := repo.CreateSession(context.Background(), session)
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}
		}

		// Test deleting all sessions for user
		err := repo.DeleteAllUserSessions(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to delete all user sessions: %v", err)
		}

		// Verify all sessions were deleted
		retrievedSessions, err := repo.ListSessions(context.Background(), userID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}

		if len(retrievedSessions) != 0 {
			t.Errorf("Expected 0 sessions after deletion, got %d", len(retrievedSessions))
		}
	})
}

// TestAuthRepository_CreateRefreshToken tests creating a refresh token
func TestAuthRepository_CreateRefreshToken(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test refresh token
		token := testdata.CreateTestRefreshToken(userID)

		// Test creating refresh token
		err := repo.CreateRefreshToken(context.Background(), token)
		if err != nil {
			t.Fatalf("Failed to create refresh token: %v", err)
		}

		// Verify refresh token was created by retrieving it
		retrievedToken, err := repo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
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

// TestAuthRepository_GetRefreshTokenByHash tests retrieving a refresh token by hash
func TestAuthRepository_GetRefreshTokenByHash(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test refresh token
		token := testdata.CreateTestRefreshToken(userID)
		err := repo.CreateRefreshToken(context.Background(), token)
		if err != nil {
			t.Fatalf("Failed to create refresh token: %v", err)
		}

		// Test getting refresh token by hash
		retrievedToken, err := repo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
		if err != nil {
			t.Fatalf("Failed to get refresh token by hash: %v", err)
		}

		// Verify token data
		if retrievedToken.TokenID != token.TokenID {
			t.Errorf("Expected token ID %s, got %s", token.TokenID, retrievedToken.TokenID)
		}

		if retrievedToken.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, retrievedToken.UserID)
		}

		if retrievedToken.TokenHash != token.TokenHash {
			t.Errorf("Expected token hash %s, got %s", token.TokenHash, retrievedToken.TokenHash)
		}
	})
}

// Note: GetRefreshTokensByUserID is not in the AuthRepository interface
// This functionality would be implemented in a separate repository or service

// TestAuthRepository_RevokeRefreshToken tests revoking a refresh token
func TestAuthRepository_RevokeRefreshToken(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create test refresh token
		token := testdata.CreateTestRefreshToken(userID)
		err := repo.CreateRefreshToken(context.Background(), token)
		if err != nil {
			t.Fatalf("Failed to create refresh token: %v", err)
		}

		// Test revoking refresh token
		err = repo.RevokeRefreshToken(context.Background(), token.TokenID)
		if err != nil {
			t.Fatalf("Failed to revoke refresh token: %v", err)
		}

		// Verify token was revoked
		retrievedToken, err := repo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
		if err != nil {
			t.Fatalf("Failed to get refresh token: %v", err)
		}

		if !retrievedToken.Revoked {
			t.Error("Expected token to be revoked")
		}
	})
}

// TestAuthRepository_RevokeAllUserTokens tests revoking all refresh tokens for a user
func TestAuthRepository_RevokeAllUserTokens(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewAuthRepository(db)

		// Create multiple test refresh tokens
		tokens := testdata.CreateTestRefreshTokens(3, userID)
		for _, token := range tokens {
			err := repo.CreateRefreshToken(context.Background(), token)
			if err != nil {
				t.Fatalf("Failed to create refresh token: %v", err)
			}
		}

		// Test revoking all refresh tokens for user
		err := repo.RevokeAllUserTokens(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to revoke all user tokens: %v", err)
		}

		// Note: We can't easily verify all tokens were revoked since GetRefreshTokensByUserID
		// is not in the interface, but we can verify that tokens can still be retrieved individually
		for _, token := range tokens {
			retrievedToken, err := repo.GetRefreshTokenByHash(context.Background(), token.TokenHash)
			if err != nil {
				t.Errorf("Expected token to still exist after revocation, but got error: %v", err)
			}
			if !retrievedToken.Revoked {
				t.Error("Expected token to be revoked")
			}
		}
	})
}

// Note: DeleteRefreshToken is not in the AuthRepository interface
// This functionality would be implemented in a separate repository or service

// Note: DeleteExpiredRefreshTokens is not in the AuthRepository interface
// This functionality would be implemented in a separate repository or service

// Note: Password reset operations are in the PasswordResetRepository interface, not AuthRepository
// Password reset tests should be in a separate test file
