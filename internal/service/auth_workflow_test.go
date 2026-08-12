package service

import (
	"context"
	stderrors "errors"
	"sync"
	"testing"
	"time"

	domainerrors "auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"

	"github.com/pquerna/otp/totp"
)

type workflowStore struct {
	interfaces.TenantRepository
	interfaces.UserRepository
	interfaces.AuthRepository
	interfaces.MFAMethodRepository
	interfaces.AuditRepository

	mu            sync.Mutex
	tenants       map[string]models.Tenant
	users         map[string]models.User
	sessions      map[string]models.Session
	refreshTokens map[string]models.RefreshToken
	mfaMethods    map[string]models.UserMFAMethod
	audits        []models.AuditLog
}

func newWorkflowStore() *workflowStore {
	return &workflowStore{
		tenants:       make(map[string]models.Tenant),
		users:         make(map[string]models.User),
		sessions:      make(map[string]models.Session),
		refreshTokens: make(map[string]models.RefreshToken),
		mfaMethods:    make(map[string]models.UserMFAMethod),
	}
}

func (s *workflowStore) CreateTenant(_ context.Context, tenant *models.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenants[tenant.TenantID] = *tenant
	return nil
}

func (s *workflowStore) GetTenantByID(_ context.Context, id string) (*models.Tenant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant, ok := s.tenants[id]
	if !ok {
		return nil, domainerrors.ErrTenantNotFound
	}
	return &tenant, nil
}

func (s *workflowStore) CreateUser(_ context.Context, user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.UserID] = *user
	return nil
}

func (s *workflowStore) GetUserByID(_ context.Context, id string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return nil, domainerrors.ErrUserNotFound
	}
	return &user, nil
}

func (s *workflowStore) GetUserByEmail(_ context.Context, tenantID, email string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.TenantID == tenantID && user.Email == email {
			copy := user
			return &copy, nil
		}
	}
	return nil, domainerrors.ErrUserNotFound
}

func (s *workflowStore) UpdateLastLogin(context.Context, string) error { return nil }

func (s *workflowStore) CreateSession(_ context.Context, session *models.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.SessionID] = *session
	return nil
}

func (s *workflowStore) GetSessionByID(_ context.Context, id string) (*models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, domainerrors.ErrSessionNotFound
	}
	return &session, nil
}

func (s *workflowStore) DeleteSession(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

func (s *workflowStore) CreateRefreshToken(_ context.Context, token *models.RefreshToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[token.TokenHash] = *token
	return nil
}

func (s *workflowStore) GetRefreshTokenByHash(_ context.Context, hash string) (*models.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.refreshTokens[hash]
	if !ok {
		return nil, domainerrors.ErrInvalidToken
	}
	return &token, nil
}

func (s *workflowStore) RevokeRefreshToken(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, token := range s.refreshTokens {
		if token.TokenID == id {
			token.Revoked = true
			s.refreshTokens[hash] = token
		}
	}
	return nil
}

func (s *workflowStore) RevokeAllUserTokens(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, token := range s.refreshTokens {
		if token.UserID == userID {
			token.Revoked = true
			s.refreshTokens[hash] = token
		}
	}
	return nil
}

func (s *workflowStore) CreateMFAMethod(_ context.Context, method *models.UserMFAMethod) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mfaMethods[method.MFAID] = *method
	return nil
}

func (s *workflowStore) ListMFAMethods(_ context.Context, userID string) ([]models.UserMFAMethod, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var methods []models.UserMFAMethod
	for _, method := range s.mfaMethods {
		if method.UserID == userID {
			methods = append(methods, method)
		}
	}
	return methods, nil
}

func (s *workflowStore) UpdateMFAMethod(_ context.Context, id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	method, ok := s.mfaMethods[id]
	if !ok {
		return domainerrors.ErrMFAMethodNotFound
	}
	method.Enabled = enabled
	s.mfaMethods[id] = method
	return nil
}

