package database

import (
	"context"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
)

// N+1 detection context key
type n1DetectorKey struct{}

// N1QueryTracker tracks queries per request for N+1 detection
type N1QueryTracker struct {
	mu       sync.Mutex
	queries  map[string]int // table -> query count
	patterns map[string]int // query pattern -> count
}

// NewN1QueryTracker creates a new N+1 query tracker
func NewN1QueryTracker() *N1QueryTracker {
	return &N1QueryTracker{
		queries:  make(map[string]int),
		patterns: make(map[string]int),
	}
}

// Track records a query for N+1 detection
func (t *N1QueryTracker) Track(sql string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Extract table name from query
	table := extractTableName(sql)
	if table != "" {
		t.queries[table]++
	}

	// Track by query pattern (normalize values)
	pattern := normalizeQueryPattern(sql)
	t.patterns[pattern]++
}

// GetN1Violations returns tables that were queried more than threshold times
func (t *N1QueryTracker) GetN1Violations(threshold int) map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()

	violations := make(map[string]int)
	for table, count := range t.queries {
		if count > threshold {
			violations[table] = count
		}
	}
	return violations
}

// GetPatternViolations returns query patterns that were executed more than threshold times
func (t *N1QueryTracker) GetPatternViolations(threshold int) map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()

	violations := make(map[string]int)
	for pattern, count := range t.patterns {
		if count > threshold {
			violations[pattern] = count
		}
	}
	return violations
}

// WithN1Detection adds N+1 detection to the context
func WithN1Detection(ctx context.Context) context.Context {
	return context.WithValue(ctx, n1DetectorKey{}, NewN1QueryTracker())
}

// GetN1Tracker retrieves the N+1 tracker from context
func GetN1Tracker(ctx context.Context) *N1QueryTracker {
	if tracker, ok := ctx.Value(n1DetectorKey{}).(*N1QueryTracker); ok {
		return tracker
	}
	return nil
}

// CheckN1Violations checks for N+1 query violations and logs warnings
func CheckN1Violations(ctx context.Context, threshold int) {
	tracker := GetN1Tracker(ctx)
	if tracker == nil {
		return
	}

	violations := tracker.GetPatternViolations(threshold)
	if len(violations) > 0 {
		log.Warn().
			Interface("violations", violations).
			Int("threshold", threshold).
			Msg("potential N+1 query detected - consider using Preload() or eager loading")
	}
}

// extractTableName extracts the main table name from a SQL query
func extractTableName(sql string) string {
	sql = strings.ToLower(sql)

	// SELECT ... FROM table
	if strings.HasPrefix(sql, "select") {
		re := regexp.MustCompile(`from\s+"?(\w+)"?`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return matches[1]
		}
	}

	// INSERT INTO table
	if strings.HasPrefix(sql, "insert") {
		re := regexp.MustCompile(`into\s+"?(\w+)"?`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return matches[1]
		}
	}

	// UPDATE table
	if strings.HasPrefix(sql, "update") {
		re := regexp.MustCompile(`update\s+"?(\w+)"?`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return matches[1]
		}
	}

	// DELETE FROM table
	if strings.HasPrefix(sql, "delete") {
		re := regexp.MustCompile(`from\s+"?(\w+)"?`)
		if matches := re.FindStringSubmatch(sql); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// normalizeQueryPattern normalizes a query by replacing values with placeholders
func normalizeQueryPattern(sql string) string {
	// Replace numeric values
	re := regexp.MustCompile(`\b\d+\b`)
	sql = re.ReplaceAllString(sql, "?")

	// Replace quoted strings
	re = regexp.MustCompile(`'[^']*'`)
	sql = re.ReplaceAllString(sql, "?")

	// Remove extra whitespace
	re = regexp.MustCompile(`\s+`)
	sql = re.ReplaceAllString(sql, " ")

	return strings.TrimSpace(sql)
}

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
	// N1DetectionEnabled enables N+1 query detection (development only)
	N1DetectionEnabled bool
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

	// Enable N+1 detection in development only
	n1Detection := false
	env := os.Getenv("GO_ENV")
	if env == "" || env == "development" || env == "dev" {
		n1Detection = getEnvBool("DB_N1_DETECTION", true) // Enabled by default in dev
	}

	return &QueryProfiler{
		SlowQueryThreshold:        threshold,
		Logger:                    log.With().Str("component", "gorm").Logger(),
		LogLevel:                  logLevel,
		IgnoreRecordNotFoundError: true,
		N1DetectionEnabled:        n1Detection,
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

	// Track query for N+1 detection (development only)
	if p.N1DetectionEnabled {
		if tracker := GetN1Tracker(ctx); tracker != nil {
			tracker.Track(sql)
		}
	}

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
