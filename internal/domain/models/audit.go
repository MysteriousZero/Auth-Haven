package models

import (
	"time"
)

type AuditLog struct {
	LogID     string                 `json:"log_id" db:"log_id"`
	UserID    *string                `json:"user_id" db:"user_id"` // NULL for system actions
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	Action    string                 `json:"action" db:"action"`
	TargetID  *string                `json:"target_id" db:"target_id"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	TraceID   string                 `json:"trace_id" db:"trace_id"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

type PaginatedAuditLogs struct {
	Logs       []AuditLog `json:"logs"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	Total      int        `json:"total"`
	NextCursor *string    `json:"next_cursor,omitempty"`
}

// Audit actions
const (
	AuditActionUserLogin                   = "user.login"
	AuditActionUserLoginFailed             = "user.login_failed"
	AuditActionUserLogout                  = "user.logout"
	AuditActionUserLogoutAll               = "user.logout_all"
	AuditActionUserMFASuccess              = "user.mfa_success"
	AuditActionUserMFAFailed               = "user.mfa_failed"
	AuditActionUserRegistered              = "user.registered"
	AuditActionUserRegisteredViaInvitation = "user.registered_via_invitation"
	AuditActionUserPasswordResetRequested  = "user.password_reset_requested"
	AuditActionUserPasswordReset           = "user.password_reset"
	AuditActionUserPasswordChanged         = "user.password_changed"
	AuditActionUserMFAEnrolled             = "user.mfa_enrolled"
	AuditActionUserMFADisabled             = "user.mfa_disabled"
	AuditUserStatusChanged                 = "user.status_changed"
	AuditUserRoleAssigned                  = "user.role_assigned"
	AuditInvitationSent                    = "invitation.sent"
	AuditInvitationAccepted                = "invitation.accepted"
	AuditInvitationRevoked                 = "invitation.revoked"
	AuditRoleCreated                       = "role.created"
	AuditRoleDeleted                       = "role.deleted"
	AuditRolePermissionsUpdated            = "role.permissions_updated"
	AuditSessionRevoked                    = "session.revoked"
	AuditDeviceRemoved                     = "device.removed"
	AuditInvitationResent                  = "invitation.resent"
)
