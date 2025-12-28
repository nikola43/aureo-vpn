// Package security provides centralized security configuration
package security

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// =============================================================================
// Security Configuration Errors
// =============================================================================

var (
	ErrMissingJWTSecret   = errors.New("JWT_SECRET environment variable is required")
	ErrInsecureProduction = errors.New("insecure configuration not allowed in production")
)

// =============================================================================
// Main Security Configuration
// =============================================================================

// Config holds all security-related configuration for the application
type Config struct {
	// Environment
	Environment string // "development", "staging", "production"
	Debug       bool

	// JWT settings
	JWT JWTConfig

	// Password settings
	Password PasswordConfig

	// Rate limiting
	RateLimit AppRateLimitConfig

	// Privacy settings
	Privacy PrivacyConfig

	// Database security
	Database DatabaseSecurityConfig

	// Resilience settings
	Resilience ResilienceConfig

	// Encryption
	Encryption EncryptionConfig

	// Session settings
	Session SessionConfig
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret            *SecureString
	Issuer            string
	Audience          []string
	AccessTokenLife   time.Duration
	RefreshTokenLife  time.Duration
	AllowedClockSkew  time.Duration
	EnableFingerprint bool
	EnableIPBinding   bool
}

// PasswordConfig holds password policy configuration
type PasswordConfig struct {
	MinLength         int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireNumber     bool
	RequireSpecial    bool
	BlockCommon       bool
	Argon2Memory      uint32
	Argon2Iterations  uint32
	Argon2Parallelism uint8
}

// AppRateLimitConfig holds application-level rate limiting configuration
type AppRateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	BurstSize         int
	CleanupInterval   time.Duration
	BlockDuration     time.Duration
}

// ResilienceConfig holds resilience pattern configuration
type ResilienceConfig struct {
	CircuitBreaker CircuitBreakerConfig
	Retry          RetryConfig
	Bulkhead       BulkheadConfig
	Timeout        time.Duration
}

// BulkheadConfig holds bulkhead configuration
type BulkheadConfig struct {
	MaxConcurrent int
	Timeout       time.Duration
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Algorithm    string // "aes-gcm", "chacha20-poly1305", "xchacha20-poly1305"
	KeySize      int    // 128 or 256 bits
	KeyRotation  bool
	GracePeriod  time.Duration
}

// SessionConfig holds session configuration
type SessionConfig struct {
	MaxActiveSessions     int
	InactivityTimeout     time.Duration
	AbsoluteTimeout       time.Duration
	ConcurrentLoginPolicy string // "allow", "deny", "force-logout"
}

// =============================================================================
// Configuration Loading
// =============================================================================

