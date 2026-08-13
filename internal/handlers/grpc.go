package handlers

import (
	domainerrors "auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/validation"
	"context"
	"errors"
	"fmt"
	"time"

	pb "auth-haven/pkg/proto"
	"auth-haven/pkg/proto/common"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedUserServiceServer
	pb.UnimplementedSessionServiceServer

	authService         interfaces.AuthService
	registrationService interfaces.RegistrationService
	passwordService     interfaces.PasswordService
	sessionService      interfaces.SessionService
	validator           *validation.CustomValidator
}

var (
	_ pb.AuthServiceServer    = (*GRPCHandler)(nil)
	_ pb.UserServiceServer    = (*GRPCHandler)(nil)
	_ pb.SessionServiceServer = (*GRPCHandler)(nil)
)

func NewGRPCHandler(
	authService interfaces.AuthService,
	registrationService interfaces.RegistrationService,
	passwordService interfaces.PasswordService,
	sessionService interfaces.SessionService,
) *GRPCHandler {
	return &GRPCHandler{
		authService:         authService,
		registrationService: registrationService,
		passwordService:     passwordService,
		sessionService:      sessionService,
		validator:           validation.NewCustomValidator(),
	}
}

func (h *GRPCHandler) VerifyMFA(ctx context.Context, req *pb.VerifyMFARequest) (*pb.LoginResponse, error) {
	input := models.VerifyMFARequest{TempToken: req.GetTempToken(), Code: req.GetCode()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}

	result, err := h.authService.VerifyMFA(ctx, input.TempToken, input.Code)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}
	return loginResponse(result), nil
}

// Login implements gRPC login
func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Extract tenant ID from request or context
	tenantID := req.GetTenantId()
	if tenantID == "" {
		if tid, ok := ctx.Value("tenant-id").(string); ok {
			tenantID = tid
		}
	}

	input := models.LoginRequest{TenantID: tenantID, Email: req.GetEmail(), Password: req.GetPassword()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}

	result, err := h.authService.Login(ctx, input.TenantID, input.Email, input.Password)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}

	return loginResponse(result), nil
}

// RefreshToken implements gRPC token refresh
func (h *GRPCHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*common.Tokens, error) {
	input := models.RefreshTokenRequest{RefreshToken: req.GetRefreshToken()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}
	tokens, err := h.authService.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}

	return &common.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *GRPCHandler) RequestPasswordReset(ctx context.Context, req *pb.RequestPasswordResetRequest) (*pb.RequestPasswordResetResponse, error) {
	input := models.RequestPasswordResetRequest{TenantID: req.GetTenantId(), Email: req.GetEmail()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}

	// Match HTTP anti-enumeration behavior: accepted requests are indistinguishable.
	_ = h.passwordService.RequestPasswordReset(ctx, input.TenantID, input.Email)
	return &pb.RequestPasswordResetResponse{Success: true}, nil
}

func (h *GRPCHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	input := models.ResetPasswordRequest{Token: req.GetToken(), NewPassword: req.GetNewPassword()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}
	if err := h.passwordService.ResetPassword(ctx, input.Token, input.NewPassword); err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}
	return &pb.ResetPasswordResponse{Success: true}, nil
}

func loginResponse(result *models.LoginResult) *pb.LoginResponse {
	return &pb.LoginResponse{
		MfaRequired:  result.MFARequired,
		TempToken:    result.TempToken,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		SessionId:    result.SessionID,
	}
}

