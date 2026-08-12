package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"

	"github.com/gin-gonic/gin"
)

type tokenServiceStub struct {
	interfaces.TokenService
	claims *models.Claims
	err    error
	token  string
}

func (s *tokenServiceStub) ValidateAccessToken(_ context.Context, token string) (*models.Claims, error) {
	s.token = token
	return s.claims, s.err
}

func TestAuthMiddlewareRejectsMissingMalformedAndInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		header string
		err    error
	}{
		{name: "missing"},
		{name: "wrong scheme", header: "Basic abc"},
		{name: "invalid", header: "Bearer bad", err: errors.New("expired")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokens := &tokenServiceStub{err: tc.err}
			router := gin.New()
			router.Use(AuthMiddleware(tokens))
			router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			var body map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if _, exists := body["details"]; exists {
				t.Fatalf("response exposes credential details: %s", response.Body.String())
			}
		})
	}
}

func TestAuthMiddlewareInjectsTrustedClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	roleID := int64(5)
	claims := &models.Claims{UserID: "user-a", TenantID: "tenant-a", TenantType: int16(models.TenantTypeOrganization), RoleID: &roleID, Email: "user@example.com"}
	tokens := &tokenServiceStub{claims: claims}
	router := gin.New()
	router.Use(AuthMiddleware(tokens))
	router.GET("/protected", func(c *gin.Context) {
		if c.GetString("user_id") != claims.UserID || c.GetString("tenant_id") != claims.TenantID {
			t.Fatalf("claims were not injected: user=%q tenant=%q", c.GetString("user_id"), c.GetString("tenant_id"))
		}
		if got, ok := c.Get("role_id"); !ok || got != claims.RoleID {
			t.Fatalf("role claim = %#v, %v", got, ok)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusNoContent || tokens.token != "valid" {
		t.Fatalf("status = %d, validated token = %q", response.Code, tokens.token)
	}
}