func (s *workflowStore) CreateAuditLog(_ context.Context, log *models.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audits = append(s.audits, *log)
	return nil
}

func TestAuthenticationWorkflowRegistrationLoginRefreshReplayAndLogout(t *testing.T) {
	store := newWorkflowStore()
	hasher := NewHasher()
	provider, err := NewStaticSigningKeyProvider("workflow-key", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	tokens := NewTokenService(provider, 15*time.Minute, 5*time.Minute, 0)
	registration := NewRegistrationService(store, store, nil, hasher, store, nil)
	mfaKey := []byte("0123456789abcdef0123456789abcdef")
	auth := NewAuthService(store, store, store, store, store, tokens, hasher, mfaKey, time.Hour, time.Hour)
	mfa := NewMFAService(store, store, store, fixedTOTPGenerator{secret: "JBSWY3DPEHPK3PXP"}, mfaKey)

	registered, err := registration.RegisterIndividual(context.Background(), "person@example.com", "correct horse battery staple", "Person")
	if err != nil {
		t.Fatalf("RegisterIndividual() error = %v", err)
	}
	if registered.PasswordHash != "" {
		t.Fatal("registration response exposed password hash")
	}

	login, err := auth.Login(context.Background(), registered.TenantID, registered.Email, "correct horse battery staple")
	if err != nil || login.MFARequired || login.AccessToken == "" || login.RefreshToken == "" || login.SessionID == "" {
		t.Fatalf("Login() = %#v, %v", login, err)
	}
	claims, err := tokens.ValidateAccessToken(context.Background(), login.AccessToken)
	if err != nil || claims.UserID != registered.UserID || claims.TenantID != registered.TenantID {
		t.Fatalf("ValidateAccessToken() = %#v, %v", claims, err)
	}
	enrollment, err := mfa.EnrollTOTP(context.Background(), registered.UserID)
	if err != nil {
		t.Fatalf("EnrollTOTP() error = %v", err)
	}
	mfaCode, err := totp.GenerateCode(enrollment.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := mfa.ActivateTOTP(context.Background(), registered.UserID, enrollment.MFAID, mfaCode); err != nil {
		t.Fatalf("ActivateTOTP() error = %v", err)
	}
	if err := auth.Logout(context.Background(), login.SessionID); err != nil {
		t.Fatalf("initial Logout() error = %v", err)
	}

	challenge, err := auth.Login(context.Background(), registered.TenantID, registered.Email, "correct horse battery staple")
	if err != nil || !challenge.MFARequired || challenge.TempToken == "" || challenge.AccessToken != "" {
		t.Fatalf("MFA Login() = %#v, %v", challenge, err)
	}
	mfaCode, err = totp.GenerateCode(enrollment.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	login, err = auth.VerifyMFA(context.Background(), challenge.TempToken, mfaCode)
	if err != nil || login.MFARequired || login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatalf("VerifyMFA() = %#v, %v", login, err)
	}

	rotated, err := auth.RefreshToken(context.Background(), login.RefreshToken)
	if err != nil || rotated.RefreshToken == "" || rotated.RefreshToken == login.RefreshToken {
		t.Fatalf("RefreshToken() = %#v, %v", rotated, err)
	}
	if _, err := auth.RefreshToken(context.Background(), login.RefreshToken); !stderrors.Is(err, domainerrors.ErrInvalidToken) {
		t.Fatalf("replayed RefreshToken() error = %v, want ErrInvalidToken", err)
	}

	if err := auth.Logout(context.Background(), login.SessionID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := store.GetSessionByID(context.Background(), login.SessionID); !stderrors.Is(err, domainerrors.ErrSessionNotFound) {
		t.Fatalf("session survived logout: %v", err)
	}
	if _, err := auth.RefreshToken(context.Background(), rotated.RefreshToken); !stderrors.Is(err, domainerrors.ErrInvalidToken) {
		t.Fatalf("post-logout RefreshToken() error = %v, want ErrInvalidToken", err)
	}
}
