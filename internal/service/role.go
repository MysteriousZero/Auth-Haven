package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type roleService struct {
	roleRepo   interfaces.RoleRepository
	userRepo   interfaces.UserRepository
	tenantRepo interfaces.TenantRepository
	auditRepo  interfaces.AuditRepository
}

func NewRoleService(
	roleRepo interfaces.RoleRepository,
	userRepo interfaces.UserRepository,
	tenantRepo interfaces.TenantRepository,
	auditRepo interfaces.AuditRepository,
) interfaces.RoleService {
	return &roleService{
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		auditRepo:  auditRepo,
	}
}

func (s *roleService) CreateRole(ctx context.Context, actorID, tenantID, name string) (*models.Role, error) {
	// Get actor to verify they are owner
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if actor.TenantID != tenantID {
		return nil, errors.ErrForbidden
	}

	// Verify actor is owner
	if err := s.requireOwner(ctx, actor); err != nil {
		return nil, err
	}

	// Get tenant to verify it's an org tenant
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrTenantNotFound
	}
	if tenant.Type != models.TenantTypeOrganization {
		return nil, errors.ErrForbidden
	}

	// Create role
	role := &models.Role{
		TenantID:  tenantID,
		Name:      name,
		CreatedAt: time.Now(),
	}

	err = s.roleRepo.CreateRole(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Emit audit log
	roleIDStr := fmt.Sprintf("%d", role.RoleID)
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  tenantID,
		Action:    models.AuditRoleCreated,
		TargetID:  &roleIDStr,
		Metadata:  map[string]interface{}{"role_name": name},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return role, nil
}

func (s *roleService) DeleteRole(ctx context.Context, actorID string, roleID int64) error {
	// Get actor to verify they are owner
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Get role to verify it exists and belongs to actor's tenant
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}
	if role.TenantID != actor.TenantID {
		return errors.ErrRoleNotFound
	}

	// Verify actor is owner
	if err := s.requireOwner(ctx, actor); err != nil {
		return err
	}

	// Check if any users are assigned to this role
	users, err := s.userRepo.ListUsers(ctx, actor.TenantID, 1000, 0)
	if err == nil {
		for _, user := range users {
			if user.RoleID != nil && *user.RoleID == roleID {
				return errors.ErrRoleInUse
			}
		}
	}

	// Delete role permissions first
	err = s.roleRepo.DeleteRolePermissions(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	// Delete role
	err = s.roleRepo.DeleteRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// Emit audit log
	roleIDStr := fmt.Sprintf("%d", roleID)
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  actor.TenantID,
		Action:    models.AuditRoleDeleted,
		TargetID:  &roleIDStr,
		Metadata:  map[string]interface{}{"role_name": role.Name},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *roleService) ListRoles(ctx context.Context, tenantID string, limit, offset int) ([]models.Role, error) {
	// Verify tenant exists
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrTenantNotFound
	}

	// Don't allow listing roles for personal tenants
	if tenant.Type == models.TenantTypePersonal {
		return nil, errors.ErrForbidden
	}

	return s.roleRepo.ListRoles(ctx, tenantID, limit, offset)
}

func (s *roleService) GetRolePermissions(ctx context.Context, roleID int64) ([]models.RolePermission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

func (s *roleService) SetRolePermissions(ctx context.Context, actorID string, roleID int64, keys []string) error {
	// Get actor to verify they are owner
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Get role to verify it belongs to actor's tenant
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return errors.ErrRoleNotFound
	}
	if role.TenantID != actor.TenantID {
		return errors.ErrRoleNotFound
	}

	// Verify actor is owner
	if err := s.requireOwner(ctx, actor); err != nil {
		return err
	}

	// Delete existing permissions
	err = s.roleRepo.DeleteRolePermissions(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// Add new permissions
	for _, key := range keys {
		err = s.roleRepo.AddRolePermission(ctx, roleID, key)
		if err != nil {
			return fmt.Errorf("failed to add permission %s: %w", key, err)
		}
	}

	// Emit audit log
	roleIDStr := fmt.Sprintf("%d", roleID)
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  actor.TenantID,
		Action:    models.AuditRolePermissionsUpdated,
		TargetID:  &roleIDStr,
		Metadata:  map[string]interface{}{"permissions": keys},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

// requireOwner checks that the actor's role is named "Owner" (owns the tenant).
func (s *roleService) requireOwner(ctx context.Context, actor *models.User) error {
	if actor.RoleID == nil {
		return errors.ErrForbidden
	}
	role, err := s.roleRepo.GetRoleByID(ctx, *actor.RoleID)
	if err != nil {
		return errors.ErrForbidden
	}
	if role.Name != "Owner" {
		return errors.ErrForbidden
	}
	return nil
}

