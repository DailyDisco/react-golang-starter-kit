package services

import (
	"context"
	"fmt"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// HealthService handles system health monitoring
type HealthService struct {
	startTime time.Time
}

// NewHealthService creates a new health service instance
func NewHealthService() *HealthService {
	return &HealthService{
		startTime: time.Now(),
	}
}

// db returns the database connection - accessed at runtime to avoid nil issues
func (s *HealthService) db() *gorm.DB {
	return database.DB
}

// GetSystemHealth returns overall system health status
func (s *HealthService) GetSystemHealth() (*models.SystemHealthResponse, error) {
	components := []models.HealthComponent{}
	overallStatus := "healthy"

	// Check database health
	dbHealth := s.checkDatabaseHealth()
	components = append(components, dbHealth)
	if dbHealth.Status == "unhealthy" {
		overallStatus = "unhealthy"
	} else if dbHealth.Status == "degraded" && overallStatus == "healthy" {
		overallStatus = "degraded"
	}

	// Check cache health
	cacheHealth := s.checkCacheHealth()
	components = append(components, cacheHealth)
	if cacheHealth.Status == "unhealthy" && overallStatus == "healthy" {
		overallStatus = "degraded" // Cache is not critical
	}

	// Get metrics
	metrics := s.GetSystemMetrics()

	return &models.SystemHealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now().Format(time.RFC3339),
		Components: components,
		Metrics:    metrics,
	}, nil
}

// checkDatabaseHealth checks database connection and performance
func (s *HealthService) checkDatabaseHealth() models.HealthComponent {
	start := time.Now()

	// Ping the database
	sqlDB, err := s.db().DB()
	if err != nil {
		return models.HealthComponent{
			Name:      "database",
			Status:    "unhealthy",
			Message:   fmt.Sprintf("Failed to get database connection: %v", err),
			LastCheck: time.Now().Format(time.RFC3339),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return models.HealthComponent{
			Name:      "database",
			Status:    "unhealthy",
			Message:   fmt.Sprintf("Database ping failed: %v", err),
			LastCheck: time.Now().Format(time.RFC3339),
		}
	}

	latency := time.Since(start)
	status := "healthy"
	message := "Database connection is healthy"

	if latency > 500*time.Millisecond {
		status = "degraded"
		message = "Database response time is slow"
	}

	stats := sqlDB.Stats()

	return models.HealthComponent{
		Name:      "database",
		Status:    status,
		Message:   message,
		Latency:   latency.String(),
		LastCheck: time.Now().Format(time.RFC3339),
		Details: map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"max_open":         stats.MaxOpenConnections,
		},
	}
}

