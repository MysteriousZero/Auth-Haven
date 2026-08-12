package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type mfaService struct {
	mfaRepo       interfaces.MFAMethodRepository
	userRepo      interfaces.UserRepository
	auditRepo     interfaces.AuditRepository
	totpGen       interfaces.TOTPGenerator
	encryptionKey []byte
}

func NewMFAService(
	mfaRepo interfaces.MFAMethodRepository,
	userRepo interfaces.UserRepository,
	auditRepo interfaces.AuditRepository,
	totpGen interfaces.TOTPGenerator,
	encryptionKey []byte,
) interfaces.MFAService {
	return &mfaService{
		mfaRepo:       mfaRepo,
		userRepo:      userRepo,
		auditRepo:     auditRepo,
		totpGen:       totpGen,
		encryptionKey: encryptionKey,
	}
}

func (s *mfaService) EnrollTOTP(ctx context.Context, userID string) (*models.TOTPEnrollment, error) {
	// Check if TOTP is already enabled
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list MFA methods: %w", err)
	}

	for _, method := range methods {
		if method.Type == models.MFAMethodTypeTOTP && method.Enabled {
			return nil, errors.ErrMFAAlreadyEnabled
		}
	}

	// Create MFA method record
	mfaID := uuid.New().String()
	secret := s.totpGen.GenerateSecret()
	mfaMethod := &models.UserMFAMethod{
		MFAID:   mfaID,
		UserID:  userID,
		Type:    models.MFAMethodTypeTOTP,
		Enabled: false, // Not enabled until verification
	}

	encryptedSecret, err := utils.Encrypt(secret, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt MFA secret: %w", err)
	}
	mfaMethod.Secret = &encryptedSecret
	mfaMethod.CreatedAt = time.Now().UTC()

	err = s.mfaRepo.CreateMFAMethod(ctx, mfaMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to create MFA method: %w", err)
	}

	// Get user for QR code
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate QR code URL
	qrURL := s.totpGen.GenerateQRCodeURL(secret, user.Email)

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserMFAEnrolled,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

		TraceID: utils.GetTraceID(ctx),
	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return &models.TOTPEnrollment{
		MFAID:     mfaID,
		Secret:    secret,
		QRCodeURL: qrURL,
	}, nil
}

func (s *mfaService) ActivateTOTP(ctx context.Context, userID, mfaID, code string) error {
	// Get MFA method
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list MFA methods: %w", err)
	}

	var targetMethod *models.UserMFAMethod
	for _, method := range methods {
		if method.MFAID == mfaID && method.Type == models.MFAMethodTypeTOTP {
			targetMethod = &method
			break
		}
	}

	if targetMethod == nil {
		return errors.ErrMFAMethodNotFound
	}

	if targetMethod.Enabled {
		return errors.ErrMFAAlreadyEnabled
	}

	// Decrypt the secret
	decryptedSecret, err := utils.Decrypt(*targetMethod.Secret, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	if !s.verifyTOTP(decryptedSecret, code) {
		// Get user for tenant ID
		user, _ := s.userRepo.GetUserByID(ctx, userID)
		tenantID := ""
		if user != nil {
			tenantID = user.TenantID
		}

		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			UserID:    &userID,
			TenantID:  tenantID,
			Action:    models.AuditActionUserMFAFailed,
			Metadata:  map[string]interface{}{"reason": "invalid_code", "type": "totp"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

			TraceID: utils.GetTraceID(ctx),
		})
		return errors.ErrInvalidMFACode
	}

	// Enable the MFA method
	err = s.mfaRepo.UpdateMFAMethod(ctx, mfaID, true)
	if err != nil {
		return fmt.Errorf("failed to enable MFA method: %w", err)
	}

	// Get user for audit log
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserMFASuccess,
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

func (s *mfaService) EnrollSMS(ctx context.Context, userID, phoneNumber string) error {
	// Validate phone number format
	if !s.isValidPhoneNumber(phoneNumber) {
		return errors.ErrValidation
	}

	// Check if SMS is already enabled
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list MFA methods: %w", err)
	}

	for _, method := range methods {
		if method.Type == models.MFAMethodTypeSMS && method.Enabled {
			return errors.ErrMFAAlreadyEnabled
		}
	}

	// Generate verification code
	verificationCode := s.generateSMSCode()
	// TODO: Send SMS via SMS provider
	fmt.Printf("SMS verification code: %s\n", verificationCode)

	// Create MFA method record
	mfaID := uuid.New().String()
	mfaMethod := &models.UserMFAMethod{
		MFAID:   mfaID,
		UserID:  userID,
		Type:    models.MFAMethodTypeSMS,
		Enabled: false,
		Secret:  &verificationCode,
	}

	encryptedPhone, err := utils.Encrypt(phoneNumber, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt phone number: %w", err)
	}
	mfaMethod.PhoneNumber = &encryptedPhone
	mfaMethod.CreatedAt = time.Now().UTC()

	err = s.mfaRepo.CreateMFAMethod(ctx, mfaMethod)
	if err != nil {
		return fmt.Errorf("failed to create MFA method: %w", err)
	}

	// Record audit log
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err == nil {
		auditLog := &models.AuditLog{
			LogID:     uuid.New().String(),
			UserID:    &userID,
			TenantID:  user.TenantID,
			Action:    models.AuditActionUserMFAEnrolled,
			Metadata:  map[string]interface{}{"type": "sms"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

			TraceID: utils.GetTraceID(ctx),
		}
		s.auditRepo.CreateAuditLog(ctx, auditLog)
	}

	return nil
}

