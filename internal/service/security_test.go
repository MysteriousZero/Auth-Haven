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

type tenantRepoStub struct {
	interfaces.TenantRepository
	getByID       func(context.Context, string) (*models.Tenant, error)
	getInvitation func(context.Context, string) (*models.Invitation, error)
}

func (s tenantRepoStub) GetTenantByID(ctx context.Context, id string) (*models.Tenant, error) {
	return s.getByID(ctx, id)
}

func (s tenantRepoStub) GetInvitationByHash(ctx context.Context, hash string) (*models.Invitation, error) {
	return s.getInvitation(ctx, hash)
}

type userRepoStub struct {
	interfaces.UserRepository
	getByID      func(context.Context, string) (*models.User, error)
	getByEmail   func(context.Context, string, string) (*models.User, error)
	updateLogin  func(context.Context, string) error
	updateUser   func(context.Context, *models.User) error
	updateStatus func(context.Context, string, models.UserStatus) error
	createUser   func(context.Context, *models.User) error
}

func (s userRepoStub) CreateUser(ctx context.Context, user *models.User) error {
	return s.createUser(ctx, user)
}

func (s userRepoStub) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.getByID(ctx, id)
}

func (s userRepoStub) GetUserByEmail(ctx context.Context, tenantID, email string) (*models.User, error) {
	return s.getByEmail(ctx, tenantID, email)
}

func (s userRepoStub) UpdateLastLogin(ctx context.Context, id string) error {
	return s.updateLogin(ctx, id)
}

func (s userRepoStub) UpdateUser(ctx context.Context, user *models.User) error {
	return s.updateUser(ctx, user)
}

func (s userRepoStub) UpdateUserStatus(ctx context.Context, id string, status models.UserStatus) error {
	return s.updateStatus(ctx, id, status)
}

type authRepoStub struct {
	interfaces.AuthRepository
	createSession func(context.Context, *models.Session) error
	getSession    func(context.Context, string) (*models.Session, error)
	revokeSession func(context.Context, string) error
	createRefresh func(context.Context, *models.RefreshToken) error
	getRefresh    func(context.Context, string) (*models.RefreshToken, error)
	revokeRefresh func(context.Context, string) error
}

func (s authRepoStub) CreateSession(ctx context.Context, session *models.Session) error {
	return s.createSession(ctx, session)
}

func (s authRepoStub) GetSessionByID(ctx context.Context, id string) (*models.Session, error) {
	return s.getSession(ctx, id)
}

func (s authRepoStub) RevokeSession(ctx context.Context, id string) error {
	return s.revokeSession(ctx, id)
}

func (s authRepoStub) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return s.createRefresh(ctx, token)
}

func (s authRepoStub) GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	return s.getRefresh(ctx, hash)
}

func (s authRepoStub) RevokeRefreshToken(ctx context.Context, id string) error {
	return s.revokeRefresh(ctx, id)
}

type mfaRepoStub struct {
	interfaces.MFAMethodRepository
	methods    []models.UserMFAMethod
	listFunc   func(context.Context, string) ([]models.UserMFAMethod, error)
	deleteFunc func(context.Context, string) error
}

func (s mfaRepoStub) DeleteMFAMethod(ctx context.Context, id string) error {
	return s.deleteFunc(ctx, id)
}

func (s mfaRepoStub) ListMFAMethods(ctx context.Context, userID string) ([]models.UserMFAMethod, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, userID)
	}
	return s.methods, nil
}

type auditRepoStub struct {
	interfaces.AuditRepository
	logs *[]models.AuditLog
}

func (s auditRepoStub) CreateAuditLog(_ context.Context, log *models.AuditLog) error {
	if s.logs != nil {
		*s.logs = append(*s.logs, *log)
	}
	return nil
}

type tokenServiceStub struct {
	interfaces.TokenService
	pair      *models.TokenPair
	pairErr   error
	tempToken string
}

func (s tokenServiceStub) GenerateTokenPair(string, string, *int64) (*models.TokenPair, error) {
	return s.pair, s.pairErr
}

func (s tokenServiceStub) GenerateTempToken(string) (string, error) { return s.tempToken, nil }

type hasherStub struct {
	interfaces.Hasher
	passwordMatches bool
}

