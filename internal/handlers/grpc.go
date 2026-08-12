package handlers

import (
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/validation"
	"context"
	"fmt"
	"strings"
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
	tokenService        interfaces.TokenService
	validator           *validation.CustomValidator
}

func NewGRPCHandler(
	authService interfaces.AuthService,
	registrationService interfaces.RegistrationService,
	passwordService interfaces.PasswordService,
	sessionService interfaces.SessionService,
	tokenService interfaces.TokenService,
) *GRPCHandler {
	return &GRPCHandler{
		authService:         authService,
		registrationService: registrationService,
		passwordService:     passwordService,
		sessionService:      sessionService,
		tokenService:        tokenService,
		validator:           validation.NewCustomValidator(),
	}
}

// Login implements gRPC login
func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate input
	if err := h.validator.ValidateStruct(req); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("validation: %w", err))
	}

	// Extract tenant ID from request or context
	tenantID := req.TenantId
	if tenantID == "" {
		if tid, ok := ctx.Value("tenant-id").(string); ok {
			tenantID = tid
		}
	}

	if tenantID == "" {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("tenant_id required"))
	}

	result, err := h.authService.Login(ctx, tenantID, req.Email, req.Password)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}

	if result.MFARequired {
		return &pb.LoginResponse{
			MfaRequired: true,
			TempToken:   result.TempToken,
		}, nil
	}

	return &pb.LoginResponse{
		MfaRequired:  false,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		SessionId:    result.SessionID,
	}, nil
}

// RefreshToken implements gRPC token refresh
func (h *GRPCHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*common.Tokens, error) {
	tokens, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, h.mapErrorToGRPCStatus(err)
	}

	return &common.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// CreatePersonalUser implements gRPC personal user creation
func (h *GRPCHandler) CreatePersonalUser(ctx context.Context, req *pb.CreatePersonalUserRequest) (*pb.RegisterResponse, error) {
	// Validate input
	if err := h.validator.ValidateStruct(req); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("validation: %w", err))
	}

	user, err := h.registrationService.RegisterIndividual(ctx, req.Email, req.Password, req.FullName)
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
	// Validate input
	if err := h.validator.ValidateStruct(req); err != nil {
		return nil, h.mapErrorToGRPCStatus(fmt.Errorf("validation: %w", err))
	}

	// This would need to create a tenant first, then register the user
	// For now, we'll implement a simplified version

	user, err := h.registrationService.RegisterOrgUser(ctx, req.OwnerEmail, req.OwnerPassword, req.OwnerFullName)
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

	// Import the errors package to check error types
	switch {
	case strings.Contains(err.Error(), "validation"):
		return status.Error(codes.InvalidArgument, err.Error())
	case strings.Contains(err.Error(), "tenant not found"):
		return status.Errorf(codes.Unauthenticated, "authentication failed")
	case strings.Contains(err.Error(), "invalid credentials"):
		return status.Errorf(codes.Unauthenticated, "authentication failed")
	case strings.Contains(err.Error(), "user disabled"):
		return status.Errorf(codes.PermissionDenied, "user account disabled")
	case strings.Contains(err.Error(), "tenant suspended"):
		return status.Errorf(codes.PermissionDenied, "tenant suspended")
	case strings.Contains(err.Error(), "email taken"):
		return status.Errorf(codes.AlreadyExists, "email already registered")
	case strings.Contains(err.Error(), "invalid token"):
		return status.Errorf(codes.Unauthenticated, "invalid or expired token")
	case strings.Contains(err.Error(), "forbidden"):
		return status.Errorf(codes.PermissionDenied, "access denied")
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
