// Package database provides database connection and pool management.
package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// PoolConfig holds configuration for the database connection pool.
type PoolConfig struct {
	// Maximum number of connections in the pool
	MaxConns int32
	// Minimum number of connections to maintain
	MinConns int32
	// Maximum time a connection can be open
	MaxConnLifetime time.Duration
	// Maximum time a connection can be idle
	MaxConnIdleTime time.Duration
	// Health check period for idle connections
	HealthCheckPeriod time.Duration
}

// DefaultPoolConfig returns sensible default pool settings.
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   5 * time.Minute,
		MaxConnIdleTime:   1 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
}

// LoadPoolConfig loads pool configuration from environment variables with sensible defaults.
// Environment variables:
//   - DB_POOL_MAX_CONNS: Maximum connections (default: 25)
//   - DB_POOL_MIN_CONNS: Minimum connections (default: 5)
//   - DB_POOL_MAX_CONN_LIFETIME_MINUTES: Max connection lifetime in minutes (default: 5)
//   - DB_POOL_MAX_CONN_IDLE_TIME_MINUTES: Max idle time in minutes (default: 1)
//   - DB_POOL_HEALTH_CHECK_SECONDS: Health check period in seconds (default: 30)
func LoadPoolConfig() *PoolConfig {
	config := DefaultPoolConfig()

	if val := os.Getenv("DB_POOL_MAX_CONNS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxConns = int32(n)
		}
	}

	if val := os.Getenv("DB_POOL_MIN_CONNS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n >= 0 {
			config.MinConns = int32(n)
		}
	}

	if val := os.Getenv("DB_POOL_MAX_CONN_LIFETIME_MINUTES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxConnLifetime = time.Duration(n) * time.Minute
		}
	}

	if val := os.Getenv("DB_POOL_MAX_CONN_IDLE_TIME_MINUTES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.MaxConnIdleTime = time.Duration(n) * time.Minute
		}
	}

	if val := os.Getenv("DB_POOL_HEALTH_CHECK_SECONDS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.HealthCheckPeriod = time.Duration(n) * time.Second
		}
	}

	return config
}

// Pool is the shared pgx connection pool for the application.
// This pool is used by both GORM and River jobs to reduce connection overhead.
var Pool *pgxpool.Pool

// InitPool initializes the shared pgx connection pool.
// This should be called before initializing GORM or the jobs system.
func InitPool(config *PoolConfig) error {
	if config == nil {
		config = DefaultPoolConfig()
	}

	connString := buildConnectionString()

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Apply pool settings
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	// Retry connection with exponential backoff
	maxRetries := 10
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			// Test the connection
			if pingErr := Pool.Ping(context.Background()); pingErr == nil {
				log.Info().
					Int32("max_conns", config.MaxConns).
					Int32("min_conns", config.MinConns).
					Dur("max_lifetime", config.MaxConnLifetime).
					Dur("max_idle_time", config.MaxConnIdleTime).
					Msg("Shared pgx pool initialized")
				return nil
			} else {
				lastErr = pingErr
				Pool.Close()
			}
		} else {
			lastErr = err
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 2 * time.Second
			log.Error().
				Err(lastErr).
				Int("attempt", i+1).
				Int("max_retries", maxRetries).
				Dur("retry_in", waitTime).
				Msg("Failed to connect to PostgreSQL, retrying")
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("failed to initialize pgx pool after %d retries: %w", maxRetries, lastErr)
}

// ClosePool closes the shared connection pool.
func ClosePool() {
	if Pool != nil {
		Pool.Close()
		log.Info().Msg("Shared pgx pool closed")
	}
}

// GetPool returns the shared connection pool.
// Returns nil if the pool hasn't been initialized.
func GetPool() *pgxpool.Pool {
	return Pool
}

// PoolStats returns current pool statistics.
type PoolStats struct {
	AcquireCount         int64 `json:"acquire_count"`
	AcquiredConns        int32 `json:"acquired_conns"`
	CanceledAcquireCount int64 `json:"canceled_acquire_count"`
	ConstructingConns    int32 `json:"constructing_conns"`
	EmptyAcquireCount    int64 `json:"empty_acquire_count"`
	IdleConns            int32 `json:"idle_conns"`
	MaxConns             int32 `json:"max_conns"`
	TotalConns           int32 `json:"total_conns"`
}

// GetPoolStats returns current connection pool statistics.
func GetPoolStats() *PoolStats {
	if Pool == nil {
		return nil
	}

	stats := Pool.Stat()
	return &PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}
}

// buildConnectionString builds a PostgreSQL connection string from environment variables.
func buildConnectionString() string {
	host := getPoolEnv("PGHOST", getPoolEnv("DB_HOST", "localhost"))
	port := getPoolEnv("PGPORT", getPoolEnv("DB_PORT", "5432"))
	user := getPoolEnv("PGUSER", getPoolEnv("DB_USER", "devuser"))
	password := getPoolEnv("PGPASSWORD", getPoolEnv("DB_PASSWORD", "devpass"))
	dbname := getPoolEnv("PGDATABASE", getPoolEnv("DB_NAME", "starter_kit_db"))
	sslmode := getPoolEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

func getPoolEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
