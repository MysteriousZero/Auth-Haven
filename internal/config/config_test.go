package config

import "testing"

func TestProductionRequiresSigningKey(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_KEY_ID", "2026-08-primary")
	t.Setenv("JWT_PRIVATE_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted production configuration without a signing key")
	}
}

func TestProductionRequiresExplicitKeyID(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_PRIVATE_KEY", "configured")
	t.Setenv("JWT_KEY_ID", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted production configuration without an explicit key ID")
	}
}

func TestAuthDurationsFailClosed(t *testing.T) {
	t.Setenv("ACCESS_TOKEN_TTL", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an invalid access-token TTL")
	}
}
