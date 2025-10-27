package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// JWT configuration
	JWT JWTConfig

	// Redis configuration
	Redis RedisConfig

	// Logging configuration
	Logging LoggingConfig

	// Security configuration
	Security SecurityConfig

	// Metrics configuration
	Metrics MetricsConfig

	// VPN configuration
	VPN VPNConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	BodyLimit       int
	TLS             TLSConfig
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	TimeZone        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	LogLevel        string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Enabled  bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level       string
	Format      string
	AddSource   bool
	Service     string
	Version     string
	Environment string
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	CORS               CORSConfig
	RateLimit          RateLimitConfig
	AllowedOrigins     []string
	TrustedProxies     []string
	PasswordMinLength  int
	MaxLoginAttempts   int
	LockoutDuration    time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	AllowCredentials bool
	MaxAge         int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool
	MaxRequests int
	WindowSize  time.Duration
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool
	Port    string
	Path    string
}

// VPNConfig holds VPN-specific configuration
type VPNConfig struct {
	NodeID                string
	DefaultProtocol       string
	SessionTimeout        time.Duration
	MaxSessionsPerUser    int
	DataTransferLimitGB   float64
	EnableKillSwitch      bool
	EnableDNSProtection   bool
	EnableMultiHop        bool
	EnableObfuscation     bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			Host:            getEnv("HOST", "0.0.0.0"),
			ReadTimeout:     getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getEnvAsDuration("IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
			BodyLimit:       getEnvAsInt("BODY_LIMIT", 10*1024*1024), // 10MB
			TLS: TLSConfig{
				Enabled:  getEnvAsBool("TLS_ENABLED", false),
				CertFile: getEnv("TLS_CERT_FILE", ""),
				KeyFile:  getEnv("TLS_KEY_FILE", ""),
			},
		},

		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "aureo_vpn"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			TimeZone:        getEnv("DB_TIMEZONE", "UTC"),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour),
			LogLevel:        getEnv("DB_LOG_LEVEL", "warn"),
		},

		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", ""),
			AccessTokenDuration:  getEnvAsDuration("JWT_ACCESS_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("JWT_REFRESH_DURATION", 7*24*time.Hour),
			Issuer:               getEnv("JWT_ISSUER", "aureo-vpn"),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnvAsBool("REDIS_ENABLED", false),
		},

		Logging: LoggingConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			Format:      getEnv("LOG_FORMAT", "json"),
			AddSource:   getEnvAsBool("LOG_ADD_SOURCE", true),
			Service:     getEnv("SERVICE_NAME", "aureo-vpn"),
			Version:     getEnv("VERSION", "1.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},

		Security: SecurityConfig{
			CORS: CORSConfig{
				Enabled:          getEnvAsBool("CORS_ENABLED", true),
				AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{}),
				AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
				AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
				MaxAge:           getEnvAsInt("CORS_MAX_AGE", 3600),
			},
			RateLimit: RateLimitConfig{
				Enabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
				MaxRequests: getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 100),
				WindowSize:  getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute),
			},
			AllowedOrigins:    getEnvAsSlice("ALLOWED_ORIGINS", []string{}),
			TrustedProxies:    getEnvAsSlice("TRUSTED_PROXIES", []string{}),
			PasswordMinLength: getEnvAsInt("PASSWORD_MIN_LENGTH", 8),
			MaxLoginAttempts:  getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:   getEnvAsDuration("LOCKOUT_DURATION", 15*time.Minute),
		},

		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "9090"),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},

		VPN: VPNConfig{
			NodeID:              getEnv("NODE_ID", ""),
			DefaultProtocol:     getEnv("DEFAULT_PROTOCOL", "wireguard"),
			SessionTimeout:      getEnvAsDuration("SESSION_TIMEOUT", 24*time.Hour),
			MaxSessionsPerUser:  getEnvAsInt("MAX_SESSIONS_PER_USER", 5),
			DataTransferLimitGB: getEnvAsFloat("DATA_TRANSFER_LIMIT_GB", 0), // 0 = unlimited
			EnableKillSwitch:    getEnvAsBool("ENABLE_KILL_SWITCH", true),
			EnableDNSProtection: getEnvAsBool("ENABLE_DNS_PROTECTION", true),
			EnableMultiHop:      getEnvAsBool("ENABLE_MULTIHOP", true),
			EnableObfuscation:   getEnvAsBool("ENABLE_OBFUSCATION", true),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate JWT secret in production
	if c.Logging.Environment == "production" {
		if c.JWT.Secret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
		}
		if c.Database.Password == "" {
			return fmt.Errorf("DB_PASSWORD is required in production")
		}
		if c.Server.TLS.Enabled && (c.Server.TLS.CertFile == "" || c.Server.TLS.KeyFile == "") {
			return fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE are required when TLS is enabled")
		}
		if len(c.Security.CORS.AllowedOrigins) == 0 {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must be set in production")
		}
	}

	// Validate database configuration
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Logging.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Logging.Environment == "production"
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Host, c.User, c.Password, c.DBName, c.Port, c.SSLMode, c.TimeZone,
	)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
