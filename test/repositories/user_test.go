package repositories

import (
	"context"
	"database/sql"
	"testing"

	"auth-haven/internal/domain/models"
	"auth-haven/internal/repository"
	"auth-haven/test/repositories/testdata"
)

// TestUserRepository_CreateUser tests creating a new user
func TestUserRepository_CreateUser(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test getting the user that was created by the helper
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		// Verify user data
		if user.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, user.UserID)
		}

		if user.TenantID != tenantID {
			t.Errorf("Expected tenant ID %s, got %s", tenantID, user.TenantID)
		}

		if user.Email != "test-"+userID[:8]+"@example.com" {
			t.Errorf("Expected email pattern, got %s", user.Email)
		}

		if user.FullName != "Test User" {
			t.Errorf("Expected full name 'Test User', got '%s'", user.FullName)
		}

		if user.Status != models.UserStatusActive {
			t.Errorf("Expected status %d, got %d", models.UserStatusActive, user.Status)
		}
	})
}

// TestUserRepository_GetUserByID tests retrieving a user by ID
func TestUserRepository_GetUserByID(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test getting user by ID
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		// Verify user data
		if user.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, user.UserID)
		}

		if user.TenantID != tenantID {
			t.Errorf("Expected tenant ID %s, got %s", tenantID, user.TenantID)
		}

		if user.RoleID == nil {
			t.Error("Expected user to have a role")
		}
	})
}

// TestUserRepository_GetUserByEmail tests retrieving a user by email
func TestUserRepository_GetUserByEmail(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Get the user to retrieve their email
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		// Test getting user by email
		foundUser, err := repo.GetUserByEmail(context.Background(), tenantID, user.Email)
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}

		// Verify user data
		if foundUser.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, foundUser.UserID)
		}

		if foundUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
		}

		if foundUser.TenantID != tenantID {
			t.Errorf("Expected tenant ID %s, got %s", tenantID, foundUser.TenantID)
		}
	})
}

// TestUserRepository_ListUsers tests listing users for a tenant
func TestUserRepository_ListUsers(t *testing.T) {
	RunWithTenant(t, func(db *sql.DB, tenantID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Create multiple test users
		userIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			userID := CreateTestUser(t, db, tenantID, i%2 == 0) // Alternate with/without role
			userIDs[i] = userID
		}

		// Test listing users
		users, err := repo.ListUsers(context.Background(), tenantID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}

		// Verify users were listed
		if len(users) < 3 {
			t.Errorf("Expected at least 3 users, got %d", len(users))
		}

		// Verify all users belong to the correct tenant
		for _, user := range users {
			if user.TenantID != tenantID {
				t.Errorf("Expected tenant ID %s, got %s", tenantID, user.TenantID)
			}
		}
	})
}

// TestUserRepository_UpdateUser tests updating a user
func TestUserRepository_UpdateUser(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Get the user to update
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		// Update user data
		user.FullName = "Updated User Name"
		user.Status = models.UserStatusDisabled

		// Test updating user
		err = repo.UpdateUser(context.Background(), user)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Verify user was updated
		updatedUser, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		if updatedUser.FullName != "Updated User Name" {
			t.Errorf("Expected updated full name 'Updated User Name', got '%s'", updatedUser.FullName)
		}

		if updatedUser.Status != models.UserStatusDisabled {
			t.Errorf("Expected updated status %d, got %d", models.UserStatusDisabled, updatedUser.Status)
		}
	})
}

// TestUserRepository_UpdateUserStatus tests updating user status
func TestUserRepository_UpdateUserStatus(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test updating user status
		err := repo.UpdateUserStatus(context.Background(), userID, models.UserStatusPending)
		if err != nil {
			t.Fatalf("Failed to update user status: %v", err)
		}

		// Verify user status was updated
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.Status != models.UserStatusPending {
			t.Errorf("Expected status %d, got %d", models.UserStatusPending, user.Status)
		}
	})
}

// TestUserRepository_UpdateUserRole tests updating user role
func TestUserRepository_UpdateUserRole(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Create a new role for the user
		newRoleID := CreateTestRole(t, db, tenantID)

		// Test updating user role
		err := repo.UpdateUserRole(context.Background(), userID, newRoleID)
		if err != nil {
			t.Fatalf("Failed to update user role: %v", err)
		}

		// Verify user role was updated
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.RoleID == nil || *user.RoleID != newRoleID {
			t.Errorf("Expected role ID %d, got %v", newRoleID, user.RoleID)
		}
	})
}

