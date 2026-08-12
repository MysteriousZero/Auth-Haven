package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainerrors "auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	pb "auth-haven/pkg/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authServiceStub struct {
	interfaces.AuthService
	loginResult *models.LoginResult
	loginErr    error
	refreshPair *models.TokenPair
	refreshErr  error
	refreshRaw  string
	logoutAllID string
}

func (s *authServiceStub) LogoutAll(_ context.Context, userID string) error {
	s.logoutAllID = userID
	return nil
}

type sessionServiceStub struct {
	interfaces.SessionService
	sessions      []models.Session
	listUserID    string
	revokeActorID string
	revokeID      string
}

func (s *sessionServiceStub) ListSessions(_ context.Context, userID string, _, _ int) ([]models.Session, error) {
	s.listUserID = userID
	return s.sessions, nil
}

func (s *sessionServiceStub) RevokeSession(_ context.Context, actorID, sessionID string) error {
	s.revokeActorID = actorID
	s.revokeID = sessionID
	return nil
}

func (s *authServiceStub) Login(context.Context, string, string, string) (*models.LoginResult, error) {
	return s.loginResult, s.loginErr
}

func (s *authServiceStub) RefreshToken(_ context.Context, raw string) (*models.TokenPair, error) {
	s.refreshRaw = raw
	return s.refreshPair, s.refreshErr
}

type passwordServiceStub struct {
	interfaces.PasswordService
	err   error
	calls int
}

func (s *passwordServiceStub) RequestPasswordReset(context.Context, string, string) error {
	s.calls++
	return s.err
}

func TestHTTPLoginContractDoesNotEnumerateAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := "70b6ff0a-3671-4e10-b78f-14f05fb9a169"
	payload := []byte(`{"tenant_id":"` + tenantID + `","email":"user@example.com","password":"correct horse battery"}`)

	for _, serviceErr := range []error{domainerrors.ErrInvalidCredentials, errors.New("database unavailable")} {
		service := &authServiceStub{loginErr: serviceErr}
		router := gin.New()
		router.POST("/login", NewAuthHandler(service).Login)
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		wantStatus := http.StatusUnauthorized
		if !errors.Is(serviceErr, domainerrors.ErrInvalidCredentials) {
			wantStatus = http.StatusInternalServerError
		}
		if response.Code != wantStatus {
			t.Fatalf("error %v produced status %d, want %d", serviceErr, response.Code, wantStatus)
		}
		var body map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if _, exists := body["password"]; exists {
			t.Fatalf("response leaked credentials: %s", response.Body.String())
		}
	}
}

func TestForgotPasswordAlwaysReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := "70b6ff0a-3671-4e10-b78f-14f05fb9a169"
	payload := []byte(`{"tenant_id":"` + tenantID + `","email":"unknown@example.com"}`)
	service := &passwordServiceStub{err: errors.New("unknown account")}
	router := gin.New()
	router.POST("/forgot", NewPasswordHandler(service).ForgotPassword)
	request := httptest.NewRequest(http.MethodPost, "/forgot", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || response.Body.Len() != 0 || service.calls != 1 {
		t.Fatalf("response = %d %q, calls = %d", response.Code, response.Body.String(), service.calls)
	}
}

func TestGRPCRefreshTokenContractAndErrorMapping(t *testing.T) {
	service := &authServiceStub{refreshPair: &models.TokenPair{AccessToken: "access", RefreshToken: "rotated"}}
	handler := NewGRPCHandler(service, nil, nil, nil, nil)
	response, err := handler.RefreshToken(context.Background(), &pb.RefreshTokenRequest{RefreshToken: "old"})
	if err != nil || response.GetAccessToken() != "access" || response.GetRefreshToken() != "rotated" || service.refreshRaw != "old" {
		t.Fatalf("RefreshToken() = %#v, %v; raw = %q", response, err, service.refreshRaw)
	}

	service.refreshErr = domainerrors.ErrInvalidToken
	_, err = handler.RefreshToken(context.Background(), &pb.RefreshTokenRequest{RefreshToken: "replayed"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("RefreshToken() status = %s, want %s", status.Code(err), codes.Unauthenticated)
	}
}

func TestGRPCSessionContractsEnforceAuthenticatedOwner(t *testing.T) {
	auth := &authServiceStub{}
	sessions := &sessionServiceStub{sessions: []models.Session{{SessionID: "session-a", UserID: "user-a"}}}
	handler := NewGRPCHandler(auth, nil, nil, sessions, nil)
	ctx := context.WithValue(context.Background(), "user-id", "user-a")

	_, err := handler.ListSessions(ctx, &pb.ListSessionsRequest{UserId: "user-b"})
	if status.Code(err) != codes.PermissionDenied || sessions.listUserID != "" {
		t.Fatalf("cross-owner ListSessions() status = %s, queried = %q", status.Code(err), sessions.listUserID)
	}
	response, err := handler.ListSessions(ctx, &pb.ListSessionsRequest{UserId: "user-a"})
	if err != nil || response.GetTotal() != 1 || sessions.listUserID != "user-a" {
		t.Fatalf("ListSessions() = %#v, %v; queried = %q", response, err, sessions.listUserID)
	}

	if _, err := handler.RevokeSession(ctx, &pb.RevokeSessionRequest{SessionId: "session-a"}); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}
	if sessions.revokeActorID != "user-a" || sessions.revokeID != "session-a" {
		t.Fatalf("revoke actor=%q session=%q", sessions.revokeActorID, sessions.revokeID)
	}

	_, err = handler.RevokeAllSessions(ctx, &pb.RevokeAllSessionsRequest{UserId: "user-b"})
	if status.Code(err) != codes.PermissionDenied || auth.logoutAllID != "" {
		t.Fatalf("cross-owner RevokeAllSessions() status = %s, logout user = %q", status.Code(err), auth.logoutAllID)
	}
	if _, err := handler.RevokeAllSessions(ctx, &pb.RevokeAllSessionsRequest{}); err != nil || auth.logoutAllID != "user-a" {
		t.Fatalf("RevokeAllSessions() error = %v, logout user = %q", err, auth.logoutAllID)
	}
}
