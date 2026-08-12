package models

import (
	"time"
)

type LoginResult struct {
	MFARequired  bool   `json:"mfa_required"`
	TempToken    string `json:"temp_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserID     string    `json:"sub"`
	TenantID   string    `json:"tenant_id"`
	TenantType int16     `json:"tenant_type"`
	RoleID     *int64    `json:"role_id,omitempty"`
	Email      string    `json:"email"`
	IssuedAt   time.Time `json:"iat"`
	ExpiresAt  time.Time `json:"exp"`
	Issuer     string    `json:"iss"`
	Audience   string    `json:"aud"`
}

// Request/Response DTOs
type LoginRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=12,max=128"`
}

type VerifyMFARequest struct {
	TempToken string `json:"temp_token" validate:"required,min=32"`
	Code      string `json:"code" validate:"required,len=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=32"`
}

type RegisterIndividualRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=12,max=128,complexpassword"`
	FullName string `json:"full_name" validate:"required,min=1,max=255"`
}

type RegisterOrgUserRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=12,max=128,complexpassword"`
	FullName string `json:"full_name" validate:"required,min=1,max=255"`
}

type RegisterWithInvitationRequest struct {
	InvitationToken string `json:"invitation_token" validate:"required,min=32"`
	Password        string `json:"password" validate:"required,min=12,max=128,complexpassword"`
	FullName        string `json:"full_name" validate:"required,min=1,max=255"`
}

type RequestPasswordResetRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	Email    string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required,min=32"`
	NewPassword string `json:"new_password" validate:"required,min=12,max=128,complexpassword"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=12,max=128,complexpassword"`
}
