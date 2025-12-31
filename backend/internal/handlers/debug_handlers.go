// Package handlers provides HTTP request handlers for the API.
package handlers

import (
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/response"

	"github.com/go-chi/chi/v5"
)

// DebugHandlers provides debug endpoints for development.
// These endpoints should only be enabled in development mode.
type DebugHandlers struct {
	startTime time.Time
	router    chi.Router
}

// NewDebugHandlers creates a new DebugHandlers instance.
func NewDebugHandlers(router chi.Router) *DebugHandlers {
	return &DebugHandlers{
		startTime: time.Now(),
		router:    router,
	}
}

// SanitizedConfig returns configuration with sensitive values redacted.
type SanitizedConfig struct {
	Server    map[string]interface{} `json:"server"`
	Database  map[string]interface{} `json:"database"`
	RateLimit map[string]interface{} `json:"rate_limit"`
	Logging   map[string]interface{} `json:"logging"`
	Features  map[string]interface{} `json:"features"`
}

// GetConfig returns the current configuration with sensitive values redacted.
// GET /api/debug/config
func (h *DebugHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := SanitizedConfig{
		Server: map[string]interface{}{
			"port":  getEnvSafe("API_PORT", "8080"),
			"host":  getEnvSafe("API_HOST", "0.0.0.0"),
			"env":   getEnvSafe("GO_ENV", "development"),
			"debug": getEnvSafe("DEBUG", "false"),
		},
		Database: map[string]interface{}{
			"host":     getEnvSafe("PGHOST", getEnvSafe("DB_HOST", "localhost")),
			"port":     getEnvSafe("PGPORT", getEnvSafe("DB_PORT", "5432")),
			"name":     getEnvSafe("PGDATABASE", getEnvSafe("DB_NAME", "starter_kit_db")),
			"sslmode":  getEnvSafe("DB_SSLMODE", "disable"),
			"user":     getEnvSafe("PGUSER", getEnvSafe("DB_USER", "[set]")),
			"password": "[redacted]",
		},
		RateLimit: map[string]interface{}{
			"enabled":         getEnvSafe("RATE_LIMIT_ENABLED", "true"),
			"ip_per_minute":   getEnvSafe("RATE_LIMIT_IP_PER_MINUTE", "60"),
			"auth_per_minute": getEnvSafe("RATE_LIMIT_AUTH_PER_MINUTE", "5"),
		},
		Logging: map[string]interface{}{
			"level":  getEnvSafe("LOG_LEVEL", "info"),
			"pretty": getEnvSafe("LOG_PRETTY", "false"),
		},
		Features: map[string]interface{}{
			"cache_enabled": getEnvSafe("CACHE_ENABLED", "false"),
			"jobs_enabled":  getEnvSafe("JOBS_ENABLED", "false"),
			"csrf_enabled":  getEnvSafe("CSRF_ENABLED", "true"),
		},
	}

	response.Success(w, "Configuration retrieved", config)
}

// RouteInfo represents information about a registered route.
type RouteInfo struct {
	Method  string `json:"method"`
	Pattern string `json:"pattern"`
}

// GetRoutes returns all registered routes.
// GET /api/debug/routes
func (h *DebugHandlers) GetRoutes(w http.ResponseWriter, r *http.Request) {
	var routes []RouteInfo

	// Walk the router tree to get all routes
	chi.Walk(h.router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes = append(routes, RouteInfo{
			Method:  method,
			Pattern: route,
		})
		return nil
	})

	response.Success(w, "Routes retrieved", map[string]interface{}{
		"count":  len(routes),
		"routes": routes,
	})
}

// RuntimeStats represents runtime statistics.
type RuntimeStats struct {
	// General
	Uptime        string `json:"uptime"`
	GoVersion     string `json:"go_version"`
	NumCPU        int    `json:"num_cpu"`
	NumGoroutines int    `json:"num_goroutines"`

	// Memory
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
	HeapObjectsK float64 `json:"heap_objects_k"`
	HeapAllocMB  float64 `json:"heap_alloc_mb"`
	HeapInUseMB  float64 `json:"heap_in_use_mb"`

	// Database Pool (if available)
	DBPool *PoolStats `json:"db_pool,omitempty"`

	// Cache (if available)
	CacheAvailable bool `json:"cache_available"`
}

// PoolStats represents database connection pool statistics.
type PoolStats struct {
	OpenConnections int    `json:"open_connections"`
	InUse           int    `json:"in_use"`
	Idle            int    `json:"idle"`
	WaitCount       int64  `json:"wait_count"`
	WaitDuration    string `json:"wait_duration"`
	MaxOpenConns    int    `json:"max_open_conns"`
}

// GetStats returns runtime statistics.
// GET /api/debug/stats
func (h *DebugHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := RuntimeStats{
		Uptime:        time.Since(h.startTime).Round(time.Second).String(),
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),

		AllocMB:      float64(memStats.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(memStats.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(memStats.Sys) / 1024 / 1024,
		NumGC:        memStats.NumGC,
		HeapObjectsK: float64(memStats.HeapObjects) / 1000,
		HeapAllocMB:  float64(memStats.HeapAlloc) / 1024 / 1024,
		HeapInUseMB:  float64(memStats.HeapInuse) / 1024 / 1024,

		CacheAvailable: cache.IsAvailable(),
	}

	// Get database pool stats if available
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			dbStats := sqlDB.Stats()
			stats.DBPool = &PoolStats{
				OpenConnections: dbStats.OpenConnections,
				InUse:           dbStats.InUse,
				Idle:            dbStats.Idle,
				WaitCount:       dbStats.WaitCount,
				WaitDuration:    dbStats.WaitDuration.String(),
				MaxOpenConns:    dbStats.MaxOpenConnections,
			}
		}
	}

	response.Success(w, "Stats retrieved", stats)
}

// TriggerGC triggers garbage collection (for debugging memory issues).
// POST /api/debug/gc
func (h *DebugHandlers) TriggerGC(w http.ResponseWriter, r *http.Request) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	runtime.GC()

	runtime.ReadMemStats(&after)

	response.Success(w, "Garbage collection triggered", map[string]interface{}{
		"freed_mb":       float64(before.Alloc-after.Alloc) / 1024 / 1024,
		"heap_before_mb": float64(before.HeapAlloc) / 1024 / 1024,
		"heap_after_mb":  float64(after.HeapAlloc) / 1024 / 1024,
	})
}

// RegisterRoutes registers debug routes on the given router.
// These should only be enabled in development mode.
func (h *DebugHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/debug", func(r chi.Router) {
		r.Get("/config", h.GetConfig)
		r.Get("/routes", h.GetRoutes)
		r.Get("/stats", h.GetStats)
		r.Post("/gc", h.TriggerGC)
	})
}

// getEnvSafe returns the environment variable value or a default.
func getEnvSafe(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		// Redact sensitive values
		if isSensitiveKey(key) {
			return "[redacted]"
		}
		return value
	}
	return fallback
}

// isSensitiveKey returns true if the key contains sensitive data.
func isSensitiveKey(key string) bool {
	sensitive := []string{
		"PASSWORD", "SECRET", "KEY", "TOKEN", "CREDENTIAL",
		"PRIVATE", "AUTH", "API_KEY",
	}
	upperKey := strings.ToUpper(key)
	for _, s := range sensitive {
		if strings.Contains(upperKey, s) {
			return true
		}
	}
	return false
}