// checkCacheHealth checks Redis/cache connection
func (s *HealthService) checkCacheHealth() models.HealthComponent {
	start := time.Now()

	// Check if cache is available
	cacheInstance := cache.Instance()
	if cacheInstance == nil || !cache.IsAvailable() {
		return models.HealthComponent{
			Name:      "cache",
			Status:    "unavailable",
			Message:   "Cache is not configured or unavailable",
			LastCheck: time.Now().Format(time.RFC3339),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Ping cache
	if err := cacheInstance.Ping(ctx); err != nil {
		return models.HealthComponent{
			Name:      "cache",
			Status:    "unhealthy",
			Message:   fmt.Sprintf("Cache ping failed: %v", err),
			LastCheck: time.Now().Format(time.RFC3339),
		}
	}

	latency := time.Since(start)
	status := "healthy"
	message := "Cache connection is healthy"

	if latency > 100*time.Millisecond {
		status = "degraded"
		message = "Cache response time is slow"
	}

	return models.HealthComponent{
		Name:      "cache",
		Status:    status,
		Message:   message,
		Latency:   latency.String(),
		LastCheck: time.Now().Format(time.RFC3339),
		Details:   map[string]interface{}{"type": "redis/memory"},
	}
}

// GetSystemMetrics returns detailed system metrics
func (s *HealthService) GetSystemMetrics() *models.SystemHealthMetrics {
	return &models.SystemHealthMetrics{
		Database: s.getDatabaseMetrics(),
		Cache:    s.getCacheMetrics(),
		Storage:  s.getStorageMetrics(),
		API:      s.getAPIMetrics(),
	}
}

func (s *HealthService) getDatabaseMetrics() *models.DatabaseMetrics {
	sqlDB, err := s.db().DB()
	if err != nil {
		return &models.DatabaseMetrics{Status: "unhealthy"}
	}

	stats := sqlDB.Stats()

	// Get average query time (simplified - in production, you'd track this)
	start := time.Now()
	s.db().Raw("SELECT 1").Scan(&struct{}{})
	avgQueryTime := time.Since(start)

	return &models.DatabaseMetrics{
		Status:            "healthy",
		ConnectionsActive: stats.InUse,
		ConnectionsIdle:   stats.Idle,
		ConnectionsMax:    stats.MaxOpenConnections,
		AvgQueryTime:      avgQueryTime.String(),
		SlowQueries:       0, // Would need query logging to track
		Uptime:            time.Since(s.startTime).String(),
	}
}

func (s *HealthService) getCacheMetrics() *models.CacheMetrics {
	cacheInstance := cache.Instance()
	if cacheInstance == nil || !cache.IsAvailable() {
		return &models.CacheMetrics{Status: "unavailable"}
	}

	ctx := context.Background()

	// Test cache connectivity
	if err := cacheInstance.Ping(ctx); err != nil {
		return &models.CacheMetrics{Status: "unhealthy"}
	}

	// For basic cache metrics, we return what we can without direct Redis access
	// More detailed metrics would require exposing the underlying Redis client
	return &models.CacheMetrics{
		Status:      "healthy",
		MemoryUsed:  "N/A",
		MemoryMax:   "N/A",
		HitRate:     0,
		Keys:        0,
		Connections: 0,
	}
}

func (s *HealthService) getStorageMetrics() *models.StorageMetrics {
	var totalFiles int64
	var totalSize int64

	// Get file storage stats
	s.db().Model(&models.File{}).Count(&totalFiles)
	s.db().Model(&models.File{}).Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)

	return &models.StorageMetrics{
		Status:    "healthy",
		Used:      formatBytes(totalSize),
		Available: "N/A", // Would need disk space check
		Total:     "N/A",
		UsedPct:   0,
		FileCount: totalFiles,
	}
}

func (s *HealthService) getAPIMetrics() *models.APIMetrics {
	// These would typically come from a metrics collection system
	// For now, return placeholder data
	return &models.APIMetrics{
		RequestsPerMinute: 0, // Would need request counter
		AvgResponseTime:   "N/A",
		P50ResponseTime:   "N/A",
		P95ResponseTime:   "N/A",
		P99ResponseTime:   "N/A",
		ErrorRate:         0,
	}
}

// GetDetailedDatabaseHealth returns detailed database health info
func (s *HealthService) GetDetailedDatabaseHealth() (map[string]interface{}, error) {
	sqlDB, err := s.db().DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()

	// Run a simple query to test
	start := time.Now()
	var result int
	s.db().Raw("SELECT 1").Scan(&result)
	queryTime := time.Since(start)

	return map[string]interface{}{
		"status":               "healthy",
		"query_time":           queryTime.String(),
		"open_connections":     stats.OpenConnections,
		"in_use_connections":   stats.InUse,
		"idle_connections":     stats.Idle,
		"max_open_connections": stats.MaxOpenConnections,
		"max_idle_connections": stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}, nil
}

// GetDetailedCacheHealth returns detailed cache health info
func (s *HealthService) GetDetailedCacheHealth() (map[string]interface{}, error) {
	cacheInstance := cache.Instance()
	if cacheInstance == nil || !cache.IsAvailable() {
		return map[string]interface{}{
			"status":  "unavailable",
			"message": "Cache is not configured or unavailable",
		}, nil
	}

	ctx := context.Background()

	// Ping
	start := time.Now()
	if err := cacheInstance.Ping(ctx); err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}, nil
	}
	pingTime := time.Since(start)

	details := map[string]interface{}{
		"status":    "healthy",
		"ping_time": pingTime.String(),
		"type":      "redis/memory",
		"available": cache.IsAvailable(),
	}

	return details, nil
}

// Helper functions

func parseRedisInfo(info string) map[string]interface{} {
	result := make(map[string]interface{})

	lines := splitLines(info)
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		parts := splitKeyValue(line, ":")
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitKeyValue(s string, sep string) []string {
	idx := -1
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetRuntimeMetrics returns Go runtime metrics
func (s *HealthService) GetRuntimeMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"memory_alloc":   formatBytes(int64(m.Alloc)),
		"memory_sys":     formatBytes(int64(m.Sys)),
		"memory_heap":    formatBytes(int64(m.HeapAlloc)),
		"gc_runs":        m.NumGC,
		"gc_pause_total": time.Duration(m.PauseTotalNs).String(),
		"go_version":     runtime.Version(),
		"num_cpu":        runtime.NumCPU(),
		"uptime":         time.Since(s.startTime).String(),
	}
}
