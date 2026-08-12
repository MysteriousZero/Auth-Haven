package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment string
	Database    DatabaseConfig
	Server      ServerConfig
	Auth        AuthConfig
	Security    SecurityConfig
	Redis       RedisConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ServerConfig struct {
	HTTPPort string
	GRPCPort string
	Host     string
}

type AuthConfig struct {
	JWTKeyID            string
	JWTPrivateKey       string
	JWTVerificationKeys map[string]string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	TempTokenTTL        time.Duration
	SessionTTL          time.Duration
	TokenClockSkew      time.Duration
}

type SecurityConfig struct {
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	MaxLoginAttempts       int
	AccountLockoutDuration time.Duration
	MFAEncryptionKey       string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func Load() (*Config, error) {
	auth, err := loadAuthConfig()
	if err != nil {
		return nil, err
	}

	environment := strings.ToLower(getEnv("APP_ENV", "development"))
	switch environment {
	case "development", "test":
	case "production":
		if auth.JWTPrivateKey == "" {
			return nil, fmt.Errorf("JWT_PRIVATE_KEY is required when APP_ENV=production")
		}
		if _, explicitlySet := os.LookupEnv("JWT_KEY_ID"); !explicitlySet || strings.TrimSpace(auth.JWTKeyID) == "" {
			return nil, fmt.Errorf("JWT_KEY_ID is required when APP_ENV=production")
		}
	default:
		return nil, fmt.Errorf("APP_ENV must be development, test, or production")
	}

	return &Config{
		Environment: environment,
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "auth_haven"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Server: ServerConfig{
			HTTPPort: getEnv("HTTP_PORT", ":8080"),
			GRPCPort: getEnv("GRPC_PORT", ":50051"),
			Host:     getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Auth: auth,
		Security: SecurityConfig{
			PasswordMinLength:      getEnvAsInt("PASSWORD_MIN_LENGTH", 12),
			PasswordRequireUpper:   getEnvAsBool("PASSWORD_REQUIRE_UPPER", true),
			PasswordRequireLower:   getEnvAsBool("PASSWORD_REQUIRE_LOWER", true),
			PasswordRequireNumber:  getEnvAsBool("PASSWORD_REQUIRE_NUMBER", true),
			PasswordRequireSpecial: getEnvAsBool("PASSWORD_REQUIRE_SPECIAL", true),
			MaxLoginAttempts:       getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			AccountLockoutDuration: getEnvAsDuration("ACCOUNT_LOCKOUT_DURATION", 15*time.Minute),
			MFAEncryptionKey:       getEnv("MFA_ENCRYPTION_KEY", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}, nil
}

func loadAuthConfig() (AuthConfig, error) {
	accessTTL, err := getRequiredPositiveDuration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return AuthConfig{}, err
	}
	refreshTTL, err := getRequiredPositiveDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return AuthConfig{}, err
	}
	tempTTL, err := getRequiredPositiveDuration("TEMP_TOKEN_TTL", 10*time.Minute)
	if err != nil {
		return AuthConfig{}, err
	}
	sessionTTL, err := getRequiredPositiveDuration("SESSION_TTL", 24*time.Hour)
	if err != nil {
		return AuthConfig{}, err
	}
	clockSkew, err := getNonNegativeDuration("TOKEN_CLOCK_SKEW", 30*time.Second)
	if err != nil {
		return AuthConfig{}, err
	}

	verificationKeys := make(map[string]string)
	if raw := getEnv("JWT_VERIFICATION_KEYS", ""); raw != "" {
		if err := json.Unmarshal([]byte(raw), &verificationKeys); err != nil {
			return AuthConfig{}, fmt.Errorf("JWT_VERIFICATION_KEYS must be a JSON object of key IDs to base64 public keys: %w", err)
		}
	}

	return AuthConfig{
		JWTKeyID:            getEnv("JWT_KEY_ID", "development"),
		JWTPrivateKey:       getEnv("JWT_PRIVATE_KEY", ""),
		JWTVerificationKeys: verificationKeys,
		AccessTokenTTL:      accessTTL,
		RefreshTokenTTL:     refreshTTL,
		TempTokenTTL:        tempTTL,
		SessionTTL:          sessionTTL,
		TokenClockSkew:      clockSkew,
	}, nil
}

func getRequiredPositiveDuration(key string, fallback time.Duration) (time.Duration, error) {
	duration, err := getValidatedDuration(key, fallback)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return duration, nil
}

func getNonNegativeDuration(key string, fallback time.Duration) (time.Duration, error) {
	duration, err := getValidatedDuration(key, fallback)
	if err != nil {
		return 0, err
	}
	if duration < 0 {
		return 0, fmt.Errorf("%s must not be negative", key)
	}
	return duration, nil
}

func getValidatedDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return duration, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return fallback
}
