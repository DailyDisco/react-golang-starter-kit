package config

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ConfigFile represents the YAML configuration file structure.
// All fields are optional and will be overridden by environment variables.
type ConfigFile struct {
	Server     *ServerConfigFile     `yaml:"server,omitempty"`
	Database   *DatabaseConfigFile   `yaml:"database,omitempty"`
	JWT        *JWTConfigFile        `yaml:"jwt,omitempty"`
	RateLimit  *RateLimitConfigFile  `yaml:"rate_limit,omitempty"`
	Logging    *LoggingConfigFile    `yaml:"logging,omitempty"`
	CORS       *CORSConfigFile       `yaml:"cors,omitempty"`
	FileUpload *FileUploadConfigFile `yaml:"file_upload,omitempty"`
	AWS        *AWSConfigFile        `yaml:"aws,omitempty"`
}

type ServerConfigFile struct {
	Port    string `yaml:"port,omitempty"`
	Host    string `yaml:"host,omitempty"`
	Env     string `yaml:"env,omitempty"`
	Debug   *bool  `yaml:"debug,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

type DatabaseConfigFile struct {
	Host    string `yaml:"host,omitempty"`
	Port    string `yaml:"port,omitempty"`
	User    string `yaml:"user,omitempty"`
	Name    string `yaml:"name,omitempty"`
	SSLMode string `yaml:"sslmode,omitempty"`
	// Password should be set via environment variable for security
}

type JWTConfigFile struct {
	ExpirationHours *int `yaml:"expiration_hours,omitempty"`
	// Secret should be set via environment variable for security
}

type RateLimitConfigFile struct {
	Enabled               *bool `yaml:"enabled,omitempty"`
	IPRequestsPerMinute   *int  `yaml:"ip_per_minute,omitempty"`
	IPRequestsPerHour     *int  `yaml:"ip_per_hour,omitempty"`
	IPBurstSize           *int  `yaml:"ip_burst,omitempty"`
	UserRequestsPerMinute *int  `yaml:"user_per_minute,omitempty"`
	UserRequestsPerHour   *int  `yaml:"user_per_hour,omitempty"`
	UserBurstSize         *int  `yaml:"user_burst,omitempty"`
	AuthRequestsPerMinute *int  `yaml:"auth_per_minute,omitempty"`
	AuthRequestsPerHour   *int  `yaml:"auth_per_hour,omitempty"`
	AuthBurstSize         *int  `yaml:"auth_burst,omitempty"`
	APIRequestsPerMinute  *int  `yaml:"api_per_minute,omitempty"`
	APIRequestsPerHour    *int  `yaml:"api_per_hour,omitempty"`
	APIBurstSize          *int  `yaml:"api_burst,omitempty"`
}

type LoggingConfigFile struct {
	Level               string   `yaml:"level,omitempty"`
	Pretty              *bool    `yaml:"pretty,omitempty"`
	Enabled             *bool    `yaml:"enabled,omitempty"`
	IncludeUserContext  *bool    `yaml:"include_user_context,omitempty"`
	IncludeRequestBody  *bool    `yaml:"include_request_body,omitempty"`
	IncludeResponseBody *bool    `yaml:"include_response_body,omitempty"`
	MaxRequestBodySize  *int     `yaml:"max_request_body_size,omitempty"`
	MaxResponseBodySize *int     `yaml:"max_response_body_size,omitempty"`
	SamplingRate        *float64 `yaml:"sampling_rate,omitempty"`
	AsyncLogging        *bool    `yaml:"async,omitempty"`
	SanitizeHeaders     *bool    `yaml:"sanitize_headers,omitempty"`
	FilterSensitiveData *bool    `yaml:"filter_sensitive,omitempty"`
	AllowedHeaders      []string `yaml:"allowed_headers,omitempty"`
}

type CORSConfigFile struct {
	AllowedOrigins []string `yaml:"allowed_origins,omitempty"`
}

type FileUploadConfigFile struct {
	MaxFileSizeMB *int64 `yaml:"max_file_size_mb,omitempty"`
}

type AWSConfigFile struct {
	Region   string `yaml:"region,omitempty"`
	S3Bucket string `yaml:"s3_bucket,omitempty"`
	// Credentials should be set via environment variables for security
}

// LoadFromFile loads configuration from a YAML file.
// Returns nil if the file doesn't exist (not an error).
func LoadFromFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist, not an error
		}
		return nil, err
	}

	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return nil, err
	}

	log.Info().Str("path", path).Msg("loaded configuration from YAML file")
	return &configFile, nil
}

// MergeWithConfig merges YAML file configuration into the main config.
// Environment variables (already applied) take precedence over file config.
func MergeWithConfig(config *Config, fileConfig *ConfigFile) {
	if fileConfig == nil {
		return
	}

	// Server
	if fc := fileConfig.Server; fc != nil {
		if fc.Port != "" && os.Getenv("API_PORT") == "" {
			config.Server.Port = fc.Port
		}
		if fc.Host != "" && os.Getenv("API_HOST") == "" {
			config.Server.Host = fc.Host
		}
		if fc.Env != "" && os.Getenv("GO_ENV") == "" {
			config.Server.Env = fc.Env
		}
		if fc.Debug != nil && os.Getenv("DEBUG") == "" {
			config.Server.Debug = *fc.Debug
		}
		if fc.Timeout != "" && os.Getenv("SERVER_TIMEOUT") == "" {
			if d, err := time.ParseDuration(fc.Timeout); err == nil {
				config.Server.Timeout = d
			}
		}
	}

	// Database (non-sensitive fields only)
	if fc := fileConfig.Database; fc != nil {
		if fc.Host != "" && os.Getenv("PGHOST") == "" && os.Getenv("DB_HOST") == "" {
			config.Database.Host = fc.Host
		}
		if fc.Port != "" && os.Getenv("PGPORT") == "" && os.Getenv("DB_PORT") == "" {
			config.Database.Port = fc.Port
		}
		if fc.User != "" && os.Getenv("PGUSER") == "" && os.Getenv("DB_USER") == "" {
			config.Database.User = fc.User
		}
		if fc.Name != "" && os.Getenv("PGDATABASE") == "" && os.Getenv("DB_NAME") == "" {
			config.Database.Name = fc.Name
		}
		if fc.SSLMode != "" && os.Getenv("DB_SSLMODE") == "" {
			config.Database.SSLMode = fc.SSLMode
		}
	}

	// JWT
	if fc := fileConfig.JWT; fc != nil {
		if fc.ExpirationHours != nil && os.Getenv("JWT_EXPIRATION_HOURS") == "" {
			config.JWT.ExpirationHours = *fc.ExpirationHours
		}
	}

	// Rate Limit
	if fc := fileConfig.RateLimit; fc != nil {
		if fc.Enabled != nil && os.Getenv("RATE_LIMIT_ENABLED") == "" {
			config.RateLimit.Enabled = *fc.Enabled
		}
		if fc.IPRequestsPerMinute != nil && os.Getenv("RATE_LIMIT_IP_PER_MINUTE") == "" {
			config.RateLimit.IPRequestsPerMinute = *fc.IPRequestsPerMinute
		}
		if fc.IPRequestsPerHour != nil && os.Getenv("RATE_LIMIT_IP_PER_HOUR") == "" {
			config.RateLimit.IPRequestsPerHour = *fc.IPRequestsPerHour
		}
		if fc.IPBurstSize != nil && os.Getenv("RATE_LIMIT_IP_BURST") == "" {
			config.RateLimit.IPBurstSize = *fc.IPBurstSize
		}
		if fc.UserRequestsPerMinute != nil && os.Getenv("RATE_LIMIT_USER_PER_MINUTE") == "" {
			config.RateLimit.UserRequestsPerMinute = *fc.UserRequestsPerMinute
		}
		if fc.UserRequestsPerHour != nil && os.Getenv("RATE_LIMIT_USER_PER_HOUR") == "" {
			config.RateLimit.UserRequestsPerHour = *fc.UserRequestsPerHour
		}
		if fc.UserBurstSize != nil && os.Getenv("RATE_LIMIT_USER_BURST") == "" {
			config.RateLimit.UserBurstSize = *fc.UserBurstSize
		}
		if fc.AuthRequestsPerMinute != nil && os.Getenv("RATE_LIMIT_AUTH_PER_MINUTE") == "" {
			config.RateLimit.AuthRequestsPerMinute = *fc.AuthRequestsPerMinute
		}
		if fc.AuthRequestsPerHour != nil && os.Getenv("RATE_LIMIT_AUTH_PER_HOUR") == "" {
			config.RateLimit.AuthRequestsPerHour = *fc.AuthRequestsPerHour
		}
		if fc.AuthBurstSize != nil && os.Getenv("RATE_LIMIT_AUTH_BURST") == "" {
			config.RateLimit.AuthBurstSize = *fc.AuthBurstSize
		}
		if fc.APIRequestsPerMinute != nil && os.Getenv("RATE_LIMIT_API_PER_MINUTE") == "" {
			config.RateLimit.APIRequestsPerMinute = *fc.APIRequestsPerMinute
		}
		if fc.APIRequestsPerHour != nil && os.Getenv("RATE_LIMIT_API_PER_HOUR") == "" {
			config.RateLimit.APIRequestsPerHour = *fc.APIRequestsPerHour
		}
		if fc.APIBurstSize != nil && os.Getenv("RATE_LIMIT_API_BURST") == "" {
			config.RateLimit.APIBurstSize = *fc.APIBurstSize
		}
	}

	// Logging
	if fc := fileConfig.Logging; fc != nil {
		if fc.Level != "" && os.Getenv("LOG_LEVEL") == "" {
			config.Logging.Level = fc.Level
		}
		if fc.Pretty != nil && os.Getenv("LOG_PRETTY") == "" {
			config.Logging.Pretty = *fc.Pretty
		}
		if fc.Enabled != nil && os.Getenv("LOG_ENABLED") == "" {
			config.Logging.Enabled = *fc.Enabled
		}
		if fc.IncludeUserContext != nil && os.Getenv("LOG_INCLUDE_USER_CONTEXT") == "" {
			config.Logging.IncludeUserContext = *fc.IncludeUserContext
		}
		if fc.IncludeRequestBody != nil && os.Getenv("LOG_INCLUDE_REQUEST_BODY") == "" {
			config.Logging.IncludeRequestBody = *fc.IncludeRequestBody
		}
		if fc.IncludeResponseBody != nil && os.Getenv("LOG_INCLUDE_RESPONSE_BODY") == "" {
			config.Logging.IncludeResponseBody = *fc.IncludeResponseBody
		}
		if fc.MaxRequestBodySize != nil && os.Getenv("LOG_MAX_REQUEST_BODY_SIZE") == "" {
			config.Logging.MaxRequestBodySize = *fc.MaxRequestBodySize
		}
		if fc.MaxResponseBodySize != nil && os.Getenv("LOG_MAX_RESPONSE_BODY_SIZE") == "" {
			config.Logging.MaxResponseBodySize = *fc.MaxResponseBodySize
		}
		if fc.SamplingRate != nil && os.Getenv("LOG_SAMPLING_RATE") == "" {
			config.Logging.SamplingRate = *fc.SamplingRate
		}
		if fc.AsyncLogging != nil && os.Getenv("LOG_ASYNC") == "" {
			config.Logging.AsyncLogging = *fc.AsyncLogging
		}
		if fc.SanitizeHeaders != nil && os.Getenv("LOG_SANITIZE_HEADERS") == "" {
			config.Logging.SanitizeHeaders = *fc.SanitizeHeaders
		}
		if fc.FilterSensitiveData != nil && os.Getenv("LOG_FILTER_SENSITIVE") == "" {
			config.Logging.FilterSensitiveData = *fc.FilterSensitiveData
		}
		if len(fc.AllowedHeaders) > 0 && os.Getenv("LOG_ALLOWED_HEADERS") == "" {
			config.Logging.AllowedHeaders = fc.AllowedHeaders
		}
	}

	// CORS
	if fc := fileConfig.CORS; fc != nil {
		if len(fc.AllowedOrigins) > 0 && os.Getenv("CORS_ALLOWED_ORIGINS") == "" {
			config.CORS.AllowedOrigins = fc.AllowedOrigins
		}
	}

	// File Upload
	if fc := fileConfig.FileUpload; fc != nil {
		if fc.MaxFileSizeMB != nil && os.Getenv("MAX_FILE_SIZE_MB") == "" {
			config.FileUpload.MaxFileSizeMB = *fc.MaxFileSizeMB
		}
	}

	// AWS (non-sensitive fields only)
	if fc := fileConfig.AWS; fc != nil {
		if fc.Region != "" && os.Getenv("AWS_REGION") == "" {
			config.AWS.Region = fc.Region
		}
		if fc.S3Bucket != "" && os.Getenv("AWS_S3_BUCKET") == "" {
			config.AWS.S3Bucket = fc.S3Bucket
		}
	}
}

// LoadWithFile loads configuration with support for an optional YAML config file.
// Priority: Environment Variables > config.yaml > defaults
func LoadWithFile(configPath string) (*Config, error) {
	// First, load with defaults and environment variables
	config, err := Load()
	if err != nil {
		return nil, err
	}

	// Try to load from YAML file
	if configPath == "" {
		configPath = "config.yaml"
	}

	fileConfig, err := LoadFromFile(configPath)
	if err != nil {
		log.Warn().Err(err).Str("path", configPath).Msg("failed to load config file, using env vars only")
		return config, nil
	}

	// Merge file config (env vars already take precedence in Load())
	MergeWithConfig(config, fileConfig)

	// Re-validate after merge
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}
