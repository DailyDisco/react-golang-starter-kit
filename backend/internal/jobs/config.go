package jobs

import (
	"os"
	"strconv"
	"time"
)

// Config holds job system configuration
type Config struct {
	// General settings
	Enabled     bool
	WorkerCount int

	// Queue settings
	MaxRetries   int
	RetryBackoff time.Duration
	JobTimeout   time.Duration

	// Maintenance
	RescueStuckJobsAfter time.Duration
}

// DefaultConfig returns sensible default job configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:              false, // Disabled by default
		WorkerCount:          10,
		MaxRetries:           3,
		RetryBackoff:         5 * time.Second,
		JobTimeout:           30 * time.Second,
		RescueStuckJobsAfter: 1 * time.Hour,
	}
}

// LoadConfig loads job configuration from environment variables
func LoadConfig() *Config {
	config := DefaultConfig()

	if enabled := os.Getenv("JOBS_ENABLED"); enabled != "" {
		config.Enabled = enabled == "true"
	}

	if workers := os.Getenv("JOBS_WORKER_COUNT"); workers != "" {
		if count, err := strconv.Atoi(workers); err == nil && count > 0 {
			config.WorkerCount = count
		}
	}

	if retries := os.Getenv("JOBS_MAX_RETRIES"); retries != "" {
		if count, err := strconv.Atoi(retries); err == nil && count >= 0 {
			config.MaxRetries = count
		}
	}

	if timeout := os.Getenv("JOBS_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.JobTimeout = d
		}
	}

	return config
}
