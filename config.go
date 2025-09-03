package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Port                string
	Environment         string
	LogLevel           string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	TokenExpiration    time.Duration
	RefreshTokenExpiration time.Duration
	RateLimitPerMinute int
	EnableMetrics      bool
	MetricsPort        string
	TLSCertFile        string
	TLSKeyFile         string
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port:                   getEnv("PORT", "8080"),
		Environment:           getEnv("ENVIRONMENT", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://localhost/token_manager?sslmode=disable"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:             getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		TokenExpiration:       getDurationEnv("TOKEN_EXPIRATION", 15*time.Minute),
		RefreshTokenExpiration: getDurationEnv("REFRESH_TOKEN_EXPIRATION", 24*time.Hour),
		RateLimitPerMinute:    getIntEnv("RATE_LIMIT_PER_MINUTE", 100),
		EnableMetrics:         getBoolEnv("ENABLE_METRICS", true),
		MetricsPort:           getEnv("METRICS_PORT", "9090"),
		TLSCertFile:           getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:            getEnv("TLS_KEY_FILE", ""),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getIntEnv gets an integer environment variable with a fallback value
func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getBoolEnv gets a boolean environment variable with a fallback value
func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

// getDurationEnv gets a duration environment variable with a fallback value
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