// CreatePersonalUser implements gRPC personal user creation
func (h *GRPCHandler) CreatePersonalUser(ctx context.Context, req *pb.CreatePersonalUserRequest) (*pb.RegisterResponse, error) {
	input := models.RegisterIndividualRequest{Email: req.GetEmail(), Password: req.GetPassword(), FullName: req.GetFullName()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}

	user, err := h.registrationService.RegisterIndividual(ctx, input.Email, input.Password, input.FullName)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}

	// Return user info in RegisterResponse
	var roleId int64
	if user.RoleID != nil {
		roleId = *user.RoleID
	}
	return &pb.RegisterResponse{
		User: &pb.User{
			UserId:    user.UserID,
			TenantId:  user.TenantID,
			Email:     user.Email,
			FullName:  user.FullName,
			Status:    fmt.Sprintf("%d", int16(user.Status)),
			RoleId:    roleId,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

// CreateCompanyAndOwner implements gRPC company and owner creation
func (h *GRPCHandler) CreateCompanyAndOwner(ctx context.Context, req *pb.CreateCompanyAndOwnerRequest) (*pb.RegisterResponse, error) {
	input := models.RegisterOrgUserRequest{Email: req.GetOwnerEmail(), Password: req.GetOwnerPassword(), FullName: req.GetOwnerFullName()}
	if err := h.validator.ValidateStruct(input); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("%w: %v", domainerrors.ErrValidation, err))
	}

	// This would need to create a tenant first, then register the user
	// For now, we'll implement a simplified version

	user, err := h.registrationService.RegisterOrgUser(ctx, input.Email, input.Password, input.FullName)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}

	// Return user info in RegisterResponse
	var roleId int64
	if user.RoleID != nil {
		roleId = *user.RoleID
	}
	return &pb.RegisterResponse{
		User: &pb.User{
			UserId:    user.UserID,
			TenantId:  user.TenantID,
			Email:     user.Email,
			FullName:  user.FullName,
			Status:    fmt.Sprintf("%d", int16(user.Status)),
			RoleId:    roleId,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

// ListSessions implements gRPC session listing
func (h *GRPCHandler) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	// Get user ID from context
	userID, ok := ctx.Value("user-id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	if req.UserId != "" && req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	sessions, err := h.sessionService.ListSessions(ctx, userID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}
	result := make([]*pb.Session, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, &pb.Session{
			SessionId: session.SessionID,
			UserId:    session.UserID,
			IpAddress: session.IPAddress,
			UserAgent: session.UserAgent,
			CreatedAt: session.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return &pb.ListSessionsResponse{
		Sessions: result,
		Total:    int32(len(result)),
	}, nil
}

// RevokeSession implements gRPC session revocation
func (h *GRPCHandler) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse, error) {
	// Get user ID from context for authorization
	userID, ok := ctx.Value("user-id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	if err := h.sessionService.RevokeSession(ctx, userID, req.SessionId); err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}
	return &pb.RevokeSessionResponse{
		Success: true,
	}, nil
}

// RevokeAllSessions implements gRPC revocation of all user sessions
func (h *GRPCHandler) RevokeAllSessions(ctx context.Context, req *pb.RevokeAllSessionsRequest) (*pb.RevokeAllSessionsResponse, error) {
	// Get user ID from context
	userID, ok := ctx.Value("user-id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	if req.UserId != "" && req.UserId != userID {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	if err := h.authService.LogoutAll(ctx, userID); err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}
	return &pb.RevokeAllSessionsResponse{
		Success:      true,
		RevokedCount: 0,
	}, nil
}

// mapErrorToGRPCStatus converts domain errors to appropriate gRPC status codes
func (h *GRPCHandler) mapErrorToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domainerrors.ErrValidation), errors.As(err, new(domainerrors.ValidationError)), errors.As(err, new(domainerrors.MultiValidationError)):
		return status.Error(codes.InvalidArgument, "invalid request")
	case errors.Is(err, domainerrors.ErrInvalidCredentials), errors.Is(err, domainerrors.ErrTenantNotFound):
		return status.Error(codes.Unauthenticated, "authentication failed")
	case errors.Is(err, domainerrors.ErrInvalidToken), errors.Is(err, domainerrors.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, "invalid or expired token")
	case errors.Is(err, domainerrors.ErrUserDisabled), errors.Is(err, domainerrors.ErrTenantSuspended):
		return status.Error(codes.PermissionDenied, "access denied")
	case errors.Is(err, domainerrors.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, "email already registered")
	case errors.Is(err, domainerrors.ErrSessionNotFound):
		return status.Error(codes.NotFound, "session not found")
	case errors.Is(err, domainerrors.ErrForbidden):
		return status.Error(codes.PermissionDenied, "access denied")
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
