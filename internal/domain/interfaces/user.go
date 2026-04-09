package interfaces

import (
	"auth-haven/internal/domain/models"
	"context"
)

type UserService interface {
	GetProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID, fullName string) (*models.User, error)
	ListUsers(ctx context.Context, tenantID string, limit, offset int) ([]models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	UpdateUserStatus(ctx context.Context, actorID, targetUserID string, status models.UserStatus) error
	AssignRole(ctx context.Context, actorID, targetUserID string, roleID int64) error
}

type RegistrationService interface {
	RegisterIndividual(ctx context.Context, email, password, fullName string) (*models.User, error)
	RegisterOrgUser(ctx context.Context, email, password, fullName string) (*models.User, error)
	RegisterWithInvitation(ctx context.Context, invitationToken, password, fullName string) (*models.User, error)
}

type SessionService interface {
	ListSessions(ctx context.Context, userID string, limit, offset int) ([]models.Session, error)
	RevokeSession(ctx context.Context, actorID, sessionID string) error
	ListDevices(ctx context.Context, userID string, limit, offset int) ([]models.Device, error)
	RemoveDevice(ctx context.Context, actorID, deviceID string) error
}

type RoleService interface {
	CreateRole(ctx context.Context, actorID, tenantID, name string) (*models.Role, error)
	DeleteRole(ctx context.Context, actorID string, roleID int64) error
	ListRoles(ctx context.Context, tenantID string, limit, offset int) ([]models.Role, error)
	GetRolePermissions(ctx context.Context, roleID int64) ([]models.RolePermission, error)
	SetRolePermissions(ctx context.Context, actorID string, roleID int64, keys []string) error
}

type InvitationService interface {
	SendInvitation(ctx context.Context, actorID, tenantID, email string, roleID int64) (*models.Invitation, error)
	ListInvitations(ctx context.Context, tenantID string, limit, offset int) ([]models.Invitation, error)
	RevokeInvitation(ctx context.Context, actorID, invitationID string) error
	ResendInvitation(ctx context.Context, actorID, invitationID string) error
}

type TenantService interface {
	CreateTenant(ctx context.Context, name, domain string) (*models.Tenant, error)
	GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error)
	UpdateTenantStatus(ctx context.Context, actorID, tenantID string, status models.TenantStatus) error
}

type AuditService interface {
	ListUserLogs(ctx context.Context, actorID, userID string, cursor string, limit int) ([]models.AuditLog, string, error)
	ListTenantLogs(ctx context.Context, actorID, tenantID string, cursor string, limit int) ([]models.AuditLog, string, error)
}