func (s hasherStub) CompareHash(string, string) bool { return s.passwordMatches }
func (s hasherStub) HashToken(token string) string   { return "hash:" + token }

type roleRepoStub struct {
	interfaces.RoleRepository
	getRole     func(context.Context, int64) (*models.Role, error)
	permissions []models.RolePermission
}

func (s roleRepoStub) GetRoleByID(ctx context.Context, id int64) (*models.Role, error) {
	return s.getRole(ctx, id)
}

func (s roleRepoStub) GetRolePermissions(context.Context, int64) ([]models.RolePermission, error) {
	return s.permissions, nil
}

func TestAuthLoginCoversSuccessMFAAndAntiEnumeration(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	tenant := &models.Tenant{TenantID: "tenant-a", Status: models.TenantStatusActive}
	user := &models.User{UserID: "user-a", TenantID: tenant.TenantID, Email: "user@example.com", PasswordHash: "hash", Status: models.UserStatusActive}
	unknown := stderrors.New("not found")

	t.Run("tenant and account failures are indistinguishable", func(t *testing.T) {
		cases := []struct {
			name       string
			tenant     *models.Tenant
			tenantErr  error
			user       *models.User
			userErr    error
			passwordOK bool
		}{
			{name: "unknown tenant", tenantErr: unknown},
			{name: "suspended tenant", tenant: &models.Tenant{TenantID: "tenant-a", Status: models.TenantStatusSuspended}},
			{name: "unknown user", tenant: tenant, user: user, userErr: unknown},
			{name: "disabled user", tenant: tenant, user: &models.User{UserID: "user-a", TenantID: tenant.TenantID, Status: models.UserStatusDisabled}, passwordOK: true},
			{name: "wrong password", tenant: tenant, user: user},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				service := NewAuthService(
					tenantRepoStub{getByID: func(context.Context, string) (*models.Tenant, error) { return tc.tenant, tc.tenantErr }},
					userRepoStub{getByEmail: func(context.Context, string, string) (*models.User, error) { return tc.user, tc.userErr }},
					authRepoStub{}, mfaRepoStub{}, auditRepoStub{}, tokenServiceStub{},
					hasherStub{passwordMatches: tc.passwordOK}, nil, time.Hour, time.Hour,
				)
				_, err := service.Login(context.Background(), tenant.TenantID, user.Email, "wrong")
				if !stderrors.Is(err, domainerrors.ErrInvalidCredentials) {
					t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
				}
			})
		}
	})

	t.Run("enabled MFA returns only a challenge", func(t *testing.T) {
		service := NewAuthService(
			tenantRepoStub{getByID: func(context.Context, string) (*models.Tenant, error) { return tenant, nil }},
			userRepoStub{getByEmail: func(context.Context, string, string) (*models.User, error) { return user, nil }},
			authRepoStub{}, mfaRepoStub{methods: []models.UserMFAMethod{{Enabled: true}}}, auditRepoStub{},
			tokenServiceStub{tempToken: "challenge"}, hasherStub{passwordMatches: true}, nil, time.Hour, time.Hour,
		)
		result, err := service.Login(context.Background(), tenant.TenantID, user.Email, "correct")
		if err != nil || !result.MFARequired || result.TempToken != "challenge" || result.AccessToken != "" {
			t.Fatalf("Login() = %#v, %v; want isolated MFA challenge", result, err)
		}
	})

	t.Run("success persists bounded session and hashed refresh token", func(t *testing.T) {
		var session *models.Session
		var refresh *models.RefreshToken
		impl := NewAuthService(
			tenantRepoStub{getByID: func(context.Context, string) (*models.Tenant, error) { return tenant, nil }},
			userRepoStub{
				getByEmail:  func(context.Context, string, string) (*models.User, error) { return user, nil },
				updateLogin: func(context.Context, string) error { return nil },
			},
			authRepoStub{
				createSession: func(_ context.Context, got *models.Session) error { session = got; return nil },
				createRefresh: func(_ context.Context, got *models.RefreshToken) error { refresh = got; return nil },
			},
			mfaRepoStub{}, auditRepoStub{}, tokenServiceStub{pair: &models.TokenPair{AccessToken: "access", RefreshToken: "refresh"}},
			hasherStub{passwordMatches: true}, nil, 2*time.Hour, 30*time.Minute,
		)
		impl.(*authService).now = func() time.Time { return now }
		result, err := impl.Login(context.Background(), tenant.TenantID, user.Email, "correct")
		if err != nil || result.AccessToken != "access" || result.RefreshToken != "refresh" {
			t.Fatalf("Login() = %#v, %v", result, err)
		}
		if session == nil || !session.ExpiresAt.Equal(now.Add(30*time.Minute)) {
			t.Fatalf("session expiry = %#v", session)
		}
		if refresh == nil || refresh.TokenHash != "hash:refresh" || !refresh.ExpiresAt.Equal(now.Add(2*time.Hour)) {
			t.Fatalf("refresh token = %#v", refresh)
		}
	})
}

