package service

import (
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"testing"
	"time"
)

type ttlAuthRepository struct {
	interfaces.AuthRepository
	session *models.Session
	refresh *models.RefreshToken
}

func (r *ttlAuthRepository) CreateSession(_ context.Context, session *models.Session) error {
	r.session = session
	return nil
}

func (r *ttlAuthRepository) CreateRefreshToken(_ context.Context, token *models.RefreshToken) error {
	r.refresh = token
	return nil
}

type ttlUserRepository struct{ interfaces.UserRepository }

func (ttlUserRepository) UpdateLastLogin(context.Context, string) error { return nil }

type ttlAuditRepository struct{ interfaces.AuditRepository }

func (ttlAuditRepository) CreateAuditLog(context.Context, *models.AuditLog) error { return nil }

type ttlTokenService struct{ interfaces.TokenService }

func (ttlTokenService) GenerateTokenPair(string, string, *int64) (*models.TokenPair, error) {
	return &models.TokenPair{AccessToken: "access", RefreshToken: "refresh"}, nil
}

type ttlHasher struct{ interfaces.Hasher }

func (ttlHasher) HashToken(token string) string { return "hashed-" + token }

func TestAuthServicePersistsConfiguredSessionAndRefreshTTLs(t *testing.T) {
	now := time.Date(2026, time.August, 12, 12, 0, 0, 0, time.UTC)
	repository := &ttlAuthRepository{}
	service := &authService{
		userRepo: ttlUserRepository{}, authRepo: repository, auditRepo: ttlAuditRepository{},
		tokenService: ttlTokenService{}, hasher: ttlHasher{},
		sessionTTL: 37 * time.Minute, refreshTTL: 11 * time.Hour,
		now: func() time.Time { return now },
	}
	user := &models.User{UserID: "user", TenantID: "tenant"}
	if _, err := service.completeLogin(context.Background(), user, &models.Tenant{}); err != nil {
		t.Fatalf("completeLogin() error = %v", err)
	}
	if got, want := repository.session.ExpiresAt, now.Add(37*time.Minute); !got.Equal(want) {
		t.Fatalf("session expiry = %v, want %v", got, want)
	}
	if got, want := repository.refresh.ExpiresAt, now.Add(11*time.Hour); !got.Equal(want) {
		t.Fatalf("refresh expiry = %v, want %v", got, want)
	}
}
