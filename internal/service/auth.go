package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type authService struct {
	tenantRepo    interfaces.TenantRepository
	userRepo      interfaces.UserRepository
	authRepo      interfaces.AuthRepository
	mfaRepo       interfaces.MFAMethodRepository
	auditRepo     interfaces.AuditRepository
	tokenService  interfaces.TokenService
	hasher        interfaces.Hasher
	encryptionKey []byte
}

func NewAuthService(
	tenantRepo interfaces.TenantRepository,
	userRepo interfaces.UserRepository,
	authRepo interfaces.AuthRepository,
	mfaRepo interfaces.MFAMethodRepository,
	auditRepo interfaces.AuditRepository,
	tokenService interfaces.TokenService,
	hasher interfaces.Hasher,
	encryptionKey []byte,
) interfaces.AuthService {
	return &authService{
		tenantRepo:    tenantRepo,
		userRepo:      userRepo,
		authRepo:      authRepo,
		mfaRepo:       mfaRepo,
		auditRepo:     auditRepo,
		tokenService:  tokenService,
		hasher:        hasher,
		encryptionKey: encryptionKey,
	}
}

func (s *authService) Login(ctx context.Context, tenantID, email, password string) (*models.LoginResult, error) {
	// Get tenant
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			TenantID:  tenantID,
			Action:    models.AuditActionUserLoginFailed,
			Metadata:  map[string]interface{}{"reason": "tenant_not_found"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

		TraceID:   utils.GetTraceID(ctx),

		})
		return nil, errors.ErrTenantNotFound
	}

	if !tenant.IsActive() {
		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			TenantID:  tenantID,
			Action:    models.AuditActionUserLoginFailed,
			Metadata:  map[string]interface{}{"reason": "tenant_suspended"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

		TraceID:   utils.GetTraceID(ctx),

		})
		return nil, errors.ErrTenantSuspended
	}

	// Get user
	user, err := s.userRepo.GetUserByEmail(ctx, tenantID, email)
	if err != nil {
		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			TenantID:  tenantID,
			Action:    models.AuditActionUserLoginFailed,
			Metadata:  map[string]interface{}{"reason": "invalid_credentials", "email_attempt": email},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

		TraceID:   utils.GetTraceID(ctx),

		})
		return nil, errors.ErrInvalidCredentials
	}

	if !user.IsActive() {
		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			UserID:    &user.UserID,
			TenantID:  tenantID,
			Action:    models.AuditActionUserLoginFailed,
			Metadata:  map[string]interface{}{"reason": "user_disabled"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

		TraceID:   utils.GetTraceID(ctx),

		})
		return nil, errors.ErrUserDisabled
	}

	// Verify password
	if !s.hasher.CompareHash(password, user.PasswordHash) {
		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			UserID:    &user.UserID,
			TenantID:  tenantID,
			Action:    models.AuditActionUserLoginFailed,
			Metadata:  map[string]interface{}{"reason": "invalid_credentials"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

		TraceID:   utils.GetTraceID(ctx),

		})
		return nil, errors.ErrInvalidCredentials
	}

	// Check MFA
	mfaMethods, err := s.mfaRepo.ListMFAMethods(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check MFA methods: %w", err)
	}

	hasEnabledMFA := false
	for _, method := range mfaMethods {
		if method.Enabled {
			hasEnabledMFA = true
			break
		}
	}

	if hasEnabledMFA {
		// Generate temp token for MFA verification
		tempToken, err := s.tokenService.GenerateTempToken(user.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp token: %w", err)
		}

		return &models.LoginResult{
			MFARequired: true,
			TempToken:   tempToken,
		}, nil
	}

	// No MFA required, complete login
	return s.completeLogin(ctx, user, tenant)
}

