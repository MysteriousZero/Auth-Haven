package repositories

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"auth-haven/internal/repository"
	"auth-haven/test/repositories/testdata"
)

// TestTransaction_Commit tests that operations within a transaction are committed successfully.
func TestTransaction_Commit(t *testing.T) {
	RunWithTenant(t, func(db *sql.DB, tenantID string) {
		txManager := repository.NewTransactionManager(db)
		userRepo := repository.NewUserRepository(db)
		roleRepo := repository.NewRoleRepository(db)

		user := testdata.CreateTestUser(tenantID, false)
		role := testdata.CreateTestRole(tenantID)

		err := txManager.WithTransaction(context.Background(), func(txCtx context.Context) error {
			// 1. Create a user
			if err := userRepo.CreateUser(txCtx, user); err != nil {
				return err
			}

			// 2. Create a role
			if err := roleRepo.CreateRole(txCtx, role); err != nil {
				return err
			}

			// 3. Update the user with the new role
			if err := userRepo.UpdateUserRole(txCtx, user.UserID, role.RoleID); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			t.Fatalf("Transaction failed: %v", err)
		}

		// Verify data was committed
		savedUser, err := userRepo.GetUserByID(context.Background(), user.UserID)
		if err != nil {
			t.Fatalf("Failed to retrieve user after transaction: %v", err)
		}

		if savedUser.RoleID == nil || *savedUser.RoleID != role.RoleID {
			t.Errorf("Expected role ID %v, got %v", role.RoleID, savedUser.RoleID)
		}

		savedRole, err := roleRepo.GetRoleByID(context.Background(), role.RoleID)
		if err != nil {
			t.Fatalf("Failed to retrieve role after transaction: %v", err)
		}

		if savedRole.Name != role.Name {
			t.Errorf("Expected role name %s, got %s", role.Name, savedRole.Name)
		}
	})
}

// TestTransaction_Rollback tests that operations are rolled back when an error is returned.
func TestTransaction_Rollback(t *testing.T) {
	RunWithTenant(t, func(db *sql.DB, tenantID string) {
		txManager := repository.NewTransactionManager(db)
		userRepo := repository.NewUserRepository(db)

		user := testdata.CreateTestUser(tenantID, false)
		expectedErr := errors.New("simulated error to trigger rollback")

		err := txManager.WithTransaction(context.Background(), func(txCtx context.Context) error {
			// 1. Create a user
			if err := userRepo.CreateUser(txCtx, user); err != nil {
				return err
			}

			// Verify the user exists inside the transaction context
			_, err := userRepo.GetUserByID(txCtx, user.UserID)
			if err != nil {
				t.Fatalf("Failed to retrieve user inside transaction: %v", err)
			}

			// 2. Return an error to force rollback
			return expectedErr
		})

		if err == nil || err.Error() != expectedErr.Error() {
			t.Fatalf("Expected transaction to return error %v, got %v", expectedErr, err)
		}

		// Verify rollback occurred
		_, err = userRepo.GetUserByID(context.Background(), user.UserID)
		if err == nil {
			t.Errorf("Expected user to be rolled back and not found, but it was retrieved successfully")
		}
	})
}
