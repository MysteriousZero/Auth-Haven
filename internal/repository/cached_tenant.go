package repository

import (
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"fmt"
	"time"
)

type cachedTenantRepository struct {
	repo  interfaces.TenantRepository
	cache *RedisCache
}

func NewCachedTenantRepository(repo interfaces.TenantRepository, cache *RedisCache) interfaces.TenantRepository {
	return &cachedTenantRepository{
		repo:  repo,
		cache: cache,
	}
}

func (r *cachedTenantRepository) GetTenantByID(ctx context.Context, tenantID string) (*models.Tenant, error) {
	key := fmt.Sprintf("auth:ten:%s", tenantID)
	var tenant models.Tenant

	if err := r.cache.Get(ctx, key, &tenant); err == nil {
		return &tenant, nil
	}

	res, err := r.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, key, res, 24*time.Hour)
	return res, nil
}

func (r *cachedTenantRepository) GetTenantByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	key := fmt.Sprintf("auth:dom:%s", domain)
	var tenant models.Tenant

	if err := r.cache.Get(ctx, key, &tenant); err == nil {
		return &tenant, nil
	}

	res, err := r.repo.GetTenantByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, key, res, 24*time.Hour)
	return res, nil
}

func (r *cachedTenantRepository) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	err := r.repo.UpdateTenant(ctx, tenant)
	if err != nil {
		return err
	}
	// Invalidate cache
	_ = r.cache.Delete(ctx, fmt.Sprintf("auth:ten:%s", tenant.TenantID))
	if tenant.Domain != nil {
		_ = r.cache.Delete(ctx, fmt.Sprintf("auth:dom:%s", *tenant.Domain))
	}
	return nil
}

// Pass-through methods
func (r *cachedTenantRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	return r.repo.CreateTenant(ctx, tenant)
}

func (r *cachedTenantRepository) CreateInvitation(ctx context.Context, invitation *models.Invitation) error {
	return r.repo.CreateInvitation(ctx, invitation)
}

func (r *cachedTenantRepository) GetInvitationByHash(ctx context.Context, tokenHash string) (*models.Invitation, error) {
	return r.repo.GetInvitationByHash(ctx, tokenHash)
}

func (r *cachedTenantRepository) GetInvitationByID(ctx context.Context, invitationID string) (*models.Invitation, error) {
	return r.repo.GetInvitationByID(ctx, invitationID)
}

func (r *cachedTenantRepository) UpdateInvitationStatus(ctx context.Context, invitationID string, status models.InvitationStatus) error {
	return r.repo.UpdateInvitationStatus(ctx, invitationID, status)
}

func (r *cachedTenantRepository) ListInvitations(ctx context.Context, tenantID string, limit, offset int) ([]models.Invitation, error) {
	return r.repo.ListInvitations(ctx, tenantID, limit, offset)
}
