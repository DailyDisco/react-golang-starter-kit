package jobs

import (
	"testing"
	"time"
)

// ============ DefaultConfig Tests ============

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.Enabled {
		t.Error("Enabled should be false by default")
	}

	if config.WorkerCount != 10 {
		t.Errorf("WorkerCount = %d, want 10", config.WorkerCount)
	}

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}

	if config.RetryBackoff != 5*time.Second {
		t.Errorf("RetryBackoff = %v, want 5s", config.RetryBackoff)
	}

	if config.JobTimeout != 30*time.Second {
		t.Errorf("JobTimeout = %v, want 30s", config.JobTimeout)
	}

	if config.RescueStuckJobsAfter != 1*time.Hour {
		t.Errorf("RescueStuckJobsAfter = %v, want 1h", config.RescueStuckJobsAfter)
	}
}

// ============ LoadConfig Tests ============

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables
	t.Setenv("JOBS_ENABLED", "")
	t.Setenv("JOBS_WORKER_COUNT", "")
	t.Setenv("JOBS_MAX_RETRIES", "")
	t.Setenv("JOBS_TIMEOUT", "")

	config := LoadConfig()

	// Should match DefaultConfig
	defaults := DefaultConfig()
	if config.Enabled != defaults.Enabled {
		t.Errorf("Enabled = %v, want %v", config.Enabled, defaults.Enabled)
	}
	if config.WorkerCount != defaults.WorkerCount {
		t.Errorf("WorkerCount = %d, want %d", config.WorkerCount, defaults.WorkerCount)
	}
	if config.MaxRetries != defaults.MaxRetries {
		t.Errorf("MaxRetries = %d, want %d", config.MaxRetries, defaults.MaxRetries)
	}
}

func TestLoadConfig_Enabled(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"true string", "true", true},
		{"false string", "false", false},
		{"empty string", "", false},
		{"random string", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JOBS_ENABLED", tt.envVal)
			config := LoadConfig()
			if config.Enabled != tt.want {
				t.Errorf("Enabled = %v, want %v", config.Enabled, tt.want)
			}
		})
	}
}

func TestLoadConfig_WorkerCount(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"valid number", "20", 20},
		{"minimum 1", "1", 1},
		{"zero uses default", "0", 10},
		{"negative uses default", "-5", 10},
		{"invalid uses default", "abc", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JOBS_WORKER_COUNT", tt.envVal)
			config := LoadConfig()
			if config.WorkerCount != tt.want {
				t.Errorf("WorkerCount = %d, want %d", config.WorkerCount, tt.want)
			}
		})
	}
}

func TestLoadConfig_MaxRetries(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   int
	}{
		{"valid number", "5", 5},
		{"zero retries", "0", 0},
		{"negative uses default", "-1", 3},
		{"invalid uses default", "abc", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JOBS_MAX_RETRIES", tt.envVal)
			config := LoadConfig()
			if config.MaxRetries != tt.want {
				t.Errorf("MaxRetries = %d, want %d", config.MaxRetries, tt.want)
			}
		})
	}
}

func TestLoadConfig_Timeout(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   time.Duration
	}{
		{"valid duration", "60s", 60 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"invalid uses default", "abc", 30 * time.Second},
		{"empty uses default", "", 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JOBS_TIMEOUT", tt.envVal)
			config := LoadConfig()
			if config.JobTimeout != tt.want {
				t.Errorf("JobTimeout = %v, want %v", config.JobTimeout, tt.want)
			}
		})
	}
}

// ============ Config Struct Tests ============

func TestConfig_Structure(t *testing.T) {
	config := Config{
		Enabled:              true,
		WorkerCount:          5,
		MaxRetries:           10,
		RetryBackoff:         10 * time.Second,
		JobTimeout:           1 * time.Minute,
		RescueStuckJobsAfter: 2 * time.Hour,
	}

	if !config.Enabled {
		t.Error("Enabled should be true")
	}
	if config.WorkerCount != 5 {
		t.Errorf("WorkerCount = %d, want 5", config.WorkerCount)
	}
	if config.MaxRetries != 10 {
		t.Errorf("MaxRetries = %d, want 10", config.MaxRetries)
	}
	if config.RetryBackoff != 10*time.Second {
		t.Errorf("RetryBackoff = %v, want 10s", config.RetryBackoff)
	}
	if config.JobTimeout != 1*time.Minute {
		t.Errorf("JobTimeout = %v, want 1m", config.JobTimeout)
	}
	if config.RescueStuckJobsAfter != 2*time.Hour {
		t.Errorf("RescueStuckJobsAfter = %v, want 2h", config.RescueStuckJobsAfter)
	}
}
