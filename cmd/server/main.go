package main

import (
	"auth-haven/internal/config"
	"auth-haven/internal/server"
	"auth-haven/internal/service"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// Connect to DB
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}
	defer conn.Close()
	conn.SetMaxOpenConns(cfg.Database.MaxConnections)
	conn.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	if err := conn.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}

	keyProvider, err := service.NewStaticSigningKeyProvider(
		cfg.Auth.JWTKeyID,
		cfg.Auth.JWTPrivateKey,
		cfg.Auth.JWTVerificationKeys,
	)
	if err != nil {
		log.Fatalf("failed to load signing keys: %v", err)
	}
	tokenService := service.NewTokenService(
		keyProvider,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.TempTokenTTL,
		cfg.Auth.TokenClockSkew,
	)

	// Start servers
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Server.HTTPPort)
		if err := server.StartHTTP(cfg, conn, tokenService); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	log.Printf("Starting gRPC server on %s", cfg.Server.GRPCPort)
	if err := server.StartGRPC(cfg, conn, tokenService); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
