package server

import (
	"auth-haven/internal/config"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/handlers"
	"auth-haven/internal/middleware"
	"auth-haven/internal/repository"
	"auth-haven/internal/service"
	"auth-haven/internal/utils"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func StartHTTP(cfg *config.Config, db *sql.DB, tokenService interfaces.TokenService) error {
	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Check Redis connection (Ping)
	ctxPing, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctxPing).Err(); err != nil {
		if !cfg.Redis.RateLimitFailOpen {
			return fmt.Errorf("required Redis dependency unavailable: %w", err)
		}
		log.Printf("Redis unavailable; cache and rate limiting are operating in configured fail-open mode: %v", err)
	}

	// Repositories
	baseTenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	baseAuthRepo := repository.NewAuthRepository(db)
	mfaRepo := repository.NewMFAMethodRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)

	// Cache Layer
	cache := repository.NewRedisCache(redisClient)
	tenantRepo := repository.NewCachedTenantRepository(baseTenantRepo, cache)
	authRepo := repository.NewCachedAuthRepository(baseAuthRepo, cache)

	// Core services
	hasher := service.NewHasher()
	emailProvider := service.NewEmailProvider(cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.Username, cfg.Email.Password, cfg.Email.From, cfg.Email.BaseURL)
	domainChecker := service.NewDomainChecker()
	totpGen := service.NewTOTPGenerator()

	// Decode MFA encryption key
	mfaKey, err := base64.StdEncoding.DecodeString(cfg.Security.MFAEncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decode MFA encryption key: %w", err)
	}
	if len(mfaKey) != 32 {
		return fmt.Errorf("MFA encryption key must be 32 bytes (AES-256), got %d", len(mfaKey))
	}

	// Domain services
	authService := service.NewAuthService(
		tenantRepo, userRepo, authRepo, mfaRepo, auditRepo, tokenService, hasher, mfaKey,
		cfg.Auth.RefreshTokenTTL, cfg.Auth.SessionTTL,
	)
	registrationService := service.NewRegistrationService(
		tenantRepo, userRepo, roleRepo, hasher, auditRepo, domainChecker,
	)
	passwordService := service.NewPasswordService(
		userRepo, passwordResetRepo, authRepo, hasher, tokenService, auditRepo, emailProvider,
	)
	mfaService := service.NewMFAService(
		mfaRepo, userRepo, auditRepo, totpGen, mfaKey,
	)
	sessionService := service.NewSessionService(
		authRepo, deviceRepo, userRepo, auditRepo,
	)
	invitationService := service.NewInvitationService(
		tenantRepo, userRepo, roleRepo, auditRepo, emailProvider, tokenService, hasher,
	)
	roleService := service.NewRoleService(
		roleRepo, userRepo, tenantRepo, auditRepo,
	)
	auditService := service.NewAuditService(
		auditRepo, userRepo, tenantRepo,
	)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	registrationHandler := handlers.NewRegistrationHandler(registrationService)
	passwordHandler := handlers.NewPasswordHandler(passwordService)
	mfaHandler := handlers.NewMFAHandler(mfaService)
	sessionHandler := handlers.NewSessionHandler(sessionService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Rate Limiter
	limiter := middleware.NewRateLimiter(redisClient, cfg.Redis.RateLimitFailOpen)

	// Suppress unused variable warning for roleService — will be used when role handler is added
	_ = roleService

	// Gin router
	router := gin.Default()
	router.Use(corsMiddleware(cfg.CORS))
	router.Use(requestIDMiddleware())
	router.Use(loggingMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/ready", readinessHandler(db, redisClient, !cfg.Redis.RateLimitFailOpen))
	router.GET("/metrics", databaseMetricsHandler(db))

	v1 := router.Group("/v1")

	// ── Public auth routes ──────────────────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/login", limiter.LoginRateLimit(), authHandler.Login)
		auth.POST("/mfa/verify", limiter.MFAVerifyRateLimit(), authHandler.VerifyMFA)
		auth.POST("/token/refresh", authHandler.RefreshToken)

		// Registration
		auth.POST("/register/individual", registrationHandler.RegisterIndividual)
		auth.POST("/register/org", registrationHandler.RegisterOrgUser)
		auth.POST("/register/invitation", registrationHandler.RegisterWithInvitation)

		// Password (public — no auth required)
		auth.POST("/password/forgot", limiter.PasswordForgotRateLimit(), passwordHandler.ForgotPassword)
		auth.POST("/password/reset", passwordHandler.ResetPassword)
	}

	// ── Protected routes (requires valid JWT) ───────────────────────────────
	authMW := middleware.AuthMiddleware(tokenService)
	me := v1.Group("/me", authMW)
	{
		// Auth actions
		me.POST("/logout", authHandler.Logout)
		me.POST("/logout-all", authHandler.LogoutAll)

		// Password (auth required for change)
		me.POST("/password/change", passwordHandler.ChangePassword)

		// MFA
		me.GET("/mfa", mfaHandler.ListMFAMethods)
		me.POST("/mfa/totp/enroll", mfaHandler.EnrollTOTP)
		me.POST("/mfa/totp/activate", mfaHandler.ActivateTOTP)
		me.DELETE("/mfa/:mfa_id", mfaHandler.DisableMFAMethod)

		// Sessions
		me.GET("/sessions", sessionHandler.ListSessions)
		me.DELETE("/sessions/:session_id", sessionHandler.RevokeSession)

		// Devices
		me.GET("/devices", sessionHandler.ListDevices)
		me.DELETE("/devices/:device_id", sessionHandler.RemoveDevice)

		// My audit logs
		me.GET("/audit-logs", auditHandler.ListUserLogs)
	}

	// ── Tenant-scoped routes ─────────────────────────────────────────────────
	tenants := v1.Group("/tenants", authMW)
	{
		// Invitations
		invitations := tenants.Group("/:tenant_id/invitations")
		{
			invitations.POST("", invitationHandler.SendInvitation)
			invitations.GET("", invitationHandler.ListInvitations)
			invitations.DELETE("/:invitation_id", invitationHandler.RevokeInvitation)
			invitations.POST("/:invitation_id/resend", invitationHandler.ResendInvitation)
		}

		// Tenant audit logs
		tenants.GET("/:tenant_id/audit-logs", auditHandler.ListTenantLogs)
	}

	addr := cfg.Server.Host + cfg.Server.HTTPPort
	log.Printf("HTTP server listening on %s", addr)
	return router.Run(addr)
}

func corsMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowed[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		_, explicit := allowed[origin]
		_, wildcard := allowed["*"]
		if origin != "" && !explicit && !wildcard {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if origin != "" {
			if wildcard {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func readinessHandler(db *sql.DB, redisClient *redis.Client, redisRequired bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		dependencies := gin.H{"database": "ok", "redis": "ok"}
		ready := true
		if err := db.PingContext(ctx); err != nil {
			dependencies["database"] = "unavailable"
			ready = false
		}
		if err := redisClient.Ping(ctx).Err(); err != nil {
			dependencies["redis"] = "unavailable"
			if redisRequired {
				ready = false
			}
		}
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"status": map[bool]string{true: "ready", false: "not_ready"}[ready], "dependencies": dependencies})
	}
}

func databaseMetricsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := db.Stats()
		lines := []string{
			fmt.Sprintf("auth_haven_db_max_open_connections %d", stats.MaxOpenConnections),
			fmt.Sprintf("auth_haven_db_open_connections %d", stats.OpenConnections),
			fmt.Sprintf("auth_haven_db_in_use_connections %d", stats.InUse),
			fmt.Sprintf("auth_haven_db_idle_connections %d", stats.Idle),
			fmt.Sprintf("auth_haven_db_wait_count %d", stats.WaitCount),
		}
		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(strings.Join(lines, "\n")+"\n"))
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("trace_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Also inject into context for services
		ctx := utils.SetTraceID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
