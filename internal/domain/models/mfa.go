package models

import (
	"time"
)

type MFAMethodType int16

const (
	MFAMethodTypeTOTP  MFAMethodType = 1
	MFAMethodTypeSMS   MFAMethodType = 2
	MFAMethodTypeEmail MFAMethodType = 3
)

type UserMFAMethod struct {
	MFAID       string        `json:"mfa_id" db:"mfa_id"`
	UserID      string        `json:"user_id" db:"user_id"`
	Type        MFAMethodType `json:"type" db:"type"`
	Secret      *string       `json:"-" db:"secret"` // Encrypted at rest
	PhoneNumber *string       `json:"phone_number,omitempty" db:"phone_number"`
	Enabled     bool          `json:"enabled" db:"enabled"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}

type TOTPEnrollment struct {
	MFAID     string `json:"mfa_id"`
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

type PasswordResetStatus int16

const (
	PasswordResetStatusPending PasswordResetStatus = 1
	PasswordResetStatusUsed    PasswordResetStatus = 2
	PasswordResetStatusExpired PasswordResetStatus = 3
)

type PasswordReset struct {
	ResetID   string              `json:"reset_id" db:"reset_id"`
	UserID    string              `json:"user_id" db:"user_id"`
	TokenHash string              `json:"-" db:"token_hash"`
	Status    PasswordResetStatus `json:"status" db:"status"`
	ExpiresAt time.Time           `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt time.Time           `json:"updated_at" db:"updated_at"`
}

type InvitationStatus int16

const (
	InvitationStatusPending  InvitationStatus = 1
	InvitationStatusAccepted InvitationStatus = 2
	InvitationStatusExpired  InvitationStatus = 3
)

type Invitation struct {
	InvitationID string           `json:"invitation_id" db:"invitation_id"`
	TenantID     string           `json:"tenant_id" db:"tenant_id"`
	RoleID       int64            `json:"role_id" db:"role_id"`
	Email        string           `json:"email" db:"email"`
	TokenHash    string           `json:"-" db:"token_hash"`
	Status       InvitationStatus `json:"status" db:"status"`
	ExpiresAt    time.Time        `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
}
