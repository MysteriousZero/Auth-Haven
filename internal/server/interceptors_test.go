package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	pb "auth-haven/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type interceptorTokenServiceStub struct {
	interfaces.TokenService
	claims *models.Claims
	err    error
}

func (s interceptorTokenServiceStub) ValidateAccessToken(context.Context, string) (*models.Claims, error) {
	return s.claims, s.err
}

func TestUnaryInterceptorRejectsCredentialsWithoutLeakingValidationDetails(t *testing.T) {
	tests := []struct {
		name     string
		metadata metadata.MD
		service  interceptorTokenServiceStub
	}{
		{name: "missing", metadata: metadata.MD{}},
		{name: "empty", metadata: metadata.Pairs("authorization", "")},
		{name: "empty bearer", metadata: metadata.Pairs("authorization", "Bearer ")},
		{name: "invalid", metadata: metadata.Pairs("authorization", "Bearer invalid"), service: interceptorTokenServiceStub{err: errors.New("signature from key secret-kid failed")}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), tc.metadata)
			_, err := Unary(tc.service)(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.SessionService/ListSessions"}, func(context.Context, any) (any, error) {
				t.Fatal("protected handler called for rejected credential")
				return nil, nil
			})
			if status.Code(err) != codes.Unauthenticated {
				t.Fatalf("status = %s, want %s", status.Code(err), codes.Unauthenticated)
			}
			if strings.Contains(status.Convert(err).Message(), "secret-kid") {
				t.Fatalf("validation detail leaked: %v", err)
			}
		})
	}
}

func TestUnaryInterceptorInjectsClaimsAndLeavesPublicMethodsPublic(t *testing.T) {
	claims := &models.Claims{UserID: "user-a", TenantID: "tenant-a"}
	service := interceptorTokenServiceStub{claims: claims}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid"))
	_, err := Unary(service)(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/auth.SessionService/ListSessions"}, func(ctx context.Context, _ any) (any, error) {
		if ctx.Value("user-id") != claims.UserID || ctx.Value("tenant-id") != claims.TenantID {
			t.Fatalf("claims not injected: user=%v tenant=%v", ctx.Value("user-id"), ctx.Value("tenant-id"))
		}
		return nil, nil
	})
	if err != nil {
		t.Fatalf("protected call error = %v", err)
	}

	_, err = Unary(interceptorTokenServiceStub{err: errors.New("must not validate")})(context.Background(), nil,
		&grpc.UnaryServerInfo{FullMethod: "/auth.AuthService/Login"},
		func(context.Context, any) (any, error) { return nil, nil },
	)
	if err != nil {
		t.Fatalf("public call error = %v", err)
	}
}

func TestUnaryInterceptorClassifiesEveryRegisteredRPC(t *testing.T) {
	public := []string{
		pb.AuthService_Login_FullMethodName,
		pb.AuthService_VerifyMFA_FullMethodName,
		pb.AuthService_RefreshToken_FullMethodName,
		pb.AuthService_RequestPasswordReset_FullMethodName,
		pb.AuthService_ResetPassword_FullMethodName,
		pb.UserService_CreatePersonalUser_FullMethodName,
		pb.UserService_CreateCompanyAndOwner_FullMethodName,
	}
	for _, method := range public {
		t.Run("public "+method, func(t *testing.T) {
			_, err := Unary(interceptorTokenServiceStub{err: errors.New("must not validate")})(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: method}, func(context.Context, any) (any, error) { return nil, nil })
			if err != nil {
				t.Fatalf("public method rejected: %v", err)
			}
		})
	}

	protected := []string{
		pb.SessionService_ListSessions_FullMethodName,
		pb.SessionService_RevokeSession_FullMethodName,
		pb.SessionService_RevokeAllSessions_FullMethodName,
	}
	for _, method := range protected {
		t.Run("protected "+method, func(t *testing.T) {
			_, err := Unary(interceptorTokenServiceStub{})(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: method}, func(context.Context, any) (any, error) {
				t.Fatal("protected handler called without credentials")
				return nil, nil
			})
			if status.Code(err) != codes.Unauthenticated {
				t.Fatalf("status = %s, want %s", status.Code(err), codes.Unauthenticated)
			}
		})
	}
}
