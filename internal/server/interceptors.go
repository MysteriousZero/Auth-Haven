package server

import (
	"context"
	"log"
	"strings"
	"time"

	"auth-haven/internal/domain/interfaces"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Unary(tokenService interfaces.TokenService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		// Extract tenant ID from metadata
		if tenantIDs := md.Get("tenant-id"); len(tenantIDs) > 0 {
			if tenantIDs[0] != "" {
				ctx = context.WithValue(ctx, "tenant-id", tenantIDs[0])
			}
		}

		// Check if this is a public endpoint that doesn't require authentication
		publicMethods := map[string]bool{
			"/auth.AuthService/Login":                 true,
			"/auth.AuthService/RefreshToken":          true,
			"/auth.AuthService/RequestPasswordReset":  true,
			"/auth.AuthService/ResetPassword":         true,
			"/auth.AuthService/IntrospectToken":       true,
			"/auth.UserService/CreatePersonalUser":    true,
			"/auth.UserService/CreateCompanyAndOwner": true,
		}

		// JWT Authentication for protected endpoints
		if !publicMethods[info.FullMethod] {
			authHeaders := md.Get("authorization")
			if len(authHeaders) == 0 {
				return nil, status.Errorf(codes.Unauthenticated, "authorization header required")
			}

			token := authHeaders[0]
			if token == "" {
				return nil, status.Errorf(codes.Unauthenticated, "authorization token required")
			}

			// Remove "Bearer " prefix if present
			token = strings.TrimPrefix(token, "Bearer ")
			if token == "" {
				return nil, status.Errorf(codes.Unauthenticated, "invalid authorization format")
			}

			claims, err := tokenService.ValidateAccessToken(ctx, token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
			}

			// Inject user claims into context
			ctx = context.WithValue(ctx, "user-id", claims.UserID)
			ctx = context.WithValue(ctx, "tenant-id", claims.TenantID)
			ctx = context.WithValue(ctx, "role-id", claims.RoleID)
		}

		return handler(ctx, req)
	}
}

// Validation interceptor
func ValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// For now, just pass through. In a full implementation,
		// this would use protoc-gen-validate or similar
		return handler(ctx, req)
	}
}

// Logging interceptor
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("handling %s", info.FullMethod)
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	log.Printf("method=%s duration=%s error=%v", info.FullMethod, duration, err)

	// TODO: Push audit log asynchronously
	return resp, err
}
