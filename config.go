package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server Configuration
	Port                   string
	Environment           string
	LogLevel              string
	
	// Database Configuration
	DatabaseURL           string
	RedisURL              string
	
	// JWT Configuration
	JWTSecret             string
	TokenExpiration       time.Duration
	RefreshTokenExpiration time.Duration
	
	// Rate Limiting
	RateLimitPerMinute    int
	
	// Monitoring
	EnableMetrics         bool
	MetricsPort           string
	
	// TLS Configuration
	TLSCertFile           string
	TLSKeyFile            string
	
	// GitHub App Configuration - NEW SECTION
	GitHubAppID           string
	GitHubInstallationID  string
	GitHubPrivateKeyPath  string
	GitHubTokenCacheTTL   time.Duration
	
	// Device Authentication - NEW SECTION
	DeviceAuthEnabled     bool
	DeviceValidationURL   string
	DeviceAuthTimeout     time.Duration
	
	// Container Registry Configuration - NEW SECTION
	RegistryURL           string
	RegistryUsername      string
}

// Load reads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		// Server Configuration
		Port:                   getEnv("PORT", "8080"),
		Environment:           getEnv("ENVIRONMENT", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		
		// Database Configuration
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://localhost/token_manager?sslmode=disable"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379"),
		
		// JWT Configuration
		JWTSecret:             getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		TokenExpiration:       getDurationEnv("TOKEN_EXPIRATION", 15*time.Minute),
		RefreshTokenExpiration: getDurationEnv("REFRESH_TOKEN_EXPIRATION", 24*time.Hour),
		
		// Rate Limiting
		RateLimitPerMinute:    getIntEnv("RATE_LIMIT_PER_MINUTE", 100),
		
		// Monitoring
		EnableMetrics:         getBoolEnv("ENABLE_METRICS", true),
		MetricsPort:           getEnv("METRICS_PORT", "9090"),
		
		// TLS Configuration
		TLSCertFile:           getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:            getEnv("TLS_KEY_FILE", ""),
		
		// GitHub App Configuration - NEW
		GitHubAppID:           getEnv("GITHUB_APP_ID", ""),
		GitHubInstallationID:  getEnv("GITHUB_INSTALLATION_ID", ""),
		GitHubPrivateKeyPath:  getEnv("GITHUB_PRIVATE_KEY_PATH", "/etc/secrets/github-app-private-key.pem"),
		GitHubTokenCacheTTL:   getDurationEnv("GITHUB_TOKEN_CACHE_TTL", 50*time.Minute), // GitHub tokens last ~60min
		
		// Device Authentication - NEW
		DeviceAuthEnabled:     getBoolEnv("DEVICE_AUTH_ENABLED", true),
		DeviceValidationURL:   getEnv("DEVICE_VALIDATION_URL", ""),
		DeviceAuthTimeout:     getDurationEnv("DEVICE_AUTH_TIMEOUT", 10*time.Second),
		
		// Container Registry Configuration - NEW
		RegistryURL:           getEnv("REGISTRY_URL", "ghcr.io"),
		RegistryUsername:      getEnv("REGISTRY_USERNAME", "ared-group"),
	}
}

// ValidateGitHubConfig checks if GitHub App configuration is properly set
func (c *Config) ValidateGitHubConfig() error {
	if c.GitHubAppID == "" {
		return fmt.Errorf("GITHUB_APP_ID is required")
	}
	if c.GitHubInstallationID == "" {
		return fmt.Errorf("GITHUB_INSTALLATION_ID is required")
	}
	if c.GitHubPrivateKeyPath == "" {
		return fmt.Errorf("GITHUB_PRIVATE_KEY_PATH is required")
	}
	
	// Check if private key file exists
	if _, err := os.Stat(c.GitHubPrivateKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("GitHub private key file not found at: %s", c.GitHubPrivateKeyPath)
	}
	
	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
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
