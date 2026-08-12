package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type passwordService struct {
	userRepo          interfaces.UserRepository
	passwordResetRepo interfaces.PasswordResetRepository
	authRepo          interfaces.AuthRepository
	hasher            interfaces.Hasher
	tokenService      interfaces.TokenService
	auditRepo         interfaces.AuditRepository
	emailProvider     interfaces.EmailProvider
}

func NewPasswordService(
	userRepo interfaces.UserRepository,
	passwordResetRepo interfaces.PasswordResetRepository,
	authRepo interfaces.AuthRepository,
	hasher interfaces.Hasher,
	tokenService interfaces.TokenService,
	auditRepo interfaces.AuditRepository,
	emailProvider interfaces.EmailProvider,
) interfaces.PasswordService {
	return &passwordService{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		authRepo:          authRepo,
		hasher:            hasher,
		tokenService:      tokenService,
		auditRepo:         auditRepo,
		emailProvider:     emailProvider,
	}
}

func (s *passwordService) RequestPasswordReset(ctx context.Context, tenantID, email string) error {
	// Find user by tenant and email
	user, err := s.userRepo.GetUserByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal if user exists or not to prevent enumeration
		return nil
	}

	if !user.IsActive() {
		// Don't reveal if user is disabled
		return nil
	}

	// Generate reset token
	resetToken := s.tokenService.GenerateResetToken()
	tokenHash := s.hasher.HashToken(resetToken)

	// Create password reset record
	resetID := uuid.New().String()
	passwordReset := &models.PasswordReset{
		ResetID:   resetID,
		UserID:    user.UserID,
		TokenHash: tokenHash,
		Status:    models.PasswordResetStatusPending,
		ExpiresAt: time.Now().UTC().Add(1 * time.Hour), // 1 hour expiry
	}

	err = s.passwordResetRepo.CreatePasswordReset(ctx, passwordReset)
	if err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	// Send reset email
	err = s.emailProvider.SendResetEmail(email, resetToken)
	if err != nil {
		// Keep the public response indistinguishable from an unknown account.
		log.Printf("password reset delivery failed: %v", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &user.UserID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserPasswordResetRequested,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

		TraceID: utils.GetTraceID(ctx),
	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *passwordService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Hash the token to find it
	tokenHash := s.hasher.HashToken(token)

	// Find password reset record
	passwordReset, err := s.passwordResetRepo.GetPasswordResetByHash(ctx, tokenHash)
	if err != nil {
		return errors.ErrInvalidToken
	}

	// Check if token is still valid
	if passwordReset.Status != models.PasswordResetStatusPending {
		return errors.ErrTokenExpired
	}

	if time.Now().UTC().After(passwordReset.ExpiresAt) {
		// Mark as expired
		s.userRepo.UpdatePasswordResetStatus(ctx, passwordReset.ResetID, models.PasswordResetStatusExpired)
		return errors.ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, passwordReset.UserID)
	if err != nil {
		return err
	}

	if !user.IsActive() {
		return errors.ErrUserDisabled
	}

	// Hash new password
	newPasswordHash, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	err = s.userRepo.UpdatePasswordHash(ctx, user.UserID, newPasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark reset token as used
	err = s.passwordResetRepo.UpdatePasswordResetStatus(ctx, passwordReset.ResetID, models.PasswordResetStatusUsed)
	if err != nil {
		fmt.Printf("Failed to mark reset token as used: %v\n", err)
	}

	// Revoke all refresh tokens for this user
	err = s.userRepo.RevokeAllUserTokens(ctx, user.UserID)
	if err != nil {
		fmt.Printf("Failed to revoke user tokens: %v\n", err)
	}

	// Delete all active sessions (spec requirement)
	err = s.authRepo.DeleteAllUserSessions(ctx, user.UserID)
	if err != nil {
		fmt.Printf("Failed to delete user sessions: %v\n", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &user.UserID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserPasswordReset,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

		TraceID: utils.GetTraceID(ctx),
	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *passwordService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsActive() {
		return errors.ErrUserDisabled
	}

	// Verify current password
	if !s.hasher.CompareHash(currentPassword, user.PasswordHash) {
		return errors.ErrInvalidCredentials
	}

	// Hash new password
	newPasswordHash, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	err = s.userRepo.UpdatePasswordHash(ctx, userID, newPasswordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Revoke all refresh tokens for this user (force re-login)
	err = s.userRepo.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		fmt.Printf("Failed to revoke user tokens: %v\n", err)
	}

	// Delete all active sessions (spec requirement)
	err = s.authRepo.DeleteAllUserSessions(ctx, userID)
	if err != nil {
		fmt.Printf("Failed to delete user sessions: %v\n", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserPasswordChanged,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

		TraceID: utils.GetTraceID(ctx),
	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}
