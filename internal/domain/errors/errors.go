package errors

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserDisabled            = errors.New("user disabled")
	ErrUserAlreadyMember       = errors.New("user already member")
	ErrEmailTaken              = errors.New("email already taken")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrPublicDomainNotAllowed  = errors.New("public domain not allowed")
	ErrWeakPassword            = errors.New("weak password")
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrTenantSuspended         = errors.New("tenant suspended")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token expired")
	ErrSessionNotFound         = errors.New("session not found")
	ErrDeviceNotFound          = errors.New("device not found")
	ErrRoleNotFound            = errors.New("role not found")
	ErrRoleInUse               = errors.New("role in use")
	ErrMFAMethodNotFound       = errors.New("mfa method not found")
	ErrMFAAlreadyEnabled       = errors.New("mfa already enabled")
	ErrInvalidMFACode          = errors.New("invalid mfa code")
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationExpiredOrUsed = errors.New("invitation expired or used")
	ErrPasswordResetNotFound   = errors.New("password reset not found")
	ErrForbidden               = errors.New("forbidden")
	ErrValidation              = errors.New("validation error")
)

// HTTP Status mapping
func HTTPError(err error) (int, string) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return 401, "ErrInvalidCredentials"
	case errors.Is(err, ErrUserDisabled):
		return 403, "ErrUserDisabled"
	case errors.Is(err, ErrTenantSuspended):
		return 403, "ErrTenantSuspended"
	case errors.Is(err, ErrEmailTaken):
		return 409, "ErrEmailTaken"
	case errors.Is(err, ErrUserAlreadyMember):
		return 409, "ErrUserAlreadyMember"
	case errors.Is(err, ErrPublicDomainNotAllowed):
		return 400, "ErrPublicDomainNotAllowed"
	case errors.Is(err, ErrWeakPassword):
		return 422, "ErrWeakPassword"
	case errors.Is(err, ErrInvalidEmail):
		return 422, "ErrInvalidEmail"
	case errors.Is(err, ErrInvalidToken):
		return 401, "ErrInvalidToken"
	case errors.Is(err, ErrTokenExpired):
		return 401, "ErrTokenExpired"
	case errors.Is(err, ErrSessionNotFound):
		return 404, "ErrSessionNotFound"
	case errors.Is(err, ErrDeviceNotFound):
		return 404, "ErrDeviceNotFound"
	case errors.Is(err, ErrRoleNotFound):
		return 404, "ErrRoleNotFound"
	case errors.Is(err, ErrMFAMethodNotFound):
		return 404, "ErrMFAMethodNotFound"
	case errors.Is(err, ErrMFAAlreadyEnabled):
		return 409, "ErrMFAAlreadyEnabled"
	case errors.Is(err, ErrInvalidMFACode):
		return 400, "ErrInvalidMFACode"
	case errors.Is(err, ErrInvitationNotFound):
		return 404, "ErrInvitationNotFound"
	case errors.Is(err, ErrInvitationExpiredOrUsed):
		return 400, "ErrInvitationExpiredOrUsed"
	case errors.Is(err, ErrPasswordResetNotFound):
		return 404, "ErrPasswordResetNotFound"
	case errors.Is(err, ErrForbidden):
		return 403, "ErrForbidden"
	case errors.Is(err, ErrValidation):
		return 400, "ErrValidation"
	default:
		return 500, "InternalServerError"
	}
}

// Custom error types
type ValidationError struct {
	Field   string
	Message string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", v.Field, v.Message)
}

type MultiValidationError struct {
	Errors []ValidationError
}

func (m MultiValidationError) Error() string {
	var msg string
	for _, err := range m.Errors {
		msg += err.Error() + "; "
	}
	return msg
}
