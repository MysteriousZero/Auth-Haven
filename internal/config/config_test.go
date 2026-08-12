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

func TestProductionRequiresExplicitCORSOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_KEY_ID", "2026-08-primary")
	t.Setenv("JWT_PRIVATE_KEY", "configured")
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("EMAIL_FROM", "auth@example.com")
	t.Setenv("PUBLIC_BASE_URL", "https://auth.example.com")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a wildcard production CORS origin")
	}
}

func TestOperationalConfiguration(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("REDIS_RATE_LIMIT_FAIL_OPEN", "false")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("EMAIL_FROM", "auth@example.com")
	t.Setenv("PUBLIC_BASE_URL", "https://auth.example.com/")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("allowed origins = %v", cfg.CORS.AllowedOrigins)
	}
	if !cfg.CORS.AllowCredentials {
		t.Fatal("credentials policy was not loaded")
	}
	if cfg.Redis.RateLimitFailOpen {
		t.Fatal("Redis failure policies were not loaded")
	}
	if cfg.Email.SMTPPort != 2525 || cfg.Email.BaseURL != "https://auth.example.com" {
		t.Fatalf("email config = %+v", cfg.Email)
	}
}
