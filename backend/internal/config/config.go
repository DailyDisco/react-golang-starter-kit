package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// DefaultAllowedOrigins contains the default CORS allowed origins for development.
// This is the single source of truth for CORS origins across all middleware.
var DefaultAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:3001",
	"http://localhost:3002",
	"http://localhost:5173",
	"http://localhost:5174",
	"http://localhost:5175",
	"http://localhost:5193",
	"http://localhost:8080",
	"http://localhost:8081",
	"http://localhost:8082",
}

// GetAllowedOrigins returns CORS allowed origins from environment or defaults.
// Use this function in all CORS-related middleware to ensure consistency.
func GetAllowedOrigins() []string {
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		return strings.Split(origins, ",")
	}
	return DefaultAllowedOrigins
}

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// JWT configuration
	JWT JWTConfig

	// Rate limiting configuration
	RateLimit RateLimitConfig

	// Logging configuration
	Logging LoggingConfig

	// CORS configuration
	CORS CORSConfig

	// File upload configuration
	FileUpload FileUploadConfig

	// AWS S3 configuration (optional)
	AWS AWSConfig
}

type ServerConfig struct {
	Port    string
	Host    string
	Env     string
	Debug   bool
	Timeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type RateLimitConfig struct {
	Enabled               bool
	IPRequestsPerMinute   int
	IPRequestsPerHour     int
	IPBurstSize           int
	UserRequestsPerMinute int
	UserRequestsPerHour   int
	UserBurstSize         int
	AuthRequestsPerMinute int
	AuthRequestsPerHour   int
	AuthBurstSize         int
	APIRequestsPerMinute  int
	APIRequestsPerHour    int
	APIBurstSize          int
}

type LoggingConfig struct {
	Level               string
	Pretty              bool
	Enabled             bool
	IncludeUserContext  bool
	IncludeRequestBody  bool
	IncludeResponseBody bool
	MaxRequestBodySize  int
	MaxResponseBodySize int
	SamplingRate        float64
	AsyncLogging        bool
	SanitizeHeaders     bool
	FilterSensitiveData bool
	AllowedHeaders      []string
	TimeFormat          string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type FileUploadConfig struct {
	MaxFileSizeMB int64
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	S3Bucket        string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (optional)
	// godotenv.Load looks in the current directory and parent directories
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using system environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("API_PORT", "8080"),
			Host:    getEnv("API_HOST", "0.0.0.0"),
			Env:     getEnv("GO_ENV", "development"),
			Debug:   getEnvBool("DEBUG", false),
			Timeout: getEnvDuration("SERVER_TIMEOUT", 30*time.Second),
		},

		Database: DatabaseConfig{
			Host:     getEnv("PGHOST", getEnv("DB_HOST", "localhost")),
			Port:     getEnv("PGPORT", getEnv("DB_PORT", "5432")),
			User:     getEnv("PGUSER", getEnv("DB_USER", "devuser")),
			Password: getEnv("PGPASSWORD", getEnv("DB_PASSWORD", "devpass")),
			Name:     getEnv("PGDATABASE", getEnv("DB_NAME", "starter_kit_db")),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		},

		RateLimit: RateLimitConfig{
			Enabled:               getEnvBool("RATE_LIMIT_ENABLED", true),
			IPRequestsPerMinute:   getEnvInt("RATE_LIMIT_IP_PER_MINUTE", 60),
			IPRequestsPerHour:     getEnvInt("RATE_LIMIT_IP_PER_HOUR", 1000),
			IPBurstSize:           getEnvInt("RATE_LIMIT_IP_BURST_SIZE", 10),
			UserRequestsPerMinute: getEnvInt("RATE_LIMIT_USER_PER_MINUTE", 100),
			UserRequestsPerHour:   getEnvInt("RATE_LIMIT_USER_PER_HOUR", 2000),
			UserBurstSize:         getEnvInt("RATE_LIMIT_USER_BURST_SIZE", 20),
			AuthRequestsPerMinute: getEnvInt("RATE_LIMIT_AUTH_PER_MINUTE", 5),
			AuthRequestsPerHour:   getEnvInt("RATE_LIMIT_AUTH_PER_HOUR", 100),
			AuthBurstSize:         getEnvInt("RATE_LIMIT_AUTH_BURST_SIZE", 2),
			APIRequestsPerMinute:  getEnvInt("RATE_LIMIT_API_PER_MINUTE", 100),
			APIRequestsPerHour:    getEnvInt("RATE_LIMIT_API_PER_HOUR", 2000),
			APIBurstSize:          getEnvInt("RATE_LIMIT_API_BURST_SIZE", 20),
		},

		Logging: LoggingConfig{
			Level:               getEnv("LOG_LEVEL", "info"),
			Pretty:              getEnvBool("LOG_PRETTY", false),
			Enabled:             getEnvBool("LOG_ENABLED", true),
			IncludeUserContext:  getEnvBool("LOG_INCLUDE_USER_CONTEXT", false),
			IncludeRequestBody:  getEnvBool("LOG_INCLUDE_REQUEST_BODY", false),
			IncludeResponseBody: getEnvBool("LOG_INCLUDE_RESPONSE_BODY", false),
			MaxRequestBodySize:  getEnvInt("LOG_MAX_REQUEST_BODY_SIZE", 1024),
			MaxResponseBodySize: getEnvInt("LOG_MAX_RESPONSE_BODY_SIZE", 1024),
			SamplingRate:        getEnvFloat64("LOG_SAMPLING_RATE", 1.0),
			AsyncLogging:        getEnvBool("LOG_ASYNC", false),
			SanitizeHeaders:     getEnvBool("LOG_SANITIZE_HEADERS", true),
			FilterSensitiveData: getEnvBool("LOG_FILTER_SENSITIVE", true),
			AllowedHeaders:      getEnvSlice("LOG_ALLOWED_HEADERS", []string{"Authorization", "Content-Type"}),
			TimeFormat:          getEnv("LOG_TIME_FORMAT", time.RFC3339),
		},

		CORS: CORSConfig{
			AllowedOrigins: GetAllowedOrigins(),
		},

		FileUpload: FileUploadConfig{
			MaxFileSizeMB: getEnvInt64("MAX_FILE_SIZE_MB", 10),
		},

		AWS: AWSConfig{
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:          getEnv("AWS_REGION", ""),
			S3Bucket:        getEnv("AWS_S3_BUCKET", ""),
		},
	}

	// Validate critical configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate ensures required configuration is present
func (c *Config) Validate() error {
	var errs []string

	// JWT secret is required
	if c.JWT.Secret == "" {
		errs = append(errs, "JWT_SECRET is required")
	}

	// Database configuration validation
	if c.Database.Host == "" {
		errs = append(errs, "database host is required")
	}
	if c.Database.Name == "" {
		errs = append(errs, "database name is required")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	validLevel := false
	for _, level := range validLogLevels {
		if strings.ToLower(c.Logging.Level) == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		errs = append(errs, fmt.Sprintf("invalid log level: %s, must be one of: %v", c.Logging.Level, validLogLevels))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			log.Warn().Str("key", key).Str("value", value).Msg("invalid boolean value, using fallback")
			return fallback
		}
		return parsed
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Warn().Str("key", key).Str("value", value).Msg("invalid integer value, using fallback")
			return fallback
		}
		return parsed
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Warn().Str("key", key).Str("value", value).Msg("invalid int64 value, using fallback")
			return fallback
		}
		return parsed
	}
	return fallback
}

func getEnvFloat64(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Warn().Str("key", key).Str("value", value).Msg("invalid float64 value, using fallback")
			return fallback
		}
		return parsed
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			log.Warn().Str("key", key).Str("value", value).Msg("invalid duration value, using fallback")
			return fallback
		}
		return parsed
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return fallback
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Server.Env) == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Server.Env) == "production"
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}