// TestUserRepository_UpdatePasswordHash tests updating user password hash
func TestUserRepository_UpdatePasswordHash(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test updating user password hash
		newPasswordHash := "$2a$12$newhashedpasswordplaceholder"
		err := repo.UpdatePasswordHash(context.Background(), userID, newPasswordHash)
		if err != nil {
			t.Fatalf("Failed to update user password hash: %v", err)
		}

		// Verify user password hash was updated
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.PasswordHash != newPasswordHash {
			t.Errorf("Expected password hash %s, got %s", newPasswordHash, user.PasswordHash)
		}
	})
}

// TestUserRepository_UpdateLastLogin tests updating user last login
func TestUserRepository_UpdateLastLogin(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test updating user last login
		err := repo.UpdateLastLogin(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to update user last login: %v", err)
		}

		// Verify user last login was updated
		user, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if user.LastLoginAt == nil {
			t.Error("Expected last login at to be set")
		}
	})
}

// TestUserRepository_DeleteUser tests deleting a user
func TestUserRepository_DeleteUser(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test deleting user
		err := repo.DeleteUser(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify user was deleted
		_, err = repo.GetUserByID(context.Background(), userID)
		if err == nil {
			t.Error("Expected user to be deleted, but it was found")
		}
	})
}

// TestUserRepository_CreateUserWithoutRole tests creating a user without a role
func TestUserRepository_CreateUserWithoutRole(t *testing.T) {
	RunWithTenant(t, func(db *sql.DB, tenantID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Create test user without role
		user := testdata.CreateTestUser(tenantID, false) // No role

		// Test creating user
		err := repo.CreateUser(context.Background(), user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Verify user was created
		retrievedUser, err := repo.GetUserByID(context.Background(), user.UserID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if retrievedUser.RoleID != nil {
			t.Error("Expected user to have no role")
		}

		// Verify user methods work
		if !retrievedUser.IsPersonalUser() {
			t.Error("Personal user should return true for IsPersonalUser()")
		}

		if retrievedUser.HasRole() {
			t.Error("Personal user should return false for HasRole()")
		}
	})
}

// TestUserRepository_CreateUserDuplicateEmail tests creating a user with duplicate email
func TestUserRepository_CreateUserDuplicateEmail(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Get the existing user to get their email
		existingUser, err := repo.GetUserByID(context.Background(), userID)
		if err != nil {
			t.Fatalf("Failed to get existing user: %v", err)
		}

		// Create a new user with the same email
		duplicateUser := testdata.CreateTestUser(tenantID, true)
		duplicateUser.Email = existingUser.Email // Use same email

		// Test creating user with duplicate email
		err = repo.CreateUser(context.Background(), duplicateUser)
		if err == nil {
			t.Error("Expected error when creating user with duplicate email")
		}

		t.Logf("Expected error for duplicate email: %v", err)
	})
}

// TestUserRepository_UserStatusMethods tests user status methods
func TestUserRepository_UserStatusMethods(t *testing.T) {
	RunWithUser(t, func(db *sql.DB, tenantID string, userID string) {
		// Create repository
		repo := repository.NewUserRepository(db)

		// Test different user statuses
		statuses := []models.UserStatus{
			models.UserStatusActive,
			models.UserStatusPending,
			models.UserStatusDisabled,
		}

		for i, status := range statuses {
			// Update user status
			err := repo.UpdateUserStatus(context.Background(), userID, status)
			if err != nil {
				t.Fatalf("Failed to update user status: %v", err)
			}

			// Get user and verify status methods
			user, err := repo.GetUserByID(context.Background(), userID)
			if err != nil {
				t.Fatalf("Failed to get user: %v", err)
			}

			// Verify status methods work correctly
			switch status {
			case models.UserStatusActive:
				if !user.IsActive() {
					t.Errorf("User should be active (test %d)", i)
				}
				if user.IsPending() || user.IsDisabled() {
					t.Errorf("Active user should not be pending or disabled (test %d)", i)
				}
			case models.UserStatusPending:
				if !user.IsPending() {
					t.Errorf("User should be pending (test %d)", i)
				}
				if user.IsActive() || user.IsDisabled() {
					t.Errorf("Pending user should not be active or disabled (test %d)", i)
				}
			case models.UserStatusDisabled:
				if !user.IsDisabled() {
					t.Errorf("User should be disabled (test %d)", i)
				}
				if user.IsActive() || user.IsPending() {
					t.Errorf("Disabled user should not be active or pending (test %d)", i)
				}
			}
		}
	})
}
