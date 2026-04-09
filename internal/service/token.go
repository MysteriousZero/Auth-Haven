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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

type tokenService struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewTokenService(privateKeyHex string) interfaces.TokenService {
	// For now, generate a new key pair if none provided
	// In production, this should be loaded from environment or secure storage
	var privateKey ed25519.PrivateKey

	if privateKeyHex == "" {
		_, privateKey, _ = ed25519.GenerateKey(rand.Reader)
	} else {
		// Parse hex private key
		privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyHex)
		if err != nil {
			panic(fmt.Sprintf("Failed to decode private key: %v", err))
		}
		if len(privateKeyBytes) != ed25519.PrivateKeySize {
			panic("Invalid private key size")
		}
		privateKey = ed25519.PrivateKey(privateKeyBytes)
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &tokenService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (s *tokenService) GenerateTokenPair(userID, tenantID string, roleID *int64) (*models.TokenPair, error) {
	now := time.Now().UTC()

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
		"exp":         now.Add(15 * time.Minute).Unix(), // 15 minutes
		"iss":         "auth-haven",
		"aud":         "auth-haven-client",
	})

	accessTokenString, err := accessToken.SignedString(s.privateKey)
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
	// Create a temporary token for MFA verification
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"sub":  userID,
		"iat":  time.Now().UTC().Unix(),
		"exp":  time.Now().UTC().Add(10 * time.Minute).Unix(), // 10 minutes
		"iss":  "auth-haven",
		"aud":  "auth-haven-mfa",
		"type": "temp",
	})

	return token.SignedString(s.privateKey)
}

func (s *tokenService) ValidateTempToken(tempToken string) (string, error) {
	parsedToken, err := jwt.Parse(tempToken, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

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

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", fmt.Errorf("temp token expired")
		}
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
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

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

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
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