func TestRefreshTokenRejectsInvalidExpiredRevokedAndReplay(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	activeUser := &models.User{UserID: "user-a", TenantID: "tenant-a", Status: models.UserStatusActive}

	cases := []struct {
		name  string
		token *models.RefreshToken
		err   error
		want  error
	}{
		{name: "malformed or unknown", err: stderrors.New("missing"), want: domainerrors.ErrInvalidToken},
		{name: "revoked or replayed", token: &models.RefreshToken{Revoked: true, ExpiresAt: now.Add(time.Hour)}, want: domainerrors.ErrInvalidToken},
		{name: "expired", token: &models.RefreshToken{ExpiresAt: now.Add(-time.Second)}, want: domainerrors.ErrTokenExpired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			impl := NewAuthService(nil,
				userRepoStub{getByID: func(context.Context, string) (*models.User, error) { return activeUser, nil }},
				authRepoStub{getRefresh: func(context.Context, string) (*models.RefreshToken, error) { return tc.token, tc.err }},
				nil, nil, nil, hasherStub{}, nil, time.Hour, time.Hour,
			)
			impl.(*authService).now = func() time.Time { return now }
			_, err := impl.RefreshToken(context.Background(), "credential")
			if !stderrors.Is(err, tc.want) {
				t.Fatalf("RefreshToken() error = %v, want %v", err, tc.want)
			}
		})
	}

	revoked := false
	created := false
	impl := NewAuthService(nil,
		userRepoStub{getByID: func(context.Context, string) (*models.User, error) { return activeUser, nil }},
		authRepoStub{
			getRefresh: func(context.Context, string) (*models.RefreshToken, error) {
				return &models.RefreshToken{TokenID: "old", UserID: activeUser.UserID, ExpiresAt: now.Add(time.Hour)}, nil
			},
			revokeRefresh: func(context.Context, string) error { revoked = true; return nil },
			createRefresh: func(context.Context, *models.RefreshToken) error { created = true; return nil },
		}, nil, nil, tokenServiceStub{pair: &models.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}},
		hasherStub{}, nil, time.Hour, time.Hour,
	)
	impl.(*authService).now = func() time.Time { return now }
	pair, err := impl.RefreshToken(context.Background(), "old-refresh")
	if err != nil || pair.RefreshToken != "new-refresh" || !revoked || !created {
		t.Fatalf("RefreshToken() = %#v, %v; revoked=%v created=%v", pair, err, revoked, created)
	}
}

