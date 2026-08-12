package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func TestTokenServiceUsesConfiguredTTLAndClockSkew(t *testing.T) {
	provider := testSigningKeyProvider(t, "active", nil)
	service := NewTokenService(provider, 2*time.Minute, time.Minute, 30*time.Second).(*tokenService)
	now := time.Date(2026, time.August, 12, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	pair, err := service.GenerateTokenPair("user", "tenant", nil)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	service.now = func() time.Time { return now.Add(2*time.Minute + 29*time.Second) }
	if _, err := service.ValidateAccessToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("token inside clock-skew window was rejected: %v", err)
	}

	service.now = func() time.Time { return now.Add(2*time.Minute + 31*time.Second) }
	if _, err := service.ValidateAccessToken(context.Background(), pair.AccessToken); err == nil {
		t.Fatal("expired token was accepted beyond the clock-skew window")
	}
}

func TestTokenServicesShareKeysAcrossReplicas(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	encodedPrivateKey := base64.StdEncoding.EncodeToString(privateKey)
	issuerProvider, err := NewStaticSigningKeyProvider("active", encodedPrivateKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	verifierProvider, err := NewStaticSigningKeyProvider("active", encodedPrivateKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	issuer := NewTokenService(issuerProvider, time.Minute, time.Minute, 0)
	verifier := NewTokenService(verifierProvider, time.Minute, time.Minute, 0)

	pair, err := issuer.GenerateTokenPair("user", "tenant", nil)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	if _, err := verifier.ValidateAccessToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("replica rejected token signed by shared provider: %v", err)
	}
}

func TestSigningKeyRotationOverlapAndRevocation(t *testing.T) {
	oldPublic, oldPrivate, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	oldProvider, err := NewStaticSigningKeyProvider("old", base64.StdEncoding.EncodeToString(oldPrivate), nil)
	if err != nil {
		t.Fatal(err)
	}
	oldService := NewTokenService(oldProvider, time.Minute, time.Minute, 0)
	oldPair, err := oldService.GenerateTokenPair("user", "tenant", nil)
	if err != nil {
		t.Fatal(err)
	}

	overlapProvider := testSigningKeyProvider(t, "new", map[string]string{
		"old": base64.StdEncoding.EncodeToString(oldPublic),
	})
	overlapService := NewTokenService(overlapProvider, time.Minute, time.Minute, 0)
	if _, err := overlapService.ValidateAccessToken(context.Background(), oldPair.AccessToken); err != nil {
		t.Fatalf("old token was rejected during rotation overlap: %v", err)
	}

	revokedProvider := testSigningKeyProvider(t, "new", nil)
	revokedService := NewTokenService(revokedProvider, time.Minute, time.Minute, 0)
	if _, err := revokedService.ValidateAccessToken(context.Background(), oldPair.AccessToken); err == nil {
		t.Fatal("old token remained valid after its verification key was removed")
	}
}

func testSigningKeyProvider(t *testing.T, keyID string, verificationKeys map[string]string) SigningKeyProvider {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewStaticSigningKeyProvider(keyID, base64.StdEncoding.EncodeToString(privateKey), verificationKeys)
	if err != nil {
		t.Fatal(err)
	}
	return provider
}
