package repository

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (user_id, tenant_id, role_id, email, password_hash, full_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
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
		// Check for unique constraint violation on tenant_id + email
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.ErrEmailTaken
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT user_id, tenant_id, role_id, email, password_hash, full_name, status, created_at, updated_at, last_login_at
		FROM users
		WHERE user_id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.TenantID,
		&user.RoleID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, tenantID, email string) (*models.User, error) {
	query := `
		SELECT user_id, tenant_id, role_id, email, password_hash, full_name, status, created_at, updated_at, last_login_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&user.UserID,
		&user.TenantID,
		&user.RoleID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) ListUsers(ctx context.Context, tenantID string, limit, offset int) ([]models.User, error) {
	query := `
		SELECT user_id, tenant_id, role_id, email, password_hash, full_name, status, created_at, updated_at, last_login_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.UserID,
			&user.TenantID,
			&user.RoleID,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET full_name = $2, status = $3, updated_at = $4
		WHERE user_id = $1
	`

	user.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.FullName,
		user.Status,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	query := `
		UPDATE users
		SET status = $2, updated_at = $3
		WHERE user_id = $1
	`

	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query, userID, status, now)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_login_at = $2
		WHERE user_id = $1
	`

	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query, userID, now)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

func (r *userRepository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE user_id = $1
	`

	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query, userID, passwordHash, now)
	if err != nil {
		return fmt.Errorf("failed to update password hash: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

// Password reset methods
func (r *userRepository) CreatePasswordReset(ctx context.Context, reset *models.PasswordReset) error {
	query := `
		INSERT INTO password_resets (reset_id, user_id, token_hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query,
		reset.ResetID,
		reset.UserID,
		reset.TokenHash,
		reset.Status,
		reset.ExpiresAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	return nil
}

func (r *userRepository) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	query := `
		SELECT reset_id, user_id, token_hash, status, expires_at, created_at, updated_at
		FROM password_resets
		WHERE token_hash = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var reset models.PasswordReset
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&reset.ResetID,
		&reset.UserID,
		&reset.TokenHash,
		&reset.Status,
		&reset.ExpiresAt,
		&reset.CreatedAt,
		&reset.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPasswordResetNotFound
		}
		return nil, fmt.Errorf("failed to get password reset: %w", err)
	}

	return &reset, nil
}

func (r *userRepository) UpdatePasswordResetStatus(ctx context.Context, resetID string, status models.PasswordResetStatus) error {
	query := `
		UPDATE password_resets
		SET status = $2, updated_at = $3
		WHERE reset_id = $1
	`

	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query, resetID, status, now)
	if err != nil {
		return fmt.Errorf("failed to update password reset status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrPasswordResetNotFound
	}

	return nil
}

// Token management
func (r *userRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $2, updated_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query, userID, now)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateUserRole(ctx context.Context, userID string, roleID int64) error {
	query := `
		UPDATE users
		SET role_id = $2, updated_at = $3
		WHERE user_id = $1
	`

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query, userID, roleID, now)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `
		DELETE FROM users
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
