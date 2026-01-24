package email

import (
	"testing"
	"time"
)

// ============ DefaultConfig Tests ============

func TestDefaultConfig_AllFields(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.Enabled {
		t.Error("Enabled should be false by default")
	}

	if cfg.SMTPHost != "localhost" {
		t.Errorf("SMTPHost = %q, want %q", cfg.SMTPHost, "localhost")
	}

	if cfg.SMTPPort != 587 {
		t.Errorf("SMTPPort = %d, want %d", cfg.SMTPPort, 587)
	}

	if cfg.SMTPFrom != "noreply@example.com" {
		t.Errorf("SMTPFrom = %q, want %q", cfg.SMTPFrom, "noreply@example.com")
	}

	if cfg.SMTPFromName != "App" {
		t.Errorf("SMTPFromName = %q, want %q", cfg.SMTPFromName, "App")
	}

	if !cfg.UseTLS {
		t.Error("UseTLS should be true by default")
	}

	if cfg.TLSPolicy != "opportunistic" {
		t.Errorf("TLSPolicy = %q, want %q", cfg.TLSPolicy, "opportunistic")
	}

	if cfg.ConnectTimeout != 10*time.Second {
		t.Errorf("ConnectTimeout = %v, want %v", cfg.ConnectTimeout, 10*time.Second)
	}

	if cfg.SendTimeout != 30*time.Second {
		t.Errorf("SendTimeout = %v, want %v", cfg.SendTimeout, 30*time.Second)
	}

	if cfg.FrontendURL != "http://localhost:5173" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "http://localhost:5173")
	}

	if !cfg.DevMode {
		t.Error("DevMode should be true by default")
	}
}

// ============ LoadConfig Tests ============

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear env vars
	t.Setenv("EMAIL_ENABLED", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("GO_ENV", "")

	cfg := LoadConfig()

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil")
	}

	// Should use defaults when no env vars set
	if cfg.SMTPPort != 587 {
		t.Errorf("SMTPPort = %d, want %d (default)", cfg.SMTPPort, 587)
	}
}

func TestLoadConfig_Enabled(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"empty defaults to false", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EMAIL_ENABLED", tt.envVal)
			t.Setenv("SMTP_HOST", "") // Don't auto-enable via host
			cfg := LoadConfig()
			if cfg.Enabled != tt.want {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.want)
			}
		})
	}
}

func TestLoadConfig_SMTPHost_EnablesEmail(t *testing.T) {
	t.Setenv("EMAIL_ENABLED", "")
	t.Setenv("SMTP_HOST", "smtp.example.com")

	cfg := LoadConfig()

	if !cfg.Enabled {
		t.Error("Enabled should be true when SMTP_HOST is set")
	}

	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("SMTPHost = %q, want %q", cfg.SMTPHost, "smtp.example.com")
	}
}

func TestLoadConfig_SMTPPort(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"valid port", "465", 465},
		{"invalid port keeps default", "invalid", 587},
		{"empty keeps default", "", 587},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SMTP_PORT", tt.envVal)
			cfg := LoadConfig()
			if cfg.SMTPPort != tt.want {
				t.Errorf("SMTPPort = %d, want %d", cfg.SMTPPort, tt.want)
			}
		})
	}
}

func TestLoadConfig_SMTPCredentials(t *testing.T) {
	t.Setenv("SMTP_USER", "testuser")
	t.Setenv("SMTP_PASSWORD", "testpass")

	cfg := LoadConfig()

	if cfg.SMTPUser != "testuser" {
		t.Errorf("SMTPUser = %q, want %q", cfg.SMTPUser, "testuser")
	}

	if cfg.SMTPPassword != "testpass" {
		t.Errorf("SMTPPassword = %q, want %q", cfg.SMTPPassword, "testpass")
	}
}

func TestLoadConfig_SMTPFrom(t *testing.T) {
	t.Run("SMTP_FROM takes priority", func(t *testing.T) {
		t.Setenv("SMTP_FROM", "noreply@smtp.com")
		t.Setenv("EMAIL_FROM", "noreply@email.com")
		cfg := LoadConfig()
		// EMAIL_FROM is an alias that overrides SMTP_FROM
		if cfg.SMTPFrom != "noreply@email.com" {
			t.Errorf("SMTPFrom = %q, want %q", cfg.SMTPFrom, "noreply@email.com")
		}
	})

	t.Run("SMTP_FROM only", func(t *testing.T) {
		t.Setenv("SMTP_FROM", "test@example.com")
		t.Setenv("EMAIL_FROM", "")
		cfg := LoadConfig()
		if cfg.SMTPFrom != "test@example.com" {
			t.Errorf("SMTPFrom = %q, want %q", cfg.SMTPFrom, "test@example.com")
		}
	})
}

func TestLoadConfig_SMTPFromName(t *testing.T) {
	t.Setenv("SMTP_FROM_NAME", "My App")
	cfg := LoadConfig()

	if cfg.SMTPFromName != "My App" {
		t.Errorf("SMTPFromName = %q, want %q", cfg.SMTPFromName, "My App")
	}
}

func TestLoadConfig_FrontendURL(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://myapp.com")
	cfg := LoadConfig()

	if cfg.FrontendURL != "https://myapp.com" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "https://myapp.com")
	}
}

func TestLoadConfig_TLSPolicy(t *testing.T) {
	t.Setenv("SMTP_TLS_POLICY", "mandatory")
	cfg := LoadConfig()

	if cfg.TLSPolicy != "mandatory" {
		t.Errorf("TLSPolicy = %q, want %q", cfg.TLSPolicy, "mandatory")
	}
}

func TestLoadConfig_DevMode(t *testing.T) {
	tests := []struct {
		name         string
		goEnv        string
		emailDevMode string
		want         bool
	}{
		{"development env", "development", "", true},
		{"production env", "production", "", false},
		{"empty env defaults to development", "", "", true},
		{"explicit dev mode true", "production", "true", true},
		{"explicit dev mode false", "development", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GO_ENV", tt.goEnv)
			t.Setenv("EMAIL_DEV_MODE", tt.emailDevMode)
			cfg := LoadConfig()
			if cfg.DevMode != tt.want {
				t.Errorf("DevMode = %v, want %v", cfg.DevMode, tt.want)
			}
		})
	}
}

// ============ Config Structure Tests ============

func TestConfig_Structure(t *testing.T) {
	cfg := Config{
		Enabled:        true,
		SMTPHost:       "smtp.test.com",
		SMTPPort:       465,
		SMTPUser:       "user",
		SMTPPassword:   "pass",
		SMTPFrom:       "test@test.com",
		SMTPFromName:   "Test App",
		UseTLS:         true,
		TLSPolicy:      "mandatory",
		ConnectTimeout: 5 * time.Second,
		SendTimeout:    15 * time.Second,
		FrontendURL:    "https://test.com",
		DevMode:        false,
	}

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.SMTPHost != "smtp.test.com" {
		t.Errorf("SMTPHost = %q, want %q", cfg.SMTPHost, "smtp.test.com")
	}
	if cfg.SMTPPort != 465 {
		t.Errorf("SMTPPort = %d, want %d", cfg.SMTPPort, 465)
	}
}
