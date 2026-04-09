package interfaces

import (
	"auth-haven/internal/domain/models"
	"context"
)

type TenantRepository interface {
	GetTenantByID(ctx context.Context, tenantID string) (*models.Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	CreateTenant(ctx context.Context, tenant *models.Tenant) error
	UpdateTenant(ctx context.Context, tenant *models.Tenant) error
	CreateInvitation(ctx context.Context, invitation *models.Invitation) error
	GetInvitationByHash(ctx context.Context, tokenHash string) (*models.Invitation, error)
	GetInvitationByID(ctx context.Context, invitationID string) (*models.Invitation, error)
	UpdateInvitationStatus(ctx context.Context, invitationID string, status models.InvitationStatus) error
	ListInvitations(ctx context.Context, tenantID string, limit, offset int) ([]models.Invitation, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, tenantID, email string) (*models.User, error)
	ListUsers(ctx context.Context, tenantID string, limit, offset int) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error
	UpdateUserRole(ctx context.Context, userID string, roleID int64) error
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	CreatePasswordReset(ctx context.Context, reset *models.PasswordReset) error
	GetPasswordResetByHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error)
	UpdatePasswordResetStatus(ctx context.Context, resetID string, status models.PasswordResetStatus) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type RoleRepository interface {
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, roleID int64) (*models.Role, error)
	GetRoleByName(ctx context.Context, tenantID, name string) (*models.Role, error)
	ListRoles(ctx context.Context, tenantID string, limit, offset int) ([]models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, roleID int64) error
	CreateRolePermission(ctx context.Context, roleID int64, permissionKey string) error
	AddRolePermission(ctx context.Context, roleID int64, permissionKey string) error
	DeleteRolePermissions(ctx context.Context, roleID int64) error
	GetRolePermissions(ctx context.Context, roleID int64) ([]models.RolePermission, error)
}

type AuthRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
	GetActiveSessions(ctx context.Context, userID string) ([]models.Session, error)
	ListSessions(ctx context.Context, userID string, limit, offset int) ([]models.Session, error)
	ListSessionsByDevice(ctx context.Context, deviceID string) ([]models.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID string) error

	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type MFAMethodRepository interface {
	CreateMFAMethod(ctx context.Context, method *models.UserMFAMethod) error
	ListMFAMethods(ctx context.Context, userID string) ([]models.UserMFAMethod, error)
	UpdateMFAMethod(ctx context.Context, mfaID string, enabled bool) error
	DeleteMFAMethod(ctx context.Context, mfaID string) error
}

type PasswordResetRepository interface {
	CreatePasswordReset(ctx context.Context, reset *models.PasswordReset) error
	GetPasswordResetByHash(ctx context.Context, tokenHash string) (*models.PasswordReset, error)
	UpdatePasswordResetStatus(ctx context.Context, resetID string, status models.PasswordResetStatus) error
}

type AuditRepository interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	ListTenantLogs(ctx context.Context, tenantID string, cursor string, limit int) ([]models.AuditLog, string, error)
	ListUserLogs(ctx context.Context, userID string, cursor string, limit int) ([]models.AuditLog, string, error)
}

type DeviceRepository interface {
	CreateDevice(ctx context.Context, device *models.Device) error
	GetDeviceByID(ctx context.Context, deviceID string) (*models.Device, error)
	ListDevices(ctx context.Context, userID string, limit, offset int) ([]models.Device, error)
	UpdateDevice(ctx context.Context, device *models.Device) error
	UpdateDeviceLastSeen(ctx context.Context, deviceID string) error
	DeleteDevice(ctx context.Context, deviceID string) error
}
