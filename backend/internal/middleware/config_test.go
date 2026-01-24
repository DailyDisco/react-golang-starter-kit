package middleware

import (
	"testing"
)

// ============ DefaultLogConfig Tests ============

func TestDefaultLogConfig_AllValues(t *testing.T) {
	config := DefaultLogConfig()

	if config == nil {
		t.Fatal("DefaultLogConfig() returned nil")
	}

	// Test default values
	if !config.Enabled {
		t.Error("DefaultLogConfig().Enabled = false, want true")
	}

	if config.Level != "info" {
		t.Errorf("DefaultLogConfig().Level = %q, want %q", config.Level, "info")
	}

	if !config.IncludeUserContext {
		t.Error("DefaultLogConfig().IncludeUserContext = false, want true")
	}

	if config.IncludeRequestBody {
		t.Error("DefaultLogConfig().IncludeRequestBody = true, want false")
	}

	if config.IncludeResponseBody {
		t.Error("DefaultLogConfig().IncludeResponseBody = true, want false")
	}

	if config.MaxRequestBodySize != 4096 {
		t.Errorf("DefaultLogConfig().MaxRequestBodySize = %d, want 4096", config.MaxRequestBodySize)
	}

	if config.MaxResponseBodySize != 4096 {
		t.Errorf("DefaultLogConfig().MaxResponseBodySize = %d, want 4096", config.MaxResponseBodySize)
	}

	if config.SamplingRate != 1.0 {
		t.Errorf("DefaultLogConfig().SamplingRate = %f, want 1.0", config.SamplingRate)
	}

	if config.AsyncLogging {
		t.Error("DefaultLogConfig().AsyncLogging = true, want false")
	}

	if !config.SanitizeHeaders {
		t.Error("DefaultLogConfig().SanitizeHeaders = false, want true")
	}

	if !config.FilterSensitiveData {
		t.Error("DefaultLogConfig().FilterSensitiveData = false, want true")
	}

	if config.TimeFormat != "2006-01-02T15:04:05Z07:00" {
		t.Errorf("DefaultLogConfig().TimeFormat = %q, want RFC3339", config.TimeFormat)
	}

	if config.Pretty {
		t.Error("DefaultLogConfig().Pretty = true, want false")
	}
}

func TestDefaultLogConfig_AllowedHeaders(t *testing.T) {
	config := DefaultLogConfig()

	expectedHeaders := []string{
		"accept",
		"accept-encoding",
		"accept-language",
		"cache-control",
		"content-length",
		"content-type",
		"user-agent",
		"x-request-id",
	}

	if len(config.AllowedHeaders) != len(expectedHeaders) {
		t.Errorf("AllowedHeaders length = %d, want %d", len(config.AllowedHeaders), len(expectedHeaders))
	}

	headerSet := make(map[string]bool)
	for _, h := range config.AllowedHeaders {
		headerSet[h] = true
	}

	for _, expected := range expectedHeaders {
		if !headerSet[expected] {
			t.Errorf("AllowedHeaders missing %q", expected)
		}
	}
}

// ============ LoadLogConfig Tests ============

func TestLoadLogConfig_WithDefaults(t *testing.T) {
	// Clear any env vars that might affect the test
	t.Setenv("LOG_ENABLED", "")
	t.Setenv("LOG_LEVEL", "")

	config := LoadLogConfig()
	if config == nil {
		t.Fatal("LoadLogConfig() returned nil")
	}
}

func TestLoadLogConfig_Enabled(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"True mixed", "True", true},
		{"false", "false", false},
		{"1", "1", false}, // Only "true" enables
		{"empty defaults to true", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_ENABLED", tt.envVal)
			config := LoadLogConfig()
			if config.Enabled != tt.want {
				t.Errorf("LoadLogConfig().Enabled = %v, want %v", config.Enabled, tt.want)
			}
		})
	}
}

func TestLoadLogConfig_Level(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   string
	}{
		{"debug", "debug", "debug"},
		{"DEBUG uppercase", "DEBUG", "debug"},
		{"info", "info", "info"},
		{"warn", "warn", "warn"},
		{"error", "error", "error"},
		{"empty defaults to info", "", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tt.envVal)
			config := LoadLogConfig()
			if config.Level != tt.want {
				t.Errorf("LoadLogConfig().Level = %q, want %q", config.Level, tt.want)
			}
		})
	}
}

func TestLoadLogConfig_BodySizes(t *testing.T) {
	t.Run("valid max request body size", func(t *testing.T) {
		t.Setenv("LOG_MAX_REQUEST_BODY_SIZE", "8192")
		config := LoadLogConfig()
		if config.MaxRequestBodySize != 8192 {
			t.Errorf("MaxRequestBodySize = %d, want 8192", config.MaxRequestBodySize)
		}
	})

	t.Run("valid max response body size", func(t *testing.T) {
		t.Setenv("LOG_MAX_RESPONSE_BODY_SIZE", "16384")
		config := LoadLogConfig()
		if config.MaxResponseBodySize != 16384 {
			t.Errorf("MaxResponseBodySize = %d, want 16384", config.MaxResponseBodySize)
		}
	})

	t.Run("invalid max request body size keeps default", func(t *testing.T) {
		t.Setenv("LOG_MAX_REQUEST_BODY_SIZE", "invalid")
		config := LoadLogConfig()
		if config.MaxRequestBodySize != 4096 {
			t.Errorf("MaxRequestBodySize = %d, want 4096 (default)", config.MaxRequestBodySize)
		}
	})

	t.Run("zero max request body size keeps default", func(t *testing.T) {
		t.Setenv("LOG_MAX_REQUEST_BODY_SIZE", "0")
		config := LoadLogConfig()
		if config.MaxRequestBodySize != 4096 {
			t.Errorf("MaxRequestBodySize = %d, want 4096 (default)", config.MaxRequestBodySize)
		}
	})

	t.Run("negative max request body size keeps default", func(t *testing.T) {
		t.Setenv("LOG_MAX_REQUEST_BODY_SIZE", "-100")
		config := LoadLogConfig()
		if config.MaxRequestBodySize != 4096 {
			t.Errorf("MaxRequestBodySize = %d, want 4096 (default)", config.MaxRequestBodySize)
		}
	})
}

