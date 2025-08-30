package middleware

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// LogConfig holds logging configuration
type LogConfig struct {
	// General logging settings
	Enabled bool
	Level   string

	// Structured logging options
	IncludeUserContext  bool
	IncludeRequestBody  bool
	IncludeResponseBody bool

	// Size limits for body logging
	MaxRequestBodySize  int
	MaxResponseBodySize int

	// Performance settings
	SamplingRate float64
	AsyncLogging bool

	// Security settings
	SanitizeHeaders     bool
	FilterSensitiveData bool
	AllowedHeaders      []string

	// Output settings
	TimeFormat string
	Pretty     bool
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Enabled: true,
		Level:   "info",

		IncludeUserContext:  true,
		IncludeRequestBody:  false,
		IncludeResponseBody: false,

		MaxRequestBodySize:  4096, // 4KB
		MaxResponseBodySize: 4096, // 4KB

		SamplingRate: 1.0, // Log all requests
		AsyncLogging: false,

		SanitizeHeaders:     true,
		FilterSensitiveData: true,
		AllowedHeaders: []string{
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"content-length",
			"content-type",
			"user-agent",
			"x-request-id",
		},

		TimeFormat: "2006-01-02T15:04:05Z07:00",
		Pretty:     false,
	}
}

// LoadLogConfig loads logging configuration from environment variables
func LoadLogConfig() *LogConfig {
	config := DefaultLogConfig()

	// General settings
	if enabled := os.Getenv("LOG_ENABLED"); enabled != "" {
		config.Enabled = strings.ToLower(enabled) == "true"
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = strings.ToLower(level)
	}

	// User context
	if includeUser := os.Getenv("LOG_INCLUDE_USER_CONTEXT"); includeUser != "" {
		config.IncludeUserContext = strings.ToLower(includeUser) == "true"
	}

	// Request/Response body logging
	if includeReqBody := os.Getenv("LOG_INCLUDE_REQUEST_BODY"); includeReqBody != "" {
		config.IncludeRequestBody = strings.ToLower(includeReqBody) == "true"
	}

	if includeRespBody := os.Getenv("LOG_INCLUDE_RESPONSE_BODY"); includeRespBody != "" {
		config.IncludeResponseBody = strings.ToLower(includeRespBody) == "true"
	}

	// Size limits
	if maxReqSize := os.Getenv("LOG_MAX_REQUEST_BODY_SIZE"); maxReqSize != "" {
		if size, err := strconv.Atoi(maxReqSize); err == nil && size > 0 {
			config.MaxRequestBodySize = size
		}
	}

	if maxRespSize := os.Getenv("LOG_MAX_RESPONSE_BODY_SIZE"); maxRespSize != "" {
		if size, err := strconv.Atoi(maxRespSize); err == nil && size > 0 {
			config.MaxResponseBodySize = size
		}
	}

	// Performance settings
	if samplingRate := os.Getenv("LOG_SAMPLING_RATE"); samplingRate != "" {
		if rate, err := strconv.ParseFloat(samplingRate, 64); err == nil && rate >= 0 && rate <= 1 {
			config.SamplingRate = rate
		}
	}

	if async := os.Getenv("LOG_ASYNC"); async != "" {
		config.AsyncLogging = strings.ToLower(async) == "true"
	}

	// Security settings
	if sanitize := os.Getenv("LOG_SANITIZE_HEADERS"); sanitize != "" {
		config.SanitizeHeaders = strings.ToLower(sanitize) == "true"
	}

	if filter := os.Getenv("LOG_FILTER_SENSITIVE_DATA"); filter != "" {
		config.FilterSensitiveData = strings.ToLower(filter) == "true"
	}

	if headers := os.Getenv("LOG_ALLOWED_HEADERS"); headers != "" {
		config.AllowedHeaders = strings.Split(headers, ",")
		for i, header := range config.AllowedHeaders {
			config.AllowedHeaders[i] = strings.TrimSpace(strings.ToLower(header))
		}
	}

	// Output settings
	if timeFormat := os.Getenv("LOG_TIME_FORMAT"); timeFormat != "" {
		config.TimeFormat = timeFormat
	}

	if pretty := os.Getenv("LOG_PRETTY"); pretty != "" {
		config.Pretty = strings.ToLower(pretty) == "true"
	}

	return config
}

// ShouldLogRequest determines if a request should be logged based on sampling rate
func (c *LogConfig) ShouldLogRequest() bool {
	if c.SamplingRate >= 1.0 {
		return true
	}
	if c.SamplingRate <= 0.0 {
		return false
	}
	// Simple random sampling - in production, you'd want a better sampling strategy
	return time.Now().UnixNano()%1000 < int64(c.SamplingRate*1000)
}

// IsHeaderAllowed checks if a header should be logged
func (c *LogConfig) IsHeaderAllowed(headerName string) bool {
	if !c.SanitizeHeaders {
		return true
	}

	lowerName := strings.ToLower(headerName)
	for _, allowed := range c.AllowedHeaders {
		if lowerName == allowed {
			return true
		}
	}
	return false
}
