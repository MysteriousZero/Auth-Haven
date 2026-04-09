package server

import (
	domainErrors "auth-haven/internal/domain/errors"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapDomainErrorToGRPCCode converts domain errors to appropriate gRPC status codes
// according to the spec requirements
func MapDomainErrorToGRPCCode(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific domain errors
	switch {
	case errors.Is(err, domainErrors.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domainErrors.ErrUserNotFound):
		return status.Error(codes.Unauthenticated, "user not found") // Maps to Unauthenticated to prevent enumeration
	case errors.Is(err, domainErrors.ErrUserDisabled):
		return status.Error(codes.PermissionDenied, "user account disabled")
	case errors.Is(err, domainErrors.ErrTenantNotFound):
		return status.Error(codes.Unauthenticated, "tenant not found") // Maps to Unauthenticated to prevent enumeration
	case errors.Is(err, domainErrors.ErrTenantSuspended):
		return status.Error(codes.PermissionDenied, "tenant suspended")
	case errors.Is(err, domainErrors.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")
	case errors.Is(err, domainErrors.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, "token expired")
	case errors.Is(err, domainErrors.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, "email already registered")
	case errors.Is(err, domainErrors.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, "invalid email format")
	case errors.Is(err, domainErrors.ErrPublicDomainNotAllowed):
		return status.Error(codes.InvalidArgument, "public email domains not allowed")
	case errors.Is(err, domainErrors.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, "password does not meet strength requirements")
	case errors.Is(err, domainErrors.ErrSessionNotFound):
		return status.Error(codes.NotFound, "session not found")
	case errors.Is(err, domainErrors.ErrDeviceNotFound):
		return status.Error(codes.NotFound, "device not found")
	case errors.Is(err, domainErrors.ErrRoleNotFound):
		return status.Error(codes.NotFound, "role not found")
	case errors.Is(err, domainErrors.ErrRoleInUse):
		return status.Error(codes.FailedPrecondition, "role is currently assigned to users")
	case errors.Is(err, domainErrors.ErrMFAMethodNotFound):
		return status.Error(codes.NotFound, "MFA method not found")
	case errors.Is(err, domainErrors.ErrMFAAlreadyEnabled):
		return status.Error(codes.AlreadyExists, "MFA method already enabled")
	case errors.Is(err, domainErrors.ErrInvalidMFACode):
		return status.Error(codes.InvalidArgument, "invalid MFA code")
	case errors.Is(err, domainErrors.ErrInvitationNotFound):
		return status.Error(codes.NotFound, "invitation not found")
	case errors.Is(err, domainErrors.ErrInvitationExpiredOrUsed):
		return status.Error(codes.FailedPrecondition, "invitation expired or already used")
	case errors.Is(err, domainErrors.ErrForbidden):
		return status.Error(codes.PermissionDenied, "access denied")
	case errors.Is(err, domainErrors.ErrUserAlreadyMember):
		return status.Error(codes.AlreadyExists, "user is already a member")
	case errors.Is(err, domainErrors.ErrValidation):
		return status.Error(codes.InvalidArgument, "validation error")
	default:
		// For any unknown errors, return an internal server error
		return status.Error(codes.Internal, "internal server error")
	}
}
