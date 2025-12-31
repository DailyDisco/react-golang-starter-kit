package ai

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// SafetyLevel represents the content filtering level
type SafetyLevel string

const (
	SafetyLevelNone     SafetyLevel = "none"      // No filtering (not recommended)
	SafetyLevelLow      SafetyLevel = "low"       // Block only high-probability harmful content
	SafetyLevelMedium   SafetyLevel = "medium"    // Block medium and high probability (default)
	SafetyLevelHigh     SafetyLevel = "high"      // Block low, medium, and high probability
	SafetyLevelBlockAll SafetyLevel = "block_all" // Block all potentially harmful content
)

// Config holds AI service configuration
type Config struct {
	APIKey         string
	Model          string
	EmbeddingModel string
	MaxTokens      int
	Timeout        time.Duration
	Enabled        bool
	MaxImageSize   int64 // Max image size in bytes

	// Safety settings
	SafetyLevel SafetyLevel // Content filtering level

	// Input validation limits
	MaxPromptLength    int // Max characters per prompt
	MaxMessagesPerChat int // Max messages in a single chat request
	MaxTextsPerEmbed   int // Max texts in a single embedding request

	// Feature flags
	AllowFunctionCalling bool // Enable function calling / tools
	AllowJSONMode        bool // Enable structured JSON output
}

// DefaultConfig returns the default AI configuration
func DefaultConfig() *Config {
	return &Config{
		APIKey:         "",
		Model:          "gemini-2.0-flash",
		EmbeddingModel: "text-embedding-004",
		MaxTokens:      8192,
		Timeout:        30 * time.Second,
		Enabled:        false,
		MaxImageSize:   10 * 1024 * 1024, // 10MB

		// Safety defaults
		SafetyLevel: SafetyLevelMedium,

		// Input validation defaults
		MaxPromptLength:    100000, // ~25k tokens
		MaxMessagesPerChat: 50,     // Reasonable chat history
		MaxTextsPerEmbed:   100,    // Batch embedding limit

		// Feature flags - enabled by default
		AllowFunctionCalling: true,
		AllowJSONMode:        true,
	}
}

// LoadConfig loads AI configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	if val := os.Getenv("GEMINI_API_KEY"); val != "" {
		config.APIKey = val
	}
	if val := os.Getenv("GEMINI_MODEL"); val != "" {
		config.Model = val
	}
	if val := os.Getenv("GEMINI_EMBEDDING_MODEL"); val != "" {
		config.EmbeddingModel = val
	}
	if val := os.Getenv("GEMINI_MAX_TOKENS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxTokens = n
		}
	}
	if val := os.Getenv("GEMINI_TIMEOUT_SECONDS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.Timeout = time.Duration(n) * time.Second
		}
	}
	if val := os.Getenv("GEMINI_MAX_IMAGE_SIZE_MB"); val != "" {
		if n, err := strconv.ParseInt(val, 10, 64); err == nil && n > 0 {
			config.MaxImageSize = n * 1024 * 1024
		}
	}

	// Safety level
	if val := os.Getenv("GEMINI_SAFETY_LEVEL"); val != "" {
		switch strings.ToLower(val) {
		case "none":
			config.SafetyLevel = SafetyLevelNone
		case "low":
			config.SafetyLevel = SafetyLevelLow
		case "medium":
			config.SafetyLevel = SafetyLevelMedium
		case "high":
			config.SafetyLevel = SafetyLevelHigh
		case "block_all":
			config.SafetyLevel = SafetyLevelBlockAll
		}
	}

	// Input validation limits
	if val := os.Getenv("GEMINI_MAX_PROMPT_LENGTH"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxPromptLength = n
		}
	}
	if val := os.Getenv("GEMINI_MAX_MESSAGES_PER_CHAT"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxMessagesPerChat = n
		}
	}
	if val := os.Getenv("GEMINI_MAX_TEXTS_PER_EMBED"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxTextsPerEmbed = n
		}
	}

	// Feature flags
	if val := os.Getenv("GEMINI_ALLOW_FUNCTION_CALLING"); val != "" {
		config.AllowFunctionCalling = strings.ToLower(val) == "true" || val == "1"
	}
	if val := os.Getenv("GEMINI_ALLOW_JSON_MODE"); val != "" {
		config.AllowJSONMode = strings.ToLower(val) == "true" || val == "1"
	}

	// Enable AI if API key is provided
	if val := os.Getenv("GEMINI_ENABLED"); val != "" {
		config.Enabled = strings.ToLower(val) == "true" || val == "1"
	} else {
		config.Enabled = config.APIKey != ""
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.APIKey == "" {
		return ErrMissingAPIKey
	}

	return nil
}
