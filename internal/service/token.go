package service

import (
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

type tokenService struct {
	keyProvider SigningKeyProvider
	accessTTL   time.Duration
	tempTTL     time.Duration
	clockSkew   time.Duration
	now         func() time.Time
}

// SigningKeyProvider is the secret-manager/KMS boundary used by token signing.
// Providers return an active private key for signing and public keys retained
// during a rotation overlap window for verification.
type SigningKeyProvider interface {
	ActiveKeyID() string
	SignJWT(token *jwt.Token) (string, error)
	VerificationKey(keyID string) (ed25519.PublicKey, error)
}

type staticSigningKeyProvider struct {
	activeKeyID string
	privateKey  ed25519.PrivateKey
	publicKeys  map[string]ed25519.PublicKey
}

// NewStaticSigningKeyProvider builds the environment-backed provider. An empty
// private key is allowed only for local development and produces an ephemeral key.
func NewStaticSigningKeyProvider(activeKeyID, privateKeyBase64 string, verificationKeys map[string]string) (SigningKeyProvider, error) {
	if strings.TrimSpace(activeKeyID) == "" {
		return nil, fmt.Errorf("signing key ID must not be empty")
	}

	var privateKey ed25519.PrivateKey
	if privateKeyBase64 == "" {
		_, generated, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate development signing key: %w", err)
		}
		privateKey = generated
	} else {
		decoded, err := base64.StdEncoding.DecodeString(privateKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("decode signing private key: %w", err)
		}
		if len(decoded) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("signing private key must be %d bytes", ed25519.PrivateKeySize)
		}
		privateKey = ed25519.PrivateKey(decoded)
	}

	publicKeys := make(map[string]ed25519.PublicKey, len(verificationKeys)+1)
	for keyID, encoded := range verificationKeys {
		if keyID == "" {
			return nil, fmt.Errorf("verification key ID must not be empty")
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode verification key %q: %w", keyID, err)
		}
		if len(decoded) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("verification key %q must be %d bytes", keyID, ed25519.PublicKeySize)
		}
		publicKeys[keyID] = ed25519.PublicKey(decoded)
	}
	publicKeys[activeKeyID] = privateKey.Public().(ed25519.PublicKey)

	return &staticSigningKeyProvider{activeKeyID: activeKeyID, privateKey: privateKey, publicKeys: publicKeys}, nil
}

func (p *staticSigningKeyProvider) ActiveKeyID() string {
	return p.activeKeyID
}

func (p *staticSigningKeyProvider) SignJWT(token *jwt.Token) (string, error) {
	return token.SignedString(p.privateKey)
}

func (p *staticSigningKeyProvider) VerificationKey(keyID string) (ed25519.PublicKey, error) {
	key, ok := p.publicKeys[keyID]
	if !ok {
		return nil, fmt.Errorf("unknown signing key ID")
	}
	return key, nil
}

func NewTokenService(provider SigningKeyProvider, accessTTL, tempTTL, clockSkew time.Duration) interfaces.TokenService {
	return &tokenService{keyProvider: provider, accessTTL: accessTTL, tempTTL: tempTTL, clockSkew: clockSkew, now: time.Now}
}

func (s *tokenService) GenerateTokenPair(userID, tenantID string, roleID *int64) (*models.TokenPair, error) {
	now := s.now().UTC()

	// Get tenant type (would need to fetch from DB, for now assume 1)
	tenantType := int16(1)

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"sub":         userID,
		"tenant_id":   tenantID,
		"tenant_type": tenantType,
		"role_id":     roleID,
		"email":       "", // Would need to fetch from user
		"iat":         now.Unix(),
		"exp":         now.Add(s.accessTTL).Unix(),
		"iss":         "auth-haven",
		"aud":         "auth-haven-client",
	})

	accessToken.Header["kid"] = s.keyProvider.ActiveKeyID()
	accessTokenString, err := s.keyProvider.SignJWT(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
	}, nil
}

func (s *tokenService) GenerateTempToken(userID string) (string, error) {
	now := s.now().UTC()
	// Create a temporary token for MFA verification
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"sub":  userID,
		"iat":  now.Unix(),
		"exp":  now.Add(s.tempTTL).Unix(),
		"iss":  "auth-haven",
		"aud":  "auth-haven-mfa",
		"type": "temp",
	})

	token.Header["kid"] = s.keyProvider.ActiveKeyID()
	return s.keyProvider.SignJWT(token)
}

func (s *tokenService) ValidateTempToken(tempToken string) (string, error) {
	parsedToken, err := jwt.Parse(tempToken, s.verificationKey,
		jwt.WithLeeway(s.clockSkew), jwt.WithTimeFunc(s.now), jwt.WithAudience("auth-haven-mfa"),
		jwt.WithIssuer("auth-haven"), jwt.WithExpirationRequired())

	if err != nil {
		return "", fmt.Errorf("invalid temp token")
	}

	if !parsedToken.Valid {
		return "", fmt.Errorf("invalid temp token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Check if it's a temp token
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "temp" {
		return "", fmt.Errorf("not a temp token")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID in token")
	}

	return userID, nil
}

func (s *tokenService) GenerateResetToken() string {
	token, _ := s.generateSecureToken(32)
	return token
}

func (s *tokenService) GenerateInvitationToken() string {
	token, _ := s.generateSecureToken(32)
	return token
}

func (s *tokenService) generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

type hasher struct {
	// Argon2 parameters
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen uint32
}

func NewHasher() interfaces.Hasher {
	return &hasher{
		time:    1,
		memory:  64 * 1024, // 64MB
		threads: 4,
		keyLen:  32,
		saltLen: 16,
	}
}

func (h *hasher) HashPassword(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	// Combine salt and hash
	fullHash := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(fullHash), nil
}

func (h *hasher) CompareHash(password, hash string) bool {
	fullHash, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}

	if len(fullHash) < int(h.saltLen) {
		return false
	}

	salt := fullHash[:h.saltLen]
	expectedHash := fullHash[h.saltLen:]

	actualHash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	// Constant-time comparison
	if len(expectedHash) != len(actualHash) {
		return false
	}

	var result byte
	for i := 0; i < len(expectedHash); i++ {
		result |= expectedHash[i] ^ actualHash[i]
	}

	return result == 0
}

func (h *hasher) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, token string) (*models.Claims, error) {
	// Parse JWT token
	parsedToken, err := jwt.Parse(token, s.verificationKey,
		jwt.WithLeeway(s.clockSkew), jwt.WithTimeFunc(s.now), jwt.WithAudience("auth-haven-client"),
		jwt.WithIssuer("auth-haven"), jwt.WithExpirationRequired())

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Build claims object
	userClaims := &models.Claims{
		UserID:   claims["sub"].(string),
		TenantID: claims["tenant_id"].(string),
		Email:    claims["email"].(string),
		Issuer:   claims["iss"].(string),
		Audience: claims["aud"].(string),
	}

	if roleID, ok := claims["role_id"].(float64); ok {
		roleIDInt := int64(roleID)
		userClaims.RoleID = &roleIDInt
	}

	if tenantType, ok := claims["tenant_type"].(float64); ok {
		userClaims.TenantType = int16(tenantType)
	}

	return userClaims, nil
}

func (s *tokenService) verificationKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	keyID, ok := token.Header["kid"].(string)
	if !ok || keyID == "" {
		return nil, fmt.Errorf("missing signing key ID")
	}
	return s.keyProvider.VerificationKey(keyID)
}
