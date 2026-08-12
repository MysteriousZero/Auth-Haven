package service

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	domainerrors "auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
)

type passwordResetRepoStub struct {
	interfaces.PasswordResetRepository
	reset     *models.PasswordReset
	getErr    error
	updatedTo models.PasswordResetStatus
}

func (s *passwordResetRepoStub) GetPasswordResetByHash(context.Context, string) (*models.PasswordReset, error) {
	return s.reset, s.getErr
}

func (s *passwordResetRepoStub) UpdatePasswordResetStatus(_ context.Context, _ string, status models.PasswordResetStatus) error {
	s.updatedTo = status
	return nil
}

type passwordUserRepoStub struct {
	interfaces.UserRepository
	user           *models.User
	passwordHash   string
	revoked        bool
	expiredResetID string
}

func (s *passwordUserRepoStub) GetUserByID(context.Context, string) (*models.User, error) {
	return s.user, nil
}

func (s *passwordUserRepoStub) UpdatePasswordHash(_ context.Context, _ string, hash string) error {
	s.passwordHash = hash
	return nil
}

func (s *passwordUserRepoStub) RevokeAllUserTokens(context.Context, string) error {
	s.revoked = true
	return nil
}

func (s *passwordUserRepoStub) UpdatePasswordResetStatus(_ context.Context, resetID string, _ models.PasswordResetStatus) error {
	s.expiredResetID = resetID
	return nil
}

type passwordAuthRepoStub struct {
	interfaces.AuthRepository
	deleted bool
}

func (s *passwordAuthRepoStub) DeleteAllUserSessions(context.Context, string) error {
	s.deleted = true
	return nil
}

type passwordHasherStub struct{ interfaces.Hasher }

func (passwordHasherStub) HashToken(token string) string       { return "hash:" + token }
func (passwordHasherStub) HashPassword(string) (string, error) { return "new-hash", nil }

func TestResetPasswordRejectsUnknownExpiredAndReplayedTokens(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name  string
		reset *models.PasswordReset
		err   error
		want  error
	}{
		{name: "unknown", err: stderrors.New("missing"), want: domainerrors.ErrInvalidToken},
		{name: "replayed", reset: &models.PasswordReset{Status: models.PasswordResetStatusUsed, ExpiresAt: now.Add(time.Hour)}, want: domainerrors.ErrTokenExpired},
		{name: "expired", reset: &models.PasswordReset{ResetID: "expired", Status: models.PasswordResetStatusPending, ExpiresAt: now.Add(-time.Hour)}, want: domainerrors.ErrTokenExpired},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resets := &passwordResetRepoStub{reset: tc.reset, getErr: tc.err}
			users := &passwordUserRepoStub{}
			service := NewPasswordService(users, resets, &passwordAuthRepoStub{}, passwordHasherStub{}, nil, auditRepoStub{}, nil)
			err := service.ResetPassword(context.Background(), "credential", "new-password")
			if !stderrors.Is(err, tc.want) || users.passwordHash != "" {
				t.Fatalf("ResetPassword() error = %v, hash = %q", err, users.passwordHash)
			}
			if tc.name == "expired" && users.expiredResetID != "expired" {
				t.Fatalf("expired reset was not marked: %q", users.expiredResetID)
			}
		})
	}
}

func TestResetPasswordConsumesCredentialAndInvalidatesAuthenticationState(t *testing.T) {
	user := &models.User{UserID: "user-a", TenantID: "tenant-a", Status: models.UserStatusActive}
	resets := &passwordResetRepoStub{reset: &models.PasswordReset{
		ResetID: "reset-a", UserID: user.UserID, Status: models.PasswordResetStatusPending, ExpiresAt: time.Now().Add(time.Hour),
	}}
	users := &passwordUserRepoStub{user: user}
	auth := &passwordAuthRepoStub{}
	service := NewPasswordService(users, resets, auth, passwordHasherStub{}, nil, auditRepoStub{}, nil)

	if err := service.ResetPassword(context.Background(), "credential", "new-password"); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if users.passwordHash != "new-hash" || resets.updatedTo != models.PasswordResetStatusUsed || !users.revoked || !auth.deleted {
		t.Fatalf("state: hash=%q reset=%v revoked=%v sessionsDeleted=%v", users.passwordHash, resets.updatedTo, users.revoked, auth.deleted)
	}
}
