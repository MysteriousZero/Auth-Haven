package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type userService struct {
	userRepo   interfaces.UserRepository
	tenantRepo interfaces.TenantRepository
	roleRepo   interfaces.RoleRepository
	auditRepo  interfaces.AuditRepository
}

func NewUserService(
	userRepo interfaces.UserRepository,
	tenantRepo interfaces.TenantRepository,
	roleRepo interfaces.RoleRepository,
	auditRepo interfaces.AuditRepository,
) interfaces.UserService {
	return &userService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		roleRepo:   roleRepo,
		auditRepo:  auditRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Don't expose password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID, fullName string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.FullName = fullName
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Don't expose password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) ListUsers(ctx context.Context, tenantID string, limit, offset int) ([]models.User, error) {
	// Enforce Personal Tenant isolation
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrTenantNotFound
	}
	if tenant.Type == models.TenantTypePersonal {
		return nil, errors.ErrForbidden
	}

	users, err := s.userRepo.ListUsers(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Don't expose password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}

	return users, nil
}

func (s *userService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Don't expose password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *userService) UpdateUserStatus(ctx context.Context, actorID, targetUserID string, status models.UserStatus) error {
	// Get target user
	targetUser, err := s.userRepo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	// RBAC Check: Actor must be owner or admin in the same tenant
	if err := s.requireOwnerOrAdmin(ctx, actorID, targetUser.TenantID); err != nil {
		return err
	}

	err = s.userRepo.UpdateUserStatus(ctx, targetUserID, status)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  targetUser.TenantID,
		Action:    models.AuditUserStatusChanged,
		TargetID:  &targetUserID,
		Metadata:  map[string]interface{}{"old_status": targetUser.Status, "new_status": status},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *userService) AssignRole(ctx context.Context, actorID, targetUserID string, roleID int64) error {
	// Get target user
	targetUser, err := s.userRepo.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	// Enforce: Personal tenants cannot have roles
	tenant, err := s.tenantRepo.GetTenantByID(ctx, targetUser.TenantID)
	if err != nil {
		return err
	}
	if tenant.Type == models.TenantTypePersonal {
		return errors.ErrForbidden
	}

	// RBAC Check: Actor must be owner or admin in the same tenant
	if err := s.requireOwnerOrAdmin(ctx, actorID, targetUser.TenantID); err != nil {
		return err
	}

	// Get role to ensure it exists and belongs to same tenant
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	if role.TenantID != targetUser.TenantID {
		return errors.ErrForbidden
	}

	// Update user role
	targetUser.RoleID = &roleID
	err = s.userRepo.UpdateUser(ctx, targetUser)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// requireOwnerOrAdmin checks that the actor's role has owner or admin-level permissions (users.write).
func (s *userService) requireOwnerOrAdmin(ctx context.Context, actorID, tenantID string) error {
	actor, err := s.userRepo.GetUserByID(ctx, actorID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if actor.TenantID != tenantID {
		return errors.ErrForbidden
	}

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

type registrationService struct {
	tenantRepo    interfaces.TenantRepository
	userRepo      interfaces.UserRepository
	roleRepo      interfaces.RoleRepository
	hasher        interfaces.Hasher
	auditRepo     interfaces.AuditRepository
	domainChecker interfaces.DomainChecker
}

func NewRegistrationService(
	tenantRepo interfaces.TenantRepository,
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
	hasher interfaces.Hasher,
	auditRepo interfaces.AuditRepository,
	domainChecker interfaces.DomainChecker,
) interfaces.RegistrationService {
	return &registrationService{
		tenantRepo:    tenantRepo,
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		hasher:        hasher,
		auditRepo:     auditRepo,
		domainChecker: domainChecker,
	}
}

func (s *registrationService) RegisterIndividual(ctx context.Context, email, password, fullName string) (*models.User, error) {
	// Create personal tenant
	tenantID := uuid.New().String()
	tenant := &models.Tenant{
		TenantID: tenantID,
		Name:     fullName + "'s Personal Space",
		Type:     models.TenantTypePersonal,
		Status:   models.TenantStatusActive,
	}

	err := s.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create user
	userID := uuid.New().String()
	passwordHash, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		UserID:       userID,
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Status:       models.UserStatusActive,
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  tenantID,
		Action:    models.AuditActionUserRegistered,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Don't expose password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *registrationService) RegisterOrgUser(ctx context.Context, email, password, fullName string) (*models.User, error) {
	// Validate email format and extract domain
	domain, err := s.domainChecker.ExtractDomain(email)
	if err != nil {
		return nil, errors.ErrInvalidEmail
	}

	// Check if domain is public
	if s.domainChecker.IsPublicDomain(email) {
		return nil, errors.ErrPublicDomainNotAllowed
	}

	// Try to find tenant by domain
	tenant, err := s.tenantRepo.GetTenantByDomain(ctx, domain)
	if err != nil && err != errors.ErrTenantNotFound {
		return nil, err
	}

	var tenantID string
	var roleID *int64

	if err == errors.ErrTenantNotFound {
		// Auto-create org tenant from domain
		tenantID = uuid.New().String()
		newTenant := &models.Tenant{
			TenantID: tenantID,
			Name:     fmt.Sprintf("%s Organization", strings.ToUpper(domain)),
			Domain:   &domain,
			Type:     models.TenantTypeOrganization,
			Status:   models.TenantStatusActive,
		}

		err = s.tenantRepo.CreateTenant(ctx, newTenant)
		if err != nil {
			// Race condition check: Did someone else create it between our GetTenantByDomain and CreateTenant?
			tenant, retryErr := s.tenantRepo.GetTenantByDomain(ctx, domain)
			if retryErr == nil {
				// Success! Someone else created it
				tenantID = tenant.TenantID
			} else {
				return nil, fmt.Errorf("failed to create tenant: %w", err)
			}
		}

		// First user gets owner role - create owner role
		ownerRole := &models.Role{
			TenantID: tenantID,
			Name:     "Owner",
		}

		err = s.roleRepo.CreateRole(ctx, ownerRole)
		if err != nil {
			return nil, fmt.Errorf("failed to create owner role: %w", err)
		}

		// Add owner permissions
		ownerPermissions := []string{
			"users.read", "users.write", "users.delete",
			"roles.read", "roles.write", "roles.delete",
			"tenant.read", "tenant.write",
			"audit.read",
		}

		for _, permission := range ownerPermissions {
			err = s.roleRepo.AddRolePermission(ctx, ownerRole.RoleID, permission)
			if err != nil {
				return nil, fmt.Errorf("failed to add owner permission: %w", err)
			}
		}

		roleID = &ownerRole.RoleID
	} else {
		// Existing tenant found
		if !tenant.IsActive() {
			return nil, errors.ErrTenantSuspended
		}

		tenantID = tenant.TenantID

		// Check if this is the first user in the tenant
		users, err := s.userRepo.ListUsers(ctx, tenantID, 1, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing users: %w", err)
		}

		if len(users) == 0 {
			// First user gets owner role
			ownerRole, err := s.roleRepo.GetRoleByName(ctx, tenantID, "Owner")
			if err != nil {
				return nil, fmt.Errorf("failed to get owner role: %w", err)
			}
			roleID = &ownerRole.RoleID
		} else {
			// Subsequent users get default member role
			memberRole, err := s.roleRepo.GetRoleByName(ctx, tenantID, "Member")
			if err != nil {
				// Create default member role if it doesn't exist
				newMemberRole := &models.Role{
					TenantID: tenantID,
					Name:     "Member",
				}

				err = s.roleRepo.CreateRole(ctx, newMemberRole)
				if err != nil {
					return nil, fmt.Errorf("failed to create member role: %w", err)
				}

				// Add basic member permissions
				memberPermissions := []string{"users.read", "tenant.read"}
				for _, permission := range memberPermissions {
					err = s.roleRepo.AddRolePermission(ctx, newMemberRole.RoleID, permission)
					if err != nil {
						return nil, fmt.Errorf("failed to add member permission: %w", err)
					}
				}

				roleID = &newMemberRole.RoleID
			} else {
				roleID = &memberRole.RoleID
			}
		}
	}

	// Check if email already exists in this tenant
	existingUser, err := s.userRepo.GetUserByEmail(ctx, tenantID, email)
	if err == nil && existingUser != nil {
		return nil, errors.ErrEmailTaken
	}

	// Create user
	userID := uuid.New().String()
	passwordHash, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		UserID:       userID,
		TenantID:     tenantID,
		RoleID:       roleID,
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Status:       models.UserStatusActive,
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  tenantID,
		Action:    models.AuditActionUserRegistered,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Don't expose password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *registrationService) RegisterWithInvitation(ctx context.Context, invitationToken, password, fullName string) (*models.User, error) {
	// Hash the invitation token to find it
	hasher := NewHasher()
	tokenHash := hasher.HashToken(invitationToken)

	// Get invitation
	invitation, err := s.tenantRepo.GetInvitationByHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.ErrInvitationNotFound
	}

	if invitation.Status != models.InvitationStatusPending {
		return nil, errors.ErrInvitationExpiredOrUsed
	}

	if time.Now().UTC().After(invitation.ExpiresAt) {
		return nil, errors.ErrInvitationExpiredOrUsed
	}

	// Create user
	userID := uuid.New().String()
	passwordHash, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		UserID:       userID,
		TenantID:     invitation.TenantID,
		RoleID:       &invitation.RoleID,
		Email:        invitation.Email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Status:       models.UserStatusActive,
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Mark invitation as used
	err = s.tenantRepo.UpdateInvitationStatus(ctx, invitation.InvitationID, models.InvitationStatusAccepted)
	if err != nil {
		fmt.Printf("Failed to update invitation status: %v\n", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  invitation.TenantID,
		Action:    models.AuditActionUserRegisteredViaInvitation,
		TargetID:  &invitation.InvitationID,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Don't expose password hash
	user.PasswordHash = ""
	return user, nil
}

