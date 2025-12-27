package jobs

import (
	"context"
	"os"
	"strconv"
	"time"

	"react-golang-starter/internal/database"

	"github.com/rs/zerolog/log"
)

// MetricsRetentionConfig holds configuration for metrics data retention
type MetricsRetentionConfig struct {
	// RetentionDays is the number of days to retain metrics data (default: 30)
	RetentionDays int

	// Enabled controls whether retention cleanup is active
	Enabled bool
}

// DefaultMetricsRetentionConfig returns default retention configuration
func DefaultMetricsRetentionConfig() *MetricsRetentionConfig {
	return &MetricsRetentionConfig{
		RetentionDays: 30,
		Enabled:       true,
	}
}

// LoadMetricsRetentionConfig loads retention configuration from environment variables
func LoadMetricsRetentionConfig() *MetricsRetentionConfig {
	config := DefaultMetricsRetentionConfig()

	// METRICS_RETENTION_ENABLED (default: true)
	if enabled := os.Getenv("METRICS_RETENTION_ENABLED"); enabled != "" {
		config.Enabled = enabled == "true" || enabled == "1"
	}

	// METRICS_RETENTION_DAYS (default: 30)
	if daysStr := os.Getenv("METRICS_RETENTION_DAYS"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			config.RetentionDays = days
		}
	}

	return config
}

// RunMetricsRetention performs cleanup of old metrics data
// This deletes records older than the configured retention period from:
// - container_metrics_histories
// - service_uptime_histories
func RunMetricsRetention(ctx context.Context, config *MetricsRetentionConfig) error {
	if config == nil || !config.Enabled {
		log.Info().Msg("Metrics retention disabled, skipping cleanup")
		return nil
	}

	// Calculate cutoff timestamp (Unix timestamp in seconds)
	cutoffTime := time.Now().AddDate(0, 0, -config.RetentionDays).Unix()

	log.Info().
		Int("retention_days", config.RetentionDays).
		Int64("cutoff_timestamp", cutoffTime).
		Msg("Starting metrics retention cleanup")

	var containerDeleted, serviceDeleted int64

	// Check if container_metrics_histories table exists before cleaning
	var containerTableExists bool
	database.DB.WithContext(ctx).Raw(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'container_metrics_histories')",
	).Scan(&containerTableExists)

	if containerTableExists {
		containerResult := database.DB.WithContext(ctx).Exec(
			"DELETE FROM container_metrics_histories WHERE recorded_at < ?",
			cutoffTime,
		)
		if containerResult.Error != nil {
			log.Error().Err(containerResult.Error).Msg("Failed to clean container_metrics_histories")
			return containerResult.Error
		}
		containerDeleted = containerResult.RowsAffected
	}

	// Check if service_uptime_histories table exists before cleaning
	var serviceTableExists bool
	database.DB.WithContext(ctx).Raw(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'service_uptime_histories')",
	).Scan(&serviceTableExists)

	if serviceTableExists {
		serviceResult := database.DB.WithContext(ctx).Exec(
			"DELETE FROM service_uptime_histories WHERE checked_at < ?",
			cutoffTime,
		)
		if serviceResult.Error != nil {
			log.Error().Err(serviceResult.Error).Msg("Failed to clean service_uptime_histories")
			return serviceResult.Error
		}
		serviceDeleted = serviceResult.RowsAffected
	}

	log.Info().
		Int64("container_metrics_deleted", containerDeleted).
		Int64("service_uptime_deleted", serviceDeleted).
		Int("retention_days", config.RetentionDays).
		Msg("Metrics retention cleanup completed")

	return nil
}

// StartPeriodicRetention starts a background goroutine that runs retention cleanup
// at the specified interval (default: every 24 hours)
func StartPeriodicRetention(ctx context.Context, config *MetricsRetentionConfig) {
	if config == nil || !config.Enabled {
		return
	}

	go func() {
		// Run immediately on startup
		if err := RunMetricsRetention(ctx, config); err != nil {
			log.Error().Err(err).Msg("Initial metrics retention cleanup failed")
		}

		// Then run every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Metrics retention cleanup stopped")
				return
			case <-ticker.C:
				if err := RunMetricsRetention(ctx, config); err != nil {
					log.Error().Err(err).Msg("Periodic metrics retention cleanup failed")
				}
			}
		}
	}()

	log.Info().
		Int("retention_days", config.RetentionDays).
		Msg("Started periodic metrics retention cleanup (runs every 24 hours)")
}
