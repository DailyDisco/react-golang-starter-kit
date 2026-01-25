package jobs

import (
	"testing"
)

// ============ DefaultMetricsRetentionConfig Tests ============

func TestDefaultMetricsRetentionConfig(t *testing.T) {
	config := DefaultMetricsRetentionConfig()

	if config == nil {
		t.Fatal("DefaultMetricsRetentionConfig() returned nil")
	}

	if config.RetentionDays != 30 {
		t.Errorf("RetentionDays = %d, want 30", config.RetentionDays)
	}

	if !config.Enabled {
		t.Error("Enabled should be true by default")
	}
}

// ============ LoadMetricsRetentionConfig Tests ============

func TestLoadMetricsRetentionConfig_Defaults(t *testing.T) {
	// Clear environment variables
	t.Setenv("METRICS_RETENTION_ENABLED", "")
	t.Setenv("METRICS_RETENTION_DAYS", "")

	config := LoadMetricsRetentionConfig()

	// Should match DefaultMetricsRetentionConfig
	defaults := DefaultMetricsRetentionConfig()
	if config.Enabled != defaults.Enabled {
		t.Errorf("Enabled = %v, want %v", config.Enabled, defaults.Enabled)
	}
	if config.RetentionDays != defaults.RetentionDays {
		t.Errorf("RetentionDays = %d, want %d", config.RetentionDays, defaults.RetentionDays)
	}
}

func TestLoadMetricsRetentionConfig_Enabled(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"true string", "true", true},
		{"1 string", "1", true},
		{"false string", "false", false},
		{"0 string", "0", false},
		{"empty uses default", "", true},
		{"random string", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METRICS_RETENTION_ENABLED", tt.envVal)
			config := LoadMetricsRetentionConfig()
			if config.Enabled != tt.want {
				t.Errorf("Enabled = %v, want %v", config.Enabled, tt.want)
			}
		})
	}
}

func TestLoadMetricsRetentionConfig_RetentionDays(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"valid number", "90", 90},
		{"minimum 1", "1", 1},
		{"large number", "365", 365},
		{"zero uses default", "0", 30},
		{"negative uses default", "-5", 30},
		{"invalid uses default", "abc", 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METRICS_RETENTION_DAYS", tt.envVal)
			config := LoadMetricsRetentionConfig()
			if config.RetentionDays != tt.want {
				t.Errorf("RetentionDays = %d, want %d", config.RetentionDays, tt.want)
			}
		})
	}
}

// ============ MetricsRetentionConfig Structure Tests ============

func TestMetricsRetentionConfig_Structure(t *testing.T) {
	config := MetricsRetentionConfig{
		RetentionDays: 60,
		Enabled:       false,
	}

	if config.RetentionDays != 60 {
		t.Errorf("RetentionDays = %d, want 60", config.RetentionDays)
	}

	if config.Enabled {
		t.Error("Enabled should be false")
	}
}

// ============ Combined Configuration Tests ============

func TestLoadMetricsRetentionConfig_AllValues(t *testing.T) {
	t.Setenv("METRICS_RETENTION_ENABLED", "true")
	t.Setenv("METRICS_RETENTION_DAYS", "7")

	config := LoadMetricsRetentionConfig()

	if !config.Enabled {
		t.Error("Enabled should be true")
	}
	if config.RetentionDays != 7 {
		t.Errorf("RetentionDays = %d, want 7", config.RetentionDays)
	}
}

func TestLoadMetricsRetentionConfig_DisabledWithCustomDays(t *testing.T) {
	t.Setenv("METRICS_RETENTION_ENABLED", "false")
	t.Setenv("METRICS_RETENTION_DAYS", "14")

	config := LoadMetricsRetentionConfig()

	if config.Enabled {
		t.Error("Enabled should be false")
	}
	if config.RetentionDays != 14 {
		t.Errorf("RetentionDays = %d, want 14", config.RetentionDays)
	}
}
