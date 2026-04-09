package repository

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) interfaces.AuthRepository {
	return &authRepository{db: db}
}

// Session operations
func (r *authRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (session_id, user_id, device_id, ip_address, user_agent, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	session.CreatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query,
		session.SessionID,
		session.UserID,
		session.DeviceID,
		session.IPAddress,
		session.UserAgent,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *authRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error) {
	query := `
		SELECT session_id, user_id, device_id, ip_address, user_agent, created_at, expires_at
		FROM sessions
		WHERE session_id = $1
	`

	var session models.Session
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.DeviceID,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

func (r *authRepository) GetActiveSessions(ctx context.Context, userID string) ([]models.Session, error) {
	query := `
		SELECT session_id, user_id, device_id, ip_address, user_agent, created_at, expires_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.DeviceID,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *authRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrSessionNotFound
	}

	return nil
}

func (r *authRepository) DeleteAllUserSessions(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return nil
}

// Refresh token operations
func (r *authRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token_id, user_id, token_hash, user_agent, ip_address, revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	token.CreatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query,
		token.TokenID,
		token.UserID,
		token.TokenHash,
		token.UserAgent,
		token.IPAddress,
		token.Revoked,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *authRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT token_id, user_id, token_hash, user_agent, ip_address, revoked, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token models.RefreshToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.TokenID,
		&token.UserID,
		&token.TokenHash,
		&token.UserAgent,
		&token.IPAddress,
		&token.Revoked,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token_id = $1`

	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrInvalidToken
	}

	return nil
}

func (r *authRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

func (r *authRepository) ListSessions(ctx context.Context, userID string, limit, offset int) ([]models.Session, error) {
	query := `
		SELECT session_id, user_id, device_id, ip_address, user_agent, created_at, expires_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.DeviceID,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *authRepository) ListSessionsByDevice(ctx context.Context, deviceID string) ([]models.Session, error) {
	query := `
		SELECT session_id, user_id, device_id, ip_address, user_agent, created_at, expires_at
		FROM sessions
		WHERE device_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions by device: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.DeviceID,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *authRepository) RevokeSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}