// DefaultConfig returns a secure default configuration
func DefaultConfig() *Config {
	return &Config{
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		Debug:       getEnvBool("DEBUG", false),

		JWT: JWTConfig{
			Issuer:            "aureo-vpn",
			Audience:          []string{"aureo-vpn-api"},
			AccessTokenLife:   15 * time.Minute,
			RefreshTokenLife:  7 * 24 * time.Hour,
			AllowedClockSkew:  30 * time.Second,
			EnableFingerprint: true,
			EnableIPBinding:   false,
		},

		Password: PasswordConfig{
			MinLength:         8,
			RequireUppercase:  true,
			RequireLowercase:  true,
			RequireNumber:     true,
			RequireSpecial:    true,
			BlockCommon:       true,
			Argon2Memory:      64 * 1024, // 64 MiB
			Argon2Iterations:  4,
			Argon2Parallelism: 4,
		},

		RateLimit: AppRateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 100,
			BurstSize:         200,
			CleanupInterval:   10 * time.Minute,
			BlockDuration:     15 * time.Minute,
		},

		Privacy: *DefaultPrivacyConfig(),

		Database: *DefaultDatabaseSecurityConfig(),

		Resilience: ResilienceConfig{
			CircuitBreaker: DefaultCircuitBreakerConfig(),
			Retry:          DefaultRetryConfig(),
			Bulkhead: BulkheadConfig{
				MaxConcurrent: 100,
				Timeout:       30 * time.Second,
			},
			Timeout: 30 * time.Second,
		},

		Encryption: EncryptionConfig{
			Algorithm:   "aes-gcm",
			KeySize:     256,
			KeyRotation: true,
			GracePeriod: 24 * time.Hour,
		},

		Session: SessionConfig{
			MaxActiveSessions:     5,
			InactivityTimeout:     30 * time.Minute,
			AbsoluteTimeout:       24 * time.Hour,
			ConcurrentLoginPolicy: "allow",
		},
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := DefaultConfig()

	// Load JWT secret (required)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// In production, this is required
		if config.Environment == "production" {
			return nil, ErrMissingJWTSecret
		}
		// For development, generate a random secret
		secretBytes, _ := SecureBytes(32)
		jwtSecret = SecureHashHex(secretBytes)
	}
	config.JWT.Secret = NewSecureString(jwtSecret)

	// Override from environment
	if v := os.Getenv("JWT_ACCESS_TOKEN_LIFE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.JWT.AccessTokenLife = d
		}
	}

	if v := os.Getenv("JWT_REFRESH_TOKEN_LIFE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.JWT.RefreshTokenLife = d
		}
	}

	if v := os.Getenv("JWT_ENABLE_IP_BINDING"); v != "" {
		config.JWT.EnableIPBinding = v == "true"
	}

	// Rate limit overrides
	if v := os.Getenv("RATE_LIMIT_ENABLED"); v != "" {
		config.RateLimit.Enabled = v == "true"
	}

	if v := os.Getenv("RATE_LIMIT_RPS"); v != "" {
		if rps, err := strconv.Atoi(v); err == nil && rps > 0 {
			config.RateLimit.RequestsPerSecond = rps
		}
	}

	// Database overrides
	if v := os.Getenv("DB_QUERY_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.Database.QueryTimeout = d
		}
	}

	// Validate production security requirements
	if config.Environment == "production" {
		if err := config.ValidateProduction(); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// ValidateProduction validates that the configuration meets production security requirements
func (c *Config) ValidateProduction() error {
	// JWT secret must be at least 32 bytes
	if c.JWT.Secret == nil || len(c.JWT.Secret.Bytes()) < 32 {
		return errors.New("JWT secret must be at least 32 bytes in production")
	}

	// Access token must be short-lived
	if c.JWT.AccessTokenLife > 1*time.Hour {
		return errors.New("access token lifetime must be <= 1 hour in production")
	}

	// Rate limiting must be enabled
	if !c.RateLimit.Enabled {
		return errors.New("rate limiting must be enabled in production")
	}

	// Debug must be disabled
	if c.Debug {
		return errors.New("debug mode must be disabled in production")
	}

	return nil
}

// =============================================================================
// Security Service Factory
// =============================================================================

// SecurityServices holds all security-related services
type SecurityServices struct {
	Config        *Config
	JWTService    *SecureJWTService
	Validator     *Validator
	PrivacyFilter *PrivacyFilter
	RateLimiter   *SecureRateLimiter
}

// NewSecurityServices creates all security services from configuration
func NewSecurityServices(config *Config) (*SecurityServices, error) {
	if config == nil {
		var err error
		config, err = LoadFromEnv()
		if err != nil {
			return nil, err
		}
	}

	// Create JWT service
	jwtConfig := &SecureJWTConfig{
		Secret:            config.JWT.Secret,
		Issuer:            config.JWT.Issuer,
		Audience:          config.JWT.Audience,
		AccessTokenLife:   config.JWT.AccessTokenLife,
		RefreshTokenLife:  config.JWT.RefreshTokenLife,
		AllowedClockSkew:  config.JWT.AllowedClockSkew,
		EnableFingerprint: config.JWT.EnableFingerprint,
		EnableIPBinding:   config.JWT.EnableIPBinding,
	}

	jwtService, err := NewSecureJWTService(jwtConfig)
	if err != nil {
		return nil, err
	}

	// Create validator (strict mode in production)
	validator := NewValidator(config.Environment == "production")

	// Create privacy filter
	privacyFilter := NewPrivacyFilter(&config.Privacy)

	// Create rate limiter
	var rateLimiter *SecureRateLimiter
	if config.RateLimit.Enabled {
		rateLimiter = NewSecureRateLimiter(&RateLimitConfig{
			MaxRequests:    config.RateLimit.RequestsPerSecond,
			WindowSize:     time.Second,
			BurstSize:      config.RateLimit.BurstSize,
			BlockDuration:  config.RateLimit.BlockDuration,
			PenaltyEnabled: true,
			PenaltyFactor:  2.0,
			MaxPenalty:     time.Hour,
			BlockThreshold: 10,
		})
	}

	return &SecurityServices{
		Config:        config,
		JWTService:    jwtService,
		Validator:     validator,
		PrivacyFilter: privacyFilter,
		RateLimiter:   rateLimiter,
	}, nil
}

// Close cleans up security services
func (s *SecurityServices) Close() {
	if s.Config.JWT.Secret != nil {
		s.Config.JWT.Secret.Wipe()
	}
	if s.RateLimiter != nil {
		s.RateLimiter.Stop()
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1"
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	return defaultValue
}

// =============================================================================
// Security Headers Configuration
// =============================================================================

// HTTPSecurityHeaders returns recommended security headers
func HTTPSecurityHeaders() map[string]string {
	return map[string]string{
		// Prevent clickjacking
		"X-Frame-Options": "DENY",

		// Prevent MIME type sniffing
		"X-Content-Type-Options": "nosniff",

		// Enable XSS filter
		"X-XSS-Protection": "1; mode=block",

		// Control referrer information
		"Referrer-Policy": "strict-origin-when-cross-origin",

		// Content Security Policy
		"Content-Security-Policy": "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'",

		// Strict Transport Security (HTTPS only)
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",

		// Permissions Policy (formerly Feature-Policy)
		"Permissions-Policy": "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()",

		// Cache control for sensitive data
		"Cache-Control": "no-store, no-cache, must-revalidate, proxy-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns secure CORS defaults
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{}, // Must be explicitly set
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}