func (s *mfaService) ActivateSMS(ctx context.Context, userID, mfaID, code string) error {
	// Get MFA method
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list MFA methods: %w", err)
	}

	var targetMethod *models.UserMFAMethod
	for _, method := range methods {
		if method.MFAID == mfaID && method.Type == models.MFAMethodTypeSMS {
			targetMethod = &method
			break
		}
	}

	if targetMethod == nil {
		return errors.ErrMFAMethodNotFound
	}

	if targetMethod.Enabled {
		return errors.ErrMFAAlreadyEnabled
	}

	// For SMS, we temporarily stored the verification code in the secret field
	// during enrollment. This is a simple implementation. In production,
	// use a separate verification code store (e.g. Redis).
	if targetMethod.Secret == nil || *targetMethod.Secret != code {
		// Get user for tenant ID
		user, _ := s.userRepo.GetUserByID(ctx, userID)
		tenantID := ""
		if user != nil {
			tenantID = user.TenantID
		}

		s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
			LogID:     uuid.New().String(),
			UserID:    &userID,
			TenantID:  tenantID,
			Action:    models.AuditActionUserMFAFailed,
			Metadata:  map[string]interface{}{"reason": "invalid_code", "type": "sms"},
			IPAddress: utils.GetClientIP(ctx),
			UserAgent: utils.GetUserAgent(ctx),

			TraceID: utils.GetTraceID(ctx),
		})
		return errors.ErrInvalidMFACode
	}

	// Enable the MFA method and restore phone number
	err = s.mfaRepo.UpdateMFAMethod(ctx, mfaID, true)
	if err != nil {
		return fmt.Errorf("failed to enable MFA method: %w", err)
	}

	// Get user for audit log
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserMFASuccess,
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

func (s *mfaService) DisableMFAMethod(ctx context.Context, userID, mfaID string) error {
	// Get MFA method
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list MFA methods: %w", err)
	}

	var targetMethod *models.UserMFAMethod
	for _, method := range methods {
		if method.MFAID == mfaID {
			targetMethod = &method
			break
		}
	}

	if targetMethod == nil {
		return errors.ErrMFAMethodNotFound
	}

	if !targetMethod.Enabled {
		return errors.ErrMFAMethodNotFound
	}

	// Disable the MFA method
	err = s.mfaRepo.UpdateMFAMethod(ctx, mfaID, false)
	if err != nil {
		return fmt.Errorf("failed to disable MFA method: %w", err)
	}

	// Get user for audit log
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserMFADisabled,
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

func (s *mfaService) ListMFAMethods(ctx context.Context, userID string) ([]models.UserMFAMethod, error) {
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list MFA methods: %w", err)
	}

	// Don't expose secrets in the response
	result := make([]models.UserMFAMethod, len(methods))
	for i, method := range methods {
		emptySecret := ""
		result[i] = models.UserMFAMethod{
			MFAID:     method.MFAID,
			UserID:    method.UserID,
			Type:      method.Type,
			Secret:    &emptySecret, // Don't expose secrets
			Enabled:   method.Enabled,
			CreatedAt: method.CreatedAt,
		}
	}

	return result, nil
}

func (s *mfaService) DeleteMFAMethod(ctx context.Context, userID, mfaID string) error {
	// Verify the method belongs to this user
	methods, err := s.mfaRepo.ListMFAMethods(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to list MFA methods: %w", err)
	}

	var found bool
	for _, method := range methods {
		if method.MFAID == mfaID {
			found = true
			break
		}
	}

	if !found {
		return errors.ErrMFAMethodNotFound
	}

	if err := s.mfaRepo.DeleteMFAMethod(ctx, mfaID); err != nil {
		return fmt.Errorf("failed to delete MFA method: %w", err)
	}

	// Get user for audit log
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &userID,
		TenantID:  user.TenantID,
		Action:    models.AuditActionUserMFADisabled,
		TargetID:  &mfaID,
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

		TraceID: utils.GetTraceID(ctx),
	}

	if err := s.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

// Helper methods
func (s *mfaService) verifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (s *mfaService) generateBackupCodes() []string {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code := s.generateSecureCode(8) // 8-character backup codes
		codes[i] = code
	}
	return codes
}

func (s *mfaService) generateSMSCode() string {
	return s.generateSecureCode(6) // 6-digit SMS codes
}

func (s *mfaService) generateSecureCode(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func (s *mfaService) isValidPhoneNumber(phoneNumber string) bool {
	// Basic phone number validation
	// In production, use a proper phone number validation library
	return len(phoneNumber) >= 10 && strings.HasPrefix(phoneNumber, "+")
}
