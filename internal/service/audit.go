package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"fmt"
)

type auditService struct {
	auditRepo  interfaces.AuditRepository
	userRepo   interfaces.UserRepository
	tenantRepo interfaces.TenantRepository
	roleRepo   interfaces.RoleRepository
}

func NewAuditService(
	auditRepo interfaces.AuditRepository,
	userRepo interfaces.UserRepository,
	tenantRepo interfaces.TenantRepository,
	roleRepo interfaces.RoleRepository,
) interfaces.AuditService {
	return &auditService{
		auditRepo:  auditRepo,
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		roleRepo:   roleRepo,
	}
}

func (s *auditService) ListUserLogs(ctx context.Context, actorID, userID, cursor string, limit int) ([]models.AuditLog, string, error) {
	// Get actor to verify permissions
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return nil, "", errors.ErrUserNotFound
	}

	// Get target user
	targetUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", errors.ErrUserNotFound
	}

	// Verify actor has permission to view target user's logs
	// Actor can view their own logs, or logs of users in same tenant if they are admin/owner
	if actor.UserID != targetUser.UserID {
		if actor.TenantID != targetUser.TenantID {
			return nil, "", errors.ErrForbidden
		}
		if err := s.requireAuditRead(ctx, actor); err != nil {
			return nil, "", err
		}
	}

	// Get audit logs
	logs, nextCursor, err := s.auditRepo.ListUserLogs(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list user logs: %w", err)
	}

	return logs, nextCursor, nil
}

func (s *auditService) ListTenantLogs(ctx context.Context, actorID, tenantID, cursor string, limit int) ([]models.AuditLog, string, error) {
	// Get actor to verify permissions
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return nil, "", errors.ErrUserNotFound
	}

	// Get tenant to verify it exists
	_, err = s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, "", errors.ErrTenantNotFound
	}

	// Verify actor has permission to view tenant logs
	// Actor must be in the tenant and be admin or owner
	if actor.TenantID != tenantID {
		return nil, "", errors.ErrForbidden
	}

	if err := s.requireAuditRead(ctx, actor); err != nil {
		return nil, "", err
	}

	// Get audit logs
	logs, nextCursor, err := s.auditRepo.ListTenantLogs(ctx, tenantID, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list tenant logs: %w", err)
	}

	return logs, nextCursor, nil
}

func (s *auditService) requireAuditRead(ctx context.Context, actor *models.User) error {
	if actor.RoleID == nil {
		return errors.ErrForbidden
	}
	permissions, err := s.roleRepo.GetRolePermissions(ctx, *actor.RoleID)
	if err != nil {
		return errors.ErrForbidden
	}
	for _, permission := range permissions {
		if permission.PermissionKey == "audit.read" {
			return nil
		}
	}
	return errors.ErrForbidden
}