func (s *authService) VerifyMFA(ctx context.Context, tempToken, code string) (*models.LoginResult, error) {
	// Validate temp token
	userID, err := s.tokenService.ValidateTempToken(tempToken)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsActive() {
		return nil, errors.ErrUserDisabled
	}

	// Get MFA methods
	mfaMethods, err := s.mfaRepo.ListMFAMethods(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA methods: %w", err)
	}

	// Verify code against enabled methods
	validCode := false
	for _, method := range mfaMethods {
		if method.Enabled {
			if method.Type == models.MFAMethodTypeTOTP && method.Secret != nil {
				// Decrypt the secret
				decryptedSecret, err := utils.Decrypt(*method.Secret, s.encryptionKey)
				if err != nil {
					return nil, fmt.Errorf("failed to decrypt MFA secret: %w", err)
				}

				if totp.Validate(code, decryptedSecret) {
					validCode = true
					break
				}
			} else if method.Type == models.MFAMethodTypeSMS && method.PhoneNumber != nil {
				// For SMS, we temporarily stored the verification code in the secret field
				if method.Secret != nil && *method.Secret == code {
					validCode = true
					break
				}
			}
		}
	}

	if !validCode {
		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			UserID:    &user.UserID,
			TenantID:  user.TenantID,
			Action:    models.AuditActionUserMFAFailed,
			Metadata:  map[string]interface{}{"reason": "invalid_code"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

		TraceID:   utils.GetTraceID(ctx),

		})
		return nil, errors.ErrInvalidMFACode
	}

	// Complete login
	tenant, err := s.tenantRepo.GetTenantByID(ctx, user.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return s.completeLogin(ctx, user, tenant)
}

func (s *authService) completeLogin(ctx context.Context, user *models.User, tenant *models.Tenant) (*models.LoginResult, error) {
	// Update last login
	err := s.userRepo.UpdateLastLogin(ctx, user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Create session
	sessionID := uuid.New().String()
	session := &models.Session{
		SessionID: sessionID,
		UserID:    user.UserID,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour), // 24 hour session
	}

	err = s.authRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(user.UserID, user.TenantID, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create refresh token
	refreshToken := &models.RefreshToken{
		TokenID:   uuid.New().String(),
		UserID:    user.UserID,
		TokenHash: s.hasher.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour), // 7 days
	}

	err = s.authRepo.CreateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &user.UserID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserLogin,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return &models.LoginResult{
		MFARequired:  false,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		SessionID:    sessionID,
	}, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	// Get session
	session, err := s.authRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Delete session
	err = s.authRepo.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Revoke refresh token
	err = s.authRepo.RevokeAllUserTokens(ctx, session.UserID)
	if err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &session.UserID,
		TenantID:  "", // Will be filled by middleware
		Action:    models.AuditActionUserLogout,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID string) error {
	// Delete all sessions
	err := s.authRepo.DeleteAllUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}

	// Revoke all tokens
	err = s.authRepo.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  "", // Will be filled by middleware
		Action:    models.AuditActionUserLogoutAll,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshTokenRaw string) (*models.TokenPair, error) {
	// Hash the incoming token
	tokenHash := s.hasher.HashToken(refreshTokenRaw)

	// Get refresh token from DB
	refreshToken, err := s.authRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	if refreshToken.Revoked {
		return nil, errors.ErrInvalidToken
	}

	if time.Now().UTC().After(refreshToken.ExpiresAt) {
		return nil, errors.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	if !user.IsActive() {
		return nil, errors.ErrUserDisabled
	}

	// Revoke old token
	err = s.authRepo.RevokeRefreshToken(ctx, refreshToken.TokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Generate new tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(user.UserID, user.TenantID, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Create new refresh token
	newRefreshToken := &models.RefreshToken{
		TokenID:   uuid.New().String(),
		UserID:    user.UserID,
		TokenHash: s.hasher.HashToken(tokenPair.RefreshToken),
		UserAgent: refreshToken.UserAgent,
		IPAddress: refreshToken.IPAddress,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}

	err = s.authRepo.CreateRefreshToken(ctx, newRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	return tokenPair, nil
}