func TestTenantAndOwnerIsolation(t *testing.T) {
	t.Run("session cannot be revoked by another user", func(t *testing.T) {
		called := false
		service := NewSessionService(authRepoStub{
			getSession:    func(context.Context, string) (*models.Session, error) { return &models.Session{UserID: "owner"}, nil },
			revokeSession: func(context.Context, string) error { called = true; return nil },
		}, nil, nil, auditRepoStub{})
		err := service.RevokeSession(context.Background(), "attacker", "session")
		if !stderrors.Is(err, domainerrors.ErrForbidden) || called {
			t.Fatalf("RevokeSession() error = %v, repository called = %v", err, called)
		}
	})

	t.Run("role from another tenant cannot be assigned", func(t *testing.T) {
		roleID := int64(7)
		target := &models.User{UserID: "target", TenantID: "tenant-a", Status: models.UserStatusActive}
		actor := &models.User{UserID: "actor", TenantID: "tenant-a", RoleID: &roleID, Status: models.UserStatusActive}
		updated := false
		service := NewUserService(
			userRepoStub{
				getByID: func(_ context.Context, id string) (*models.User, error) {
					if id == actor.UserID {
						return actor, nil
					}
					return target, nil
				},
				updateUser: func(context.Context, *models.User) error { updated = true; return nil },
			},
			tenantRepoStub{getByID: func(context.Context, string) (*models.Tenant, error) {
				return &models.Tenant{Type: models.TenantTypeOrganization}, nil
			}},
			roleRepoStub{
				getRole:     func(context.Context, int64) (*models.Role, error) { return &models.Role{TenantID: "tenant-b"}, nil },
				permissions: []models.RolePermission{{PermissionKey: "users.write"}},
			}, auditRepoStub{},
		)
		err := service.AssignRole(context.Background(), actor.UserID, target.UserID, roleID)
		if !stderrors.Is(err, domainerrors.ErrForbidden) || updated {
			t.Fatalf("AssignRole() error = %v, updated = %v", err, updated)
		}
	})

	t.Run("cross-tenant audit access is forbidden before query", func(t *testing.T) {
		service := NewAuditService(auditRepoStub{}, userRepoStub{getByID: func(_ context.Context, id string) (*models.User, error) {
			return &models.User{UserID: id, TenantID: map[string]string{"actor": "tenant-a", "target": "tenant-b"}[id]}, nil
		}}, tenantRepoStub{})
		_, _, err := service.ListUserLogs(context.Background(), "actor", "target", "", 20)
		if !stderrors.Is(err, domainerrors.ErrForbidden) {
			t.Fatalf("ListUserLogs() error = %v, want ErrForbidden", err)
		}
	})
}

func TestInvitationMFAAndRoleAttackCases(t *testing.T) {
	t.Run("accepted and expired invitations cannot be replayed", func(t *testing.T) {
		for name, invitation := range map[string]*models.Invitation{
			"accepted": {Status: models.InvitationStatusAccepted, ExpiresAt: time.Now().Add(time.Hour)},
			"expired":  {Status: models.InvitationStatusPending, ExpiresAt: time.Now().Add(-time.Hour)},
		} {
			t.Run(name, func(t *testing.T) {
				created := false
				service := NewRegistrationService(
					tenantRepoStub{getInvitation: func(context.Context, string) (*models.Invitation, error) { return invitation, nil }},
					userRepoStub{createUser: func(context.Context, *models.User) error { created = true; return nil }},
					nil, hasherStub{}, auditRepoStub{}, nil,
				)
				_, err := service.RegisterWithInvitation(context.Background(), "token", "password", "Name")
				if !stderrors.Is(err, domainerrors.ErrInvitationExpiredOrUsed) || created {
					t.Fatalf("RegisterWithInvitation() error = %v, created = %v", err, created)
				}
			})
		}
	})

	t.Run("MFA method belonging to another user cannot be deleted", func(t *testing.T) {
		deleted := false
		service := NewMFAService(
			mfaRepoStub{
				listFunc:   func(context.Context, string) ([]models.UserMFAMethod, error) { return nil, nil },
				deleteFunc: func(context.Context, string) error { deleted = true; return nil },
			},
			nil, auditRepoStub{}, nil, nil,
		)
		err := service.DeleteMFAMethod(context.Background(), "attacker", "owned-method")
		if !stderrors.Is(err, domainerrors.ErrMFAMethodNotFound) || deleted {
			t.Fatalf("DeleteMFAMethod() error = %v, deleted = %v", err, deleted)
		}
	})

	t.Run("role from another tenant is hidden", func(t *testing.T) {
		roleID := int64(4)
		actor := &models.User{UserID: "actor", TenantID: "tenant-a", RoleID: &roleID}
		deleted := false
		repo := roleRepoStub{getRole: func(_ context.Context, id int64) (*models.Role, error) {
			if id == roleID {
				return &models.Role{RoleID: roleID, TenantID: "tenant-b", Name: "Admin"}, nil
			}
			return nil, domainerrors.ErrRoleNotFound
		}}
		service := NewRoleService(repo,
			userRepoStub{getByID: func(context.Context, string) (*models.User, error) { return actor, nil }},
			nil, auditRepoStub{},
		)
		err := service.DeleteRole(context.Background(), actor.UserID, roleID)
		if !stderrors.Is(err, domainerrors.ErrRoleNotFound) || deleted {
			t.Fatalf("DeleteRole() error = %v, deleted = %v", err, deleted)
		}
	})
}
