package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Auth     AuthConfig
	Security SecurityConfig
	Redis    RedisConfig
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
	JWTPrivateKey   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	TempTokenTTL    time.Duration
	SessionTTL      time.Duration
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
	return &Config{
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
		Auth: AuthConfig{
			JWTPrivateKey:   getEnv("JWT_PRIVATE_KEY", ""), // Empty will generate a key (not for production)
			AccessTokenTTL:  getEnvAsDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvAsDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
			TempTokenTTL:    getEnvAsDuration("TEMP_TOKEN_TTL", 10*time.Minute),
			SessionTTL:      getEnvAsDuration("SESSION_TTL", 24*time.Hour),
		},
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
