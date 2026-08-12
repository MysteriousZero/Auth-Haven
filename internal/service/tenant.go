package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"fmt"
	"time"
)

type tenantService struct {
	tenantRepo interfaces.TenantRepository
	userRepo   interfaces.UserRepository
}

func NewTenantService(
	tenantRepo interfaces.TenantRepository,
	userRepo interfaces.UserRepository,
) interfaces.TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

func (s *tenantService) CreateTenant(ctx context.Context, name, domain string) (*models.Tenant, error) {
	// This is a superadmin or system-level operation, not user-facing
	// In a real implementation, we'd verify the caller has system admin privileges

	// Check if domain already exists
	if domain != "" {
		existingTenant, err := s.tenantRepo.GetTenantByDomain(ctx, domain)
		if err == nil && existingTenant != nil {
			return nil, errors.ErrEmailTaken // Domain already taken
		}
	}

	// Create tenant
	tenant := &models.Tenant{
		Name:      name,
		Domain:    &domain,
		Type:      models.TenantTypeOrganization,
		Status:    models.TenantStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrTenantNotFound
	}

	return tenant, nil
}

func (s *tenantService) UpdateTenantStatus(ctx context.Context, actorID, tenantID string, status models.TenantStatus) error {
	// Get actor to verify permissions
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Get tenant to verify it exists
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return errors.ErrTenantNotFound
	}

	// In a real implementation, we'd verify actor is system admin or tenant owner
	// For now, only allow system-level operations
	if actor.TenantID != tenantID {
		// This would be a system admin check in real implementation
		return errors.ErrForbidden
	}

	// Update tenant status
	tenant.Status = status
	tenant.UpdatedAt = time.Now()

	err = s.tenantRepo.UpdateTenant(ctx, tenant)
	if err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	return nil
}
