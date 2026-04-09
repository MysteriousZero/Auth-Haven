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

type invitationService struct {
	tenantRepo    interfaces.TenantRepository
	userRepo      interfaces.UserRepository
	roleRepo      interfaces.RoleRepository
	auditRepo     interfaces.AuditRepository
	emailProvider interfaces.EmailProvider
	tokenService  interfaces.TokenService
	hasher        interfaces.Hasher
}

func NewInvitationService(
	tenantRepo interfaces.TenantRepository,
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	auditRepo interfaces.AuditRepository,
	emailProvider interfaces.EmailProvider,
	tokenService interfaces.TokenService,
	hasher interfaces.Hasher,
) interfaces.InvitationService {
	return &invitationService{
		tenantRepo:    tenantRepo,
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		auditRepo:     auditRepo,
		emailProvider: emailProvider,
		tokenService:  tokenService,
		hasher:        hasher,
	}
}

func (s *invitationService) SendInvitation(ctx context.Context, actorID, tenantID, email string, roleID int64) (*models.Invitation, error) {
	// Get tenant to verify it's an org tenant
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrTenantNotFound
	}
	if tenant.Type != models.TenantTypeOrganization {
		return nil, errors.ErrForbidden
	}

	// Get actor to verify permissions
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	if actor.TenantID != tenantID {
		return nil, errors.ErrForbidden
	}

	// Check actor is owner or admin via role permissions
	if err := s.requireOwnerOrAdmin(ctx, actor); err != nil {
		return nil, err
	}

	// Check if user is already a member
	_, err = s.userRepo.GetUserByEmail(ctx, tenantID, email)
	if err == nil {
		return nil, errors.ErrUserAlreadyMember
	}

	// Get role to verify it belongs to tenant
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, errors.ErrRoleNotFound
	}
	if role.TenantID != tenantID {
		return nil, errors.ErrRoleNotFound
	}

	// Generate invitation token and hash it
	token := s.tokenService.GenerateInvitationToken()
	tokenHash := s.hasher.HashToken(token)

	// Create invitation
	invitation := &models.Invitation{
		InvitationID: uuid.New().String(),
		TenantID:     tenantID,
		Email:        email,
		RoleID:       roleID,
		TokenHash:    tokenHash,
		Status:       models.InvitationStatusPending,
		ExpiresAt:    time.Now().Add(72 * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = s.tenantRepo.CreateInvitation(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email
	err = s.emailProvider.SendInvitationEmail(email, token)
	if err != nil {
		// Log but don't fail — invitation is already persisted
		fmt.Printf("Failed to send invitation email: %v\n", err)
	}

	// Emit audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  tenantID,
		Action:    models.AuditInvitationSent,
		TargetID:  &invitation.InvitationID,
		Metadata:  map[string]interface{}{"role_id": roleID},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return invitation, nil
}

func (s *invitationService) ListInvitations(ctx context.Context, tenantID string, limit, offset int) ([]models.Invitation, error) {
	return s.tenantRepo.ListInvitations(ctx, tenantID, limit, offset)
}

func (s *invitationService) RevokeInvitation(ctx context.Context, actorID, invitationID string) error {
	// Get invitation by ID (no more list scan)
	invitation, err := s.tenantRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return errors.ErrInvitationNotFound
	}

	// Get actor to verify permissions
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return errors.ErrUserNotFound
	}
	if actor.TenantID != invitation.TenantID {
		return errors.ErrForbidden
	}

	// Check actor is owner or admin
	if err := s.requireOwnerOrAdmin(ctx, actor); err != nil {
		return err
	}

	// Check if invitation is still pending
	if invitation.Status != models.InvitationStatusPending {
		return errors.ErrInvitationExpiredOrUsed
	}

	// Update invitation status to expired (revoked)
	err = s.tenantRepo.UpdateInvitationStatus(ctx, invitationID, models.InvitationStatusExpired)
	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	// Emit audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  invitation.TenantID,
		Action:    models.AuditInvitationRevoked,
		TargetID:  &invitationID,
		Metadata:  map[string]interface{}{"reason": "revoked_by_admin"},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *invitationService) ResendInvitation(ctx context.Context, actorID, invitationID string) error {
	// Get invitation by ID
	invitation, err := s.tenantRepo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return errors.ErrInvitationNotFound
	}

	// Get actor to verify permissions
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return errors.ErrUserNotFound
	}
	if actor.TenantID != invitation.TenantID {
		return errors.ErrForbidden
	}

	// Check actor is owner or admin
	if err := s.requireOwnerOrAdmin(ctx, actor); err != nil {
		return err
	}

	// Check if invitation is still pending
	if invitation.Status != models.InvitationStatusPending {
		return errors.ErrInvitationExpiredOrUsed
	}

	// Expire old invitation and create a fresh one with new token and extended expiry
	err = s.tenantRepo.UpdateInvitationStatus(ctx, invitationID, models.InvitationStatusExpired)
	if err != nil {
		return fmt.Errorf("failed to expire old invitation: %w", err)
	}

	// Generate new token
	token := s.tokenService.GenerateInvitationToken()
	tokenHash := s.hasher.HashToken(token)

	newInvitation := &models.Invitation{
		InvitationID: uuid.New().String(),
		TenantID:     invitation.TenantID,
		Email:        invitation.Email,
		RoleID:       invitation.RoleID,
		TokenHash:    tokenHash,
		Status:       models.InvitationStatusPending,
		ExpiresAt:    time.Now().Add(72 * time.Hour),
		CreatedAt:    time.Now(),
	}

	err = s.tenantRepo.CreateInvitation(ctx, newInvitation)
	if err != nil {
		return fmt.Errorf("failed to create new invitation: %w", err)
	}

	// Send invitation email
	err = s.emailProvider.SendInvitationEmail(invitation.Email, token)
	if err != nil {
		fmt.Printf("Failed to send invitation email: %v\n", err)
	}

	// Emit audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  invitation.TenantID,
		Action:    models.AuditInvitationResent,
		TargetID:  &newInvitation.InvitationID,
		Metadata:  map[string]interface{}{"old_invitation_id": invitationID},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}
	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

// requireOwnerOrAdmin checks that the actor's role has owner or admin-level permissions.
// Uses a simple heuristic: checks for "users.write" permission which only owner/admin roles should have.
func (s *invitationService) requireOwnerOrAdmin(ctx context.Context, actor *models.User) error {
	if actor.RoleID == nil {
		return errors.ErrForbidden
	}
	permissions, err := s.roleRepo.GetRolePermissions(ctx, *actor.RoleID)
	if err != nil {
		return errors.ErrForbidden
	}
	for _, p := range permissions {
		if p.PermissionKey == "users.write" {
			return nil
		}
	}
	return errors.ErrForbidden
}

