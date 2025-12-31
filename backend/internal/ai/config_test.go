package ai

import (
	"os"
	"testing"
	"time"
)

// ============ Default Config Tests ============

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Enabled {
		t.Error("DefaultConfig().Enabled = true, want false")
	}

	if config.APIKey != "" {
		t.Errorf("DefaultConfig().APIKey = %q, want empty", config.APIKey)
	}

	if config.Model != "gemini-2.0-flash" {
		t.Errorf("DefaultConfig().Model = %q, want %q", config.Model, "gemini-2.0-flash")
	}

	if config.EmbeddingModel != "text-embedding-004" {
		t.Errorf("DefaultConfig().EmbeddingModel = %q, want %q", config.EmbeddingModel, "text-embedding-004")
	}

	if config.MaxTokens != 8192 {
		t.Errorf("DefaultConfig().MaxTokens = %d, want %d", config.MaxTokens, 8192)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("DefaultConfig().Timeout = %v, want %v", config.Timeout, 30*time.Second)
	}

	if config.MaxImageSize != 10*1024*1024 {
		t.Errorf("DefaultConfig().MaxImageSize = %d, want %d", config.MaxImageSize, 10*1024*1024)
	}

	if config.SafetyLevel != SafetyLevelMedium {
		t.Errorf("DefaultConfig().SafetyLevel = %q, want %q", config.SafetyLevel, SafetyLevelMedium)
	}

	if config.MaxPromptLength != 100000 {
		t.Errorf("DefaultConfig().MaxPromptLength = %d, want %d", config.MaxPromptLength, 100000)
	}

	if config.MaxMessagesPerChat != 50 {
		t.Errorf("DefaultConfig().MaxMessagesPerChat = %d, want %d", config.MaxMessagesPerChat, 50)
	}

	if config.MaxTextsPerEmbed != 100 {
		t.Errorf("DefaultConfig().MaxTextsPerEmbed = %d, want %d", config.MaxTextsPerEmbed, 100)
	}

	if !config.AllowFunctionCalling {
		t.Error("DefaultConfig().AllowFunctionCalling = false, want true")
	}

	if !config.AllowJSONMode {
		t.Error("DefaultConfig().AllowJSONMode = false, want true")
	}
}

// ============ Load Config Tests ============

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all Gemini env vars
	envVars := []string{
		"GEMINI_API_KEY",
		"GEMINI_MODEL",
		"GEMINI_EMBEDDING_MODEL",
		"GEMINI_MAX_TOKENS",
		"GEMINI_TIMEOUT_SECONDS",
		"GEMINI_MAX_IMAGE_SIZE_MB",
		"GEMINI_SAFETY_LEVEL",
		"GEMINI_MAX_PROMPT_LENGTH",
		"GEMINI_MAX_MESSAGES_PER_CHAT",
		"GEMINI_MAX_TEXTS_PER_EMBED",
		"GEMINI_ALLOW_FUNCTION_CALLING",
		"GEMINI_ALLOW_JSON_MODE",
		"GEMINI_ENABLED",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	config := LoadConfig()

	if config.Enabled {
		t.Error("LoadConfig() should be disabled when no API key is set")
	}

	if config.Model != "gemini-2.0-flash" {
		t.Errorf("LoadConfig().Model = %q, want default", config.Model)
	}
}

func TestLoadConfig_WithAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-api-key-12345")

	config := LoadConfig()

	if !config.Enabled {
		t.Error("LoadConfig() should be enabled when API key is set")
	}

	if config.APIKey != "test-api-key-12345" {
		t.Errorf("LoadConfig().APIKey = %q, want %q", config.APIKey, "test-api-key-12345")
	}
}

func TestLoadConfig_CustomModel(t *testing.T) {
	t.Setenv("GEMINI_MODEL", "gemini-1.5-pro")

	config := LoadConfig()

	if config.Model != "gemini-1.5-pro" {
		t.Errorf("LoadConfig().Model = %q, want %q", config.Model, "gemini-1.5-pro")
	}
}

func TestLoadConfig_CustomEmbeddingModel(t *testing.T) {
	t.Setenv("GEMINI_EMBEDDING_MODEL", "text-embedding-005")

	config := LoadConfig()

	if config.EmbeddingModel != "text-embedding-005" {
		t.Errorf("LoadConfig().EmbeddingModel = %q, want %q", config.EmbeddingModel, "text-embedding-005")
	}
}

func TestLoadConfig_CustomMaxTokens(t *testing.T) {
	t.Setenv("GEMINI_MAX_TOKENS", "4096")

	config := LoadConfig()

	if config.MaxTokens != 4096 {
		t.Errorf("LoadConfig().MaxTokens = %d, want %d", config.MaxTokens, 4096)
	}
}

