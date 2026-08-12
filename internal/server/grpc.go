package server

import (
	"auth-haven/internal/config"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/handlers"
	"auth-haven/internal/repository"
	"auth-haven/internal/service"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net"

	pb "auth-haven/pkg/proto"

	"google.golang.org/grpc"
)

func StartGRPC(cfg *config.Config, db *sql.DB, tokenService interfaces.TokenService) error {
	// Initialize repositories
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	authRepo := repository.NewAuthRepository(db)
	mfaRepo := repository.NewMFAMethodRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize services
	hasher := service.NewHasher()

	// Decode MFA encryption key
	mfaKey, err := base64.StdEncoding.DecodeString(cfg.Security.MFAEncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decode MFA encryption key: %w", err)
	}
	if len(mfaKey) != 32 {
		return fmt.Errorf("MFA encryption key must be 32 bytes (AES-256), got %d", len(mfaKey))
	}

	authService := service.NewAuthService(
		tenantRepo,
		userRepo,
		authRepo,
		mfaRepo,
		auditRepo,
		tokenService,
		hasher,
		mfaKey,
		cfg.Auth.RefreshTokenTTL,
		cfg.Auth.SessionTTL,
	)

	registrationService := service.NewRegistrationService(
		tenantRepo,
		userRepo,
		roleRepo,
		hasher,
		auditRepo,
		service.NewDomainChecker(),
	)

	// Initialize gRPC handler
	grpcHandler := handlers.NewGRPCHandler(authService, registrationService, nil, nil, tokenService)

	lis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			ValidationInterceptor(),
			Unary(tokenService),
			LoggingInterceptor,
		),
	}

	s := grpc.NewServer(opts...)

	// Register services
	pb.RegisterAuthServiceServer(s, grpcHandler)
	pb.RegisterUserServiceServer(s, grpcHandler)
	pb.RegisterSessionServiceServer(s, grpcHandler)

	log.Printf("gRPC server running on %s", cfg.Server.GRPCPort)
	return s.Serve(lis)
}
