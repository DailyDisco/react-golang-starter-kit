package database

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
)

// QueryProfiler is a GORM logger that logs slow queries.
type QueryProfiler struct {
	// SlowQueryThreshold is the duration above which a query is considered slow
	SlowQueryThreshold time.Duration
	// Logger is the zerolog logger instance
	Logger zerolog.Logger
	// LogLevel is the GORM log level
	LogLevel logger.LogLevel
	// IgnoreRecordNotFoundError ignores ErrRecordNotFound errors
	IgnoreRecordNotFoundError bool
}

// NewQueryProfiler creates a new query profiler with the given threshold.
// Default threshold is 200ms.
func NewQueryProfiler() *QueryProfiler {
	threshold := 200 * time.Millisecond
	if envThreshold := os.Getenv("DB_SLOW_QUERY_THRESHOLD"); envThreshold != "" {
		if d, err := time.ParseDuration(envThreshold); err == nil {
			threshold = d
		}
	}

	logLevel := logger.Warn
	if os.Getenv("DB_LOG_ALL_QUERIES") == "true" {
		logLevel = logger.Info
	}

	return &QueryProfiler{
		SlowQueryThreshold:        threshold,
		Logger:                    log.With().Str("component", "gorm").Logger(),
		LogLevel:                  logLevel,
		IgnoreRecordNotFoundError: true,
	}
}

// LogMode implements the gorm logger.Interface
func (p *QueryProfiler) LogMode(level logger.LogLevel) logger.Interface {
	newProfiler := *p
	newProfiler.LogLevel = level
	return &newProfiler
}

// Info implements the gorm logger.Interface
func (p *QueryProfiler) Info(ctx context.Context, msg string, data ...interface{}) {
	if p.LogLevel >= logger.Info {
		p.Logger.Info().Msgf(msg, data...)
	}
}

// Warn implements the gorm logger.Interface
func (p *QueryProfiler) Warn(ctx context.Context, msg string, data ...interface{}) {
	if p.LogLevel >= logger.Warn {
		p.Logger.Warn().Msgf(msg, data...)
	}
}

// Error implements the gorm logger.Interface
func (p *QueryProfiler) Error(ctx context.Context, msg string, data ...interface{}) {
	if p.LogLevel >= logger.Error {
		p.Logger.Error().Msgf(msg, data...)
	}
}

// Trace implements the gorm logger.Interface - this is where queries are logged
func (p *QueryProfiler) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if p.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Always log errors (except RecordNotFound if configured)
	if err != nil {
		if p.IgnoreRecordNotFoundError && err.Error() == "record not found" {
			return
		}
		p.Logger.Error().
			Err(err).
			Str("sql", truncateSQL(sql, 500)).
			Int64("rows", rows).
			Dur("duration", elapsed).
			Msg("query error")
		return
	}

	// Log slow queries as warnings
	if elapsed > p.SlowQueryThreshold {
		p.Logger.Warn().
			Str("sql", truncateSQL(sql, 500)).
			Int64("rows", rows).
			Dur("duration", elapsed).
			Dur("threshold", p.SlowQueryThreshold).
			Msg("slow query detected")
		return
	}

	// Log all queries if in Info mode
	if p.LogLevel >= logger.Info {
		p.Logger.Debug().
			Str("sql", truncateSQL(sql, 200)).
			Int64("rows", rows).
			Dur("duration", elapsed).
			Msg("query executed")
	}
}

// truncateSQL truncates a SQL string to the given length
func truncateSQL(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}

// ProfilerConfig holds configuration for the query profiler
type ProfilerConfig struct {
	// Enabled determines if query profiling is enabled
	Enabled bool
	// SlowQueryThreshold is the duration above which a query is considered slow
	SlowQueryThreshold time.Duration
	// LogAllQueries logs all queries, not just slow ones
	LogAllQueries bool
}

// LoadProfilerConfig loads profiler configuration from environment variables
func LoadProfilerConfig() *ProfilerConfig {
	config := &ProfilerConfig{
		Enabled:            getEnvBool("DB_PROFILER_ENABLED", true),
		SlowQueryThreshold: 200 * time.Millisecond,
		LogAllQueries:      getEnvBool("DB_LOG_ALL_QUERIES", false),
	}

	if envThreshold := os.Getenv("DB_SLOW_QUERY_THRESHOLD"); envThreshold != "" {
		if d, err := time.ParseDuration(envThreshold); err == nil {
			config.SlowQueryThreshold = d
		}
	}

	return config
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}
