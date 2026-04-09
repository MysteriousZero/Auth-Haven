package interfaces

import (
	"auth-haven/internal/domain/models"
	"context"
)

type AuthService interface {
	Login(ctx context.Context, tenantID, email, password string) (*models.LoginResult, error)
	VerifyMFA(ctx context.Context, tempToken, code string) (*models.LoginResult, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
	RefreshToken(ctx context.Context, refreshTokenRaw string) (*models.TokenPair, error)
}

type PasswordService interface {
	RequestPasswordReset(ctx context.Context, tenantID, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
}

type MFAService interface {
	EnrollTOTP(ctx context.Context, userID string) (*models.TOTPEnrollment, error)
	ActivateTOTP(ctx context.Context, userID, mfaID, code string) error
	EnrollSMS(ctx context.Context, userID, phoneNumber string) error
	ActivateSMS(ctx context.Context, userID, mfaID, code string) error
	DisableMFAMethod(ctx context.Context, userID, mfaID string) error
	DeleteMFAMethod(ctx context.Context, userID, mfaID string) error
	ListMFAMethods(ctx context.Context, userID string) ([]models.UserMFAMethod, error)
}

type TokenService interface {
	GenerateTokenPair(userID, tenantID string, roleID *int64) (*models.TokenPair, error)
	GenerateTempToken(userID string) (string, error)
	ValidateTempToken(tempToken string) (string, error) // returns userID
	GenerateResetToken() string
	GenerateInvitationToken() string
	ValidateAccessToken(ctx context.Context, token string) (*models.Claims, error)
}

type Hasher interface {
	HashPassword(password string) (string, error)
	CompareHash(password, hash string) bool
	HashToken(token string) string
}

type MFAVerifier interface {
	VerifyCode(method models.MFAMethodType, secret, code string) bool
}

type TOTPGenerator interface {
	GenerateSecret() string
	GenerateQRCodeURL(secret, email string) string
}

type EmailProvider interface {
	SendResetEmail(email, token string) error
	SendInvitationEmail(email, token string) error
}

type DomainChecker interface {
	IsPublicDomain(email string) bool
	ExtractDomain(email string) (string, error)
}
