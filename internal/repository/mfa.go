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

type mfaMethodRepository struct {
	db *sql.DB
}

func NewMFAMethodRepository(db *sql.DB) interfaces.MFAMethodRepository {
	return &mfaMethodRepository{db: db}
}

func (r *mfaMethodRepository) CreateMFAMethod(ctx context.Context, method *models.UserMFAMethod) error {
	query := `
		INSERT INTO user_mfa_methods (mfa_id, user_id, type, secret, phone_number, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	method.CreatedAt = time.Now().UTC()

	_, err := getDB(ctx, r.db).ExecContext(ctx, query,
		method.MFAID,
		method.UserID,
		method.Type,
		method.Secret,
		method.PhoneNumber,
		method.Enabled,
		method.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create MFA method: %w", err)
	}

	return nil
}

func (r *mfaMethodRepository) ListMFAMethods(ctx context.Context, userID string) ([]models.UserMFAMethod, error) {
	query := `
		SELECT mfa_id, user_id, type, secret, phone_number, enabled, created_at
		FROM user_mfa_methods
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := getDB(ctx, r.db).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list MFA methods: %w", err)
	}
	defer rows.Close()

	var methods []models.UserMFAMethod
	for rows.Next() {
		var method models.UserMFAMethod
		err := rows.Scan(
			&method.MFAID,
			&method.UserID,
			&method.Type,
			&method.Secret,
			&method.PhoneNumber,
			&method.Enabled,
			&method.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan MFA method: %w", err)
		}
		methods = append(methods, method)
	}

	return methods, nil
}

func (r *mfaMethodRepository) UpdateMFAMethod(ctx context.Context, mfaID string, enabled bool) error {
	query := `UPDATE user_mfa_methods SET enabled = $2 WHERE mfa_id = $1`

	result, err := getDB(ctx, r.db).ExecContext(ctx, query, mfaID, enabled)
	if err != nil {
		return fmt.Errorf("failed to update MFA method: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrMFAMethodNotFound
	}

	return nil
}

func (r *mfaMethodRepository) DeleteMFAMethod(ctx context.Context, mfaID string) error {
	query := `DELETE FROM user_mfa_methods WHERE mfa_id = $1`

	result, err := getDB(ctx, r.db).ExecContext(ctx, query, mfaID)
	if err != nil {
		return fmt.Errorf("failed to delete MFA method: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrMFAMethodNotFound
	}

	return nil
}

type passwordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) interfaces.PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) CreatePasswordReset(ctx context.Context, reset *models.PasswordReset) error {
	query := `
		INSERT INTO password_resets (reset_id, user_id, token_hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	reset.CreatedAt = time.Now().UTC()

	_, err := getDB(ctx, r.db).ExecContext(ctx, query,
		reset.ResetID,
		reset.UserID,
		reset.TokenHash,
		reset.Status,
		reset.ExpiresAt,
		reset.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	return nil
}

func (r *passwordResetRepository) GetPasswordResetByHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	query := `
		SELECT reset_id, user_id, token_hash, status, expires_at, created_at
		FROM password_resets
		WHERE token_hash = $1
	`

	var reset models.PasswordReset
	err := getDB(ctx, r.db).QueryRowContext(ctx, query, tokenHash).Scan(
		&reset.ResetID,
		&reset.UserID,
		&reset.TokenHash,
		&reset.Status,
		&reset.ExpiresAt,
		&reset.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get password reset: %w", err)
	}

	return &reset, nil
}

func (r *passwordResetRepository) UpdatePasswordResetStatus(ctx context.Context, resetID string, status models.PasswordResetStatus) error {
	query := `UPDATE password_resets SET status = $2 WHERE reset_id = $1`

	result, err := getDB(ctx, r.db).ExecContext(ctx, query, resetID, status)
	if err != nil {
		return fmt.Errorf("failed to update password reset status: %w", err)
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

type auditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) interfaces.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (log_id, user_id, tenant_id, action, target_id, metadata, ip_address, user_agent, trace_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	log.CreatedAt = time.Now().UTC()

	// Convert metadata to JSONB
	var metadata interface{} = nil
	if log.Metadata != nil {
		metadata = log.Metadata
	}

	_, err := getDB(ctx, r.db).ExecContext(ctx, query,
		log.LogID,
		log.UserID,
		log.TenantID,
		log.Action,
		log.TargetID,
		metadata,
		log.IPAddress,
		log.UserAgent,
		log.TraceID,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (r *auditRepository) ListTenantLogs(ctx context.Context, tenantID string, cursor string, limit int) ([]models.AuditLog, string, error) {
	var rows *sql.Rows
	var err error
	var query string

	if cursor == "" {
		query = `
			SELECT log_id, user_id, tenant_id, action, target_id, metadata, ip_address, user_agent, trace_id, created_at
			FROM audit_logs
			WHERE tenant_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		rows, err = getDB(ctx, r.db).QueryContext(ctx, query, tenantID, limit)
	} else {
		query = `
			SELECT log_id, user_id, tenant_id, action, target_id, metadata, ip_address, user_agent, trace_id, created_at
			FROM audit_logs
			WHERE tenant_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		rows, err = getDB(ctx, r.db).QueryContext(ctx, query, tenantID, cursor, limit)
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to list tenant logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	var lastCreatedAt time.Time

	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.LogID,
			&log.UserID,
			&log.TenantID,
			&log.Action,
			&log.TargetID,
			&log.Metadata,
			&log.IPAddress,
			&log.UserAgent,
			&log.TraceID,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
		lastCreatedAt = log.CreatedAt
	}

	var nextCursor string
	if len(logs) == limit {
		nextCursor = lastCreatedAt.Format(time.RFC3339Nano)
	}

	return logs, nextCursor, nil
}

func (r *auditRepository) ListUserLogs(ctx context.Context, userID string, cursor string, limit int) ([]models.AuditLog, string, error) {
	var rows *sql.Rows
	var err error
	var query string

	if cursor == "" {
		query = `
			SELECT log_id, user_id, tenant_id, action, target_id, metadata, ip_address, user_agent, trace_id, created_at
			FROM audit_logs
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		rows, err = getDB(ctx, r.db).QueryContext(ctx, query, userID, limit)
	} else {
		query = `
			SELECT log_id, user_id, tenant_id, action, target_id, metadata, ip_address, user_agent, trace_id, created_at
			FROM audit_logs
			WHERE user_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		rows, err = getDB(ctx, r.db).QueryContext(ctx, query, userID, cursor, limit)
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to list user logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	var lastCreatedAt time.Time

	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.LogID,
			&log.UserID,
			&log.TenantID,
			&log.Action,
			&log.TargetID,
			&log.Metadata,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
		lastCreatedAt = log.CreatedAt
	}

	var nextCursor string
	if len(logs) == limit {
		nextCursor = lastCreatedAt.Format(time.RFC3339Nano)
	}

	return logs, nextCursor, nil
}
