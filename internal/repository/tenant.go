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

type tenantRepository struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) interfaces.TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) GetTenantByID(ctx context.Context, tenantID string) (*models.Tenant, error) {
	query := `
		SELECT tenant_id, name, domain, status, created_at, updated_at
		FROM tenants
		WHERE tenant_id = $1
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.TenantID,
		&tenant.Name,
		&tenant.Domain,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

func (r *tenantRepository) GetTenantByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	query := `
		SELECT tenant_id, name, domain, status, created_at, updated_at
		FROM tenants
		WHERE domain = $1
	`

	var tenant models.Tenant
	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.TenantID,
		&tenant.Name,
		&tenant.Domain,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil for not found (auto-create scenario)
		}
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	return &tenant, nil
}

func (r *tenantRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (tenant_id, name, domain, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now().UTC()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		tenant.TenantID,
		tenant.Name,
		tenant.Domain,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation on domain
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.ErrTenantNotFound // Domain already exists
		}
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

func (r *tenantRepository) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2, domain = $3, status = $4, updated_at = $5
		WHERE tenant_id = $1
	`

	tenant.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		tenant.TenantID,
		tenant.Name,
		tenant.Domain,
		tenant.Status,
		tenant.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrTenantNotFound
	}

	return nil
}

func (r *tenantRepository) CreateInvitation(ctx context.Context, invitation *models.Invitation) error {
	query := `
		INSERT INTO invitations (invitation_id, tenant_id, role_id, email, token_hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now().UTC()
	invitation.CreatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		invitation.InvitationID,
		invitation.TenantID,
		invitation.RoleID,
		invitation.Email,
		invitation.TokenHash,
		invitation.Status,
		invitation.ExpiresAt,
		invitation.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

func (r *tenantRepository) GetInvitationByHash(ctx context.Context, tokenHash string) (*models.Invitation, error) {
	query := `
		SELECT invitation_id, tenant_id, role_id, email, token_hash, status, expires_at, created_at
		FROM invitations
		WHERE token_hash = $1
	`

	var invitation models.Invitation
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&invitation.InvitationID,
		&invitation.TenantID,
		&invitation.RoleID,
		&invitation.Email,
		&invitation.TokenHash,
		&invitation.Status,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &invitation, nil
}

func (r *tenantRepository) GetInvitationByID(ctx context.Context, invitationID string) (*models.Invitation, error) {
	query := `
		SELECT invitation_id, tenant_id, role_id, email, token_hash, status, expires_at, created_at
		FROM invitations
		WHERE invitation_id = $1
	`

	var invitation models.Invitation
	err := r.db.QueryRowContext(ctx, query, invitationID).Scan(
		&invitation.InvitationID,
		&invitation.TenantID,
		&invitation.RoleID,
		&invitation.Email,
		&invitation.TokenHash,
		&invitation.Status,
		&invitation.ExpiresAt,
		&invitation.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get invitation by ID: %w", err)
	}

	return &invitation, nil
}

func (r *tenantRepository) UpdateInvitationStatus(ctx context.Context, invitationID string, status models.InvitationStatus) error {
	query := `
		UPDATE invitations
		SET status = $2
		WHERE invitation_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, invitationID, status)
	if err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrInvitationNotFound
	}

	return nil
}

func (r *tenantRepository) ListInvitations(ctx context.Context, tenantID string, limit, offset int) ([]models.Invitation, error) {
	query := `
		SELECT invitation_id, tenant_id, role_id, email, token_hash, status, expires_at, created_at
		FROM invitations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []models.Invitation
	for rows.Next() {
		var invitation models.Invitation
		err := rows.Scan(
			&invitation.InvitationID,
			&invitation.TenantID,
			&invitation.RoleID,
			&invitation.Email,
			&invitation.TokenHash,
			&invitation.Status,
			&invitation.ExpiresAt,
			&invitation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	return invitations, nil
}