func TestLoadLogConfig_SamplingRate(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   float64
	}{
		{"0.5", "0.5", 0.5},
		{"0.0", "0.0", 0.0},
		{"1.0", "1.0", 1.0},
		{"invalid keeps default", "invalid", 1.0},
		{"negative keeps default", "-0.5", 1.0},
		{"over 1 keeps default", "1.5", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_SAMPLING_RATE", tt.envVal)
			config := LoadLogConfig()
			if config.SamplingRate != tt.want {
				t.Errorf("SamplingRate = %f, want %f", config.SamplingRate, tt.want)
			}
		})
	}
}

func TestLoadLogConfig_AllowedHeaders(t *testing.T) {
	t.Setenv("LOG_ALLOWED_HEADERS", "content-type, X-Custom-Header, authorization")
	config := LoadLogConfig()

	expected := []string{"content-type", "x-custom-header", "authorization"}
	if len(config.AllowedHeaders) != len(expected) {
		t.Fatalf("AllowedHeaders length = %d, want %d", len(config.AllowedHeaders), len(expected))
	}

	for i, header := range config.AllowedHeaders {
		if header != expected[i] {
			t.Errorf("AllowedHeaders[%d] = %q, want %q", i, header, expected[i])
		}
	}
}

// ============ ShouldLogRequest Tests ============

func TestLogConfig_ShouldLogRequest(t *testing.T) {
	tests := []struct {
		name         string
		samplingRate float64
		wantAlways   bool
		wantNever    bool
	}{
		{"sampling rate 1.0 always logs", 1.0, true, false},
		{"sampling rate 0.0 never logs", 0.0, false, true},
		{"sampling rate 0.5 sometimes logs", 0.5, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LogConfig{SamplingRate: tt.samplingRate}

			if tt.wantAlways {
				for i := 0; i < 100; i++ {
					if !config.ShouldLogRequest() {
						t.Error("ShouldLogRequest() should always return true for rate 1.0")
					}
				}
			}

			if tt.wantNever {
				for i := 0; i < 100; i++ {
					if config.ShouldLogRequest() {
						t.Error("ShouldLogRequest() should always return false for rate 0.0")
					}
				}
			}
		})
	}
}

// ============ IsHeaderAllowed Tests ============

func TestLogConfig_IsHeaderAllowed(t *testing.T) {
	config := &LogConfig{
		SanitizeHeaders: true,
		AllowedHeaders:  []string{"content-type", "accept", "user-agent"},
	}

	tests := []struct {
		name    string
		header  string
		allowed bool
	}{
		{"allowed header exact case", "content-type", true},
		{"allowed header different case", "Content-Type", true},
		{"allowed header uppercase", "CONTENT-TYPE", true},
		{"another allowed header", "accept", true},
		{"not allowed header", "authorization", false},
		{"not allowed header sensitive", "cookie", false},
		{"empty header", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if config.IsHeaderAllowed(tt.header) != tt.allowed {
				t.Errorf("IsHeaderAllowed(%q) = %v, want %v", tt.header, !tt.allowed, tt.allowed)
			}
		})
	}
}

func TestLogConfig_IsHeaderAllowed_SanitizationDisabled(t *testing.T) {
	config := &LogConfig{
		SanitizeHeaders: false,
		AllowedHeaders:  []string{"content-type"},
	}

	// When sanitization is disabled, all headers should be allowed
	tests := []string{"content-type", "authorization", "cookie", "x-custom-header"}

	for _, header := range tests {
		if !config.IsHeaderAllowed(header) {
			t.Errorf("IsHeaderAllowed(%q) = false, want true when sanitization disabled", header)
		}
	}
}

// ============ LogConfig Structure Tests ============

func TestLogConfig_Structure(t *testing.T) {
	config := LogConfig{
		Enabled:             true,
		Level:               "debug",
		IncludeUserContext:  true,
		IncludeRequestBody:  true,
		IncludeResponseBody: true,
		MaxRequestBodySize:  1024,
		MaxResponseBodySize: 2048,
		SamplingRate:        0.5,
		AsyncLogging:        true,
		SanitizeHeaders:     true,
		FilterSensitiveData: true,
		AllowedHeaders:      []string{"content-type"},
		TimeFormat:          "2006-01-02",
		Pretty:              true,
	}

	if config.Level != "debug" {
		t.Errorf("Level = %q, want %q", config.Level, "debug")
	}
	if config.MaxRequestBodySize != 1024 {
		t.Errorf("MaxRequestBodySize = %d, want %d", config.MaxRequestBodySize, 1024)
	}
	if config.SamplingRate != 0.5 {
		t.Errorf("SamplingRate = %f, want %f", config.SamplingRate, 0.5)
	}
}