func TestLoadConfig_InvalidMaxTokens(t *testing.T) {
	t.Setenv("GEMINI_MAX_TOKENS", "invalid")

	config := LoadConfig()

	// Should keep default
	if config.MaxTokens != 8192 {
		t.Errorf("LoadConfig().MaxTokens = %d, want default %d", config.MaxTokens, 8192)
	}
}

func TestLoadConfig_CustomTimeout(t *testing.T) {
	t.Setenv("GEMINI_TIMEOUT_SECONDS", "60")

	config := LoadConfig()

	if config.Timeout != 60*time.Second {
		t.Errorf("LoadConfig().Timeout = %v, want %v", config.Timeout, 60*time.Second)
	}
}

func TestLoadConfig_CustomImageSize(t *testing.T) {
	t.Setenv("GEMINI_MAX_IMAGE_SIZE_MB", "20")

	config := LoadConfig()

	if config.MaxImageSize != 20*1024*1024 {
		t.Errorf("LoadConfig().MaxImageSize = %d, want %d", config.MaxImageSize, 20*1024*1024)
	}
}

func TestLoadConfig_SafetyLevels(t *testing.T) {
	tests := []struct {
		envValue string
		expected SafetyLevel
	}{
		{"none", SafetyLevelNone},
		{"low", SafetyLevelLow},
		{"medium", SafetyLevelMedium},
		{"high", SafetyLevelHigh},
		{"block_all", SafetyLevelBlockAll},
		{"NONE", SafetyLevelNone},
		{"LOW", SafetyLevelLow},
		{"MEDIUM", SafetyLevelMedium},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			t.Setenv("GEMINI_SAFETY_LEVEL", tt.envValue)

			config := LoadConfig()

			if config.SafetyLevel != tt.expected {
				t.Errorf("LoadConfig().SafetyLevel = %q, want %q", config.SafetyLevel, tt.expected)
			}
		})
	}
}

func TestLoadConfig_FeatureFlags(t *testing.T) {
	tests := []struct {
		name      string
		envVar    string
		envValue  string
		checkFunc func(*Config) bool
	}{
		{"function calling true", "GEMINI_ALLOW_FUNCTION_CALLING", "true", func(c *Config) bool { return c.AllowFunctionCalling }},
		{"function calling false", "GEMINI_ALLOW_FUNCTION_CALLING", "false", func(c *Config) bool { return !c.AllowFunctionCalling }},
		{"function calling 1", "GEMINI_ALLOW_FUNCTION_CALLING", "1", func(c *Config) bool { return c.AllowFunctionCalling }},
		{"json mode true", "GEMINI_ALLOW_JSON_MODE", "true", func(c *Config) bool { return c.AllowJSONMode }},
		{"json mode false", "GEMINI_ALLOW_JSON_MODE", "false", func(c *Config) bool { return !c.AllowJSONMode }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envValue)

			config := LoadConfig()

			if !tt.checkFunc(config) {
				t.Errorf("LoadConfig() %s check failed", tt.name)
			}
		})
	}
}

func TestLoadConfig_ExplicitEnabled(t *testing.T) {
	t.Setenv("GEMINI_ENABLED", "true")
	t.Setenv("GEMINI_API_KEY", "test-key")

	config := LoadConfig()

	if !config.Enabled {
		t.Error("LoadConfig() should be enabled when GEMINI_ENABLED=true")
	}
}

func TestLoadConfig_ExplicitDisabled(t *testing.T) {
	t.Setenv("GEMINI_ENABLED", "false")
	t.Setenv("GEMINI_API_KEY", "test-key")

	config := LoadConfig()

	if config.Enabled {
		t.Error("LoadConfig() should be disabled when GEMINI_ENABLED=false even with API key")
	}
}

// ============ Config Validation Tests ============

func TestConfig_Validate_Disabled(t *testing.T) {
	config := &Config{
		Enabled: false,
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Config.Validate() error = %v, want nil for disabled config", err)
	}
}

func TestConfig_Validate_MissingAPIKey(t *testing.T) {
	config := &Config{
		Enabled: true,
		APIKey:  "",
	}

	err := config.Validate()
	if err != ErrMissingAPIKey {
		t.Errorf("Config.Validate() error = %v, want %v", err, ErrMissingAPIKey)
	}
}

func TestConfig_Validate_Valid(t *testing.T) {
	config := &Config{
		Enabled: true,
		APIKey:  "test-api-key",
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Config.Validate() error = %v, want nil", err)
	}
}

// ============ Safety Level Tests ============

func TestSafetyLevelConstants(t *testing.T) {
	tests := []struct {
		level SafetyLevel
		want  string
	}{
		{SafetyLevelNone, "none"},
		{SafetyLevelLow, "low"},
		{SafetyLevelMedium, "medium"},
		{SafetyLevelHigh, "high"},
		{SafetyLevelBlockAll, "block_all"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if string(tt.level) != tt.want {
				t.Errorf("SafetyLevel = %q, want %q", tt.level, tt.want)
			}
		})
	}
}
