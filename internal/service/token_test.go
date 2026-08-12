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

func TestSigningKeysSurviveProviderRestart(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(privateKey)
	beforeRestart, err := NewStaticSigningKeyProvider("active", encoded, nil)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := NewTokenService(beforeRestart, time.Minute, time.Minute, 0).GenerateTokenPair("user", "tenant", nil)
	if err != nil {
		t.Fatal(err)
	}

	afterRestart, err := NewStaticSigningKeyProvider("active", encoded, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTokenService(afterRestart, time.Minute, time.Minute, 0).ValidateAccessToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("restarted provider rejected existing token: %v", err)
	}
}

func TestSigningKeyProviderRejectsMalformedKeyMaterial(t *testing.T) {
	tests := []struct {
		name             string
		keyID            string
		privateKey       string
		verificationKeys map[string]string
	}{
		{name: "missing key ID", privateKey: base64.StdEncoding.EncodeToString(make([]byte, ed25519.PrivateKeySize))},
		{name: "malformed private key", keyID: "active", privateKey: "not-base64"},
		{name: "wrong private key length", keyID: "active", privateKey: base64.StdEncoding.EncodeToString([]byte("short"))},
		{name: "malformed retained public key", keyID: "active", privateKey: base64.StdEncoding.EncodeToString(make([]byte, ed25519.PrivateKeySize)), verificationKeys: map[string]string{"old": "not-base64"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewStaticSigningKeyProvider(test.keyID, test.privateKey, test.verificationKeys); err == nil {
				t.Fatal("NewStaticSigningKeyProvider() accepted unsafe key material")
			}
		})
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
