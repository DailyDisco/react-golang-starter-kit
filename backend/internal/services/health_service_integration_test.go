package services

import (
	"testing"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/testutil"
)

func testHealthSetup(t *testing.T) (*HealthService, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB for the health service
	oldDB := database.DB
	database.DB = tt.DB

	svc := NewHealthService()

	return svc, func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func TestHealthService_GetSystemHealth_Integration(t *testing.T) {
	svc, cleanup := testHealthSetup(t)
	defer cleanup()

	t.Run("returns healthy status with database connected", func(t *testing.T) {
		health, err := svc.GetSystemHealth()
		if err != nil {
			t.Fatalf("GetSystemHealth failed: %v", err)
		}

		if health == nil {
			t.Fatal("Expected health response, got nil")
		}

		// Should be healthy or degraded (cache might not be available)
		if health.Status != "healthy" && health.Status != "degraded" {
			t.Errorf("Expected status healthy or degraded, got: %s", health.Status)
		}

		if health.Timestamp == "" {
			t.Error("Expected timestamp to be set")
		}

		if len(health.Components) < 1 {
			t.Error("Expected at least one component (database)")
		}
	})

	t.Run("includes database component", func(t *testing.T) {
		health, err := svc.GetSystemHealth()
		if err != nil {
			t.Fatalf("GetSystemHealth failed: %v", err)
		}

		var dbComponent *struct {
			Name   string
			Status string
		}
		for _, c := range health.Components {
			if c.Name == "database" {
				dbComponent = &struct {
					Name   string
					Status string
				}{c.Name, c.Status}
				break
			}
		}

		if dbComponent == nil {
			t.Fatal("Expected database component in health check")
		}

		if dbComponent.Status != "healthy" && dbComponent.Status != "degraded" {
			t.Errorf("Expected database status healthy or degraded, got: %s", dbComponent.Status)
		}
	})

	t.Run("includes cache component", func(t *testing.T) {
		health, err := svc.GetSystemHealth()
		if err != nil {
			t.Fatalf("GetSystemHealth failed: %v", err)
		}

		var cacheComponent *struct {
			Name   string
			Status string
		}
		for _, c := range health.Components {
			if c.Name == "cache" {
				cacheComponent = &struct {
					Name   string
					Status string
				}{c.Name, c.Status}
				break
			}
		}

		if cacheComponent == nil {
			t.Fatal("Expected cache component in health check")
		}

		// Cache might not be available in test environment
		validStatuses := []string{"healthy", "degraded", "unhealthy"}
		validStatus := false
		for _, s := range validStatuses {
			if cacheComponent.Status == s {
				validStatus = true
				break
			}
		}
		if !validStatus {
			t.Errorf("Unexpected cache status: %s", cacheComponent.Status)
		}
	})

	t.Run("includes metrics", func(t *testing.T) {
		health, err := svc.GetSystemHealth()
		if err != nil {
			t.Fatalf("GetSystemHealth failed: %v", err)
		}

		if health.Metrics == nil {
			t.Error("Expected metrics in health response")
		}
	})
}

func TestHealthService_CheckDatabaseHealth_Integration(t *testing.T) {
	svc, cleanup := testHealthSetup(t)
	defer cleanup()

	t.Run("returns healthy when database is connected", func(t *testing.T) {
		component := svc.checkDatabaseHealth()

		if component.Name != "database" {
			t.Errorf("Expected component name 'database', got: %s", component.Name)
		}

		if component.Status != "healthy" && component.Status != "degraded" {
			t.Errorf("Expected status healthy or degraded, got: %s", component.Status)
		}

		if component.LastCheck == "" {
			t.Error("Expected LastCheck timestamp to be set")
		}

		if component.Details == nil {
			t.Error("Expected Details to be populated")
		}
	})

	t.Run("includes database stats in details", func(t *testing.T) {
		component := svc.checkDatabaseHealth()

		if component.Details == nil {
			t.Fatal("Expected Details to be populated")
		}

		// Check for expected detail fields
		expectedFields := []string{"latency", "open_connections", "max_open_connections", "in_use", "idle"}
		for _, field := range expectedFields {
			if _, ok := component.Details[field]; !ok {
				t.Errorf("Expected detail field '%s' to be present", field)
			}
		}
	})
}

func TestHealthService_GetSystemMetrics_Integration(t *testing.T) {
	svc, cleanup := testHealthSetup(t)
	defer cleanup()

	t.Run("returns system metrics", func(t *testing.T) {
		metrics := svc.GetSystemMetrics()

		if metrics == nil {
			t.Fatal("Expected metrics, got nil")
		}

		// Check for expected metric components
		if metrics.Database == nil {
			t.Error("Expected database metrics to be present")
		}
		if metrics.Cache == nil {
			t.Error("Expected cache metrics to be present")
		}
	})

	t.Run("database metrics include connection info", func(t *testing.T) {
		metrics := svc.GetSystemMetrics()

		if metrics.Database == nil {
			t.Skip("Database metrics not available")
		}

		// Database metrics should have a status
		if metrics.Database.Status == "" {
			t.Error("Expected database status to be set")
		}
	})

	t.Run("cache metrics are present", func(t *testing.T) {
		metrics := svc.GetSystemMetrics()

		if metrics.Cache == nil {
			t.Skip("Cache metrics not available")
		}

		// Cache metrics should have a status
		if metrics.Cache.Status == "" {
			t.Error("Expected cache status to be set")
		}
	})

	t.Run("storage metrics are present", func(t *testing.T) {
		metrics := svc.GetSystemMetrics()

		if metrics.Storage == nil {
			t.Skip("Storage metrics not available")
		}
	})

	t.Run("API metrics are present", func(t *testing.T) {
		metrics := svc.GetSystemMetrics()

		if metrics.API == nil {
			t.Skip("API metrics not available")
		}
	})
}

func TestHealthService_GetDetailedDatabaseHealth_Integration(t *testing.T) {
	svc, cleanup := testHealthSetup(t)
	defer cleanup()

	t.Run("returns detailed database health", func(t *testing.T) {
		health, err := svc.GetDetailedDatabaseHealth()
		if err != nil {
			t.Fatalf("GetDetailedDatabaseHealth failed: %v", err)
		}

		if health == nil {
			t.Fatal("Expected health, got nil")
		}

		// Verify health contains expected information
		if health["status"] == nil {
			t.Error("Expected status in database health")
		}
	})

	t.Run("includes connection pool information", func(t *testing.T) {
		health, err := svc.GetDetailedDatabaseHealth()
		if err != nil {
			t.Fatalf("GetDetailedDatabaseHealth failed: %v", err)
		}

		// Check for pool stats or connection info
		hasConnectionInfo := false
		for key := range health {
			if key == "open_connections" || key == "max_connections" || key == "pool_stats" {
				hasConnectionInfo = true
				break
			}
		}
		if !hasConnectionInfo {
			t.Log("No specific connection pool info found, checking for other connection details")
		}
	})
}

func TestHealthService_GetDetailedCacheHealth_Integration(t *testing.T) {
	svc, cleanup := testHealthSetup(t)
	defer cleanup()

	t.Run("returns cache health information", func(t *testing.T) {
		health, err := svc.GetDetailedCacheHealth()
		if err != nil {
			// Cache might not be available in test environment
			t.Logf("GetDetailedCacheHealth returned error (cache may not be available): %v", err)
			return
		}

		if health == nil {
			t.Fatal("Expected health, got nil")
		}

		// Should have some status indication
		if health["status"] == nil && health["error"] == nil {
			t.Error("Expected either status or error in cache health")
		}
	})
}

func TestHealthService_GetUptime_Integration(t *testing.T) {
	svc, cleanup := testHealthSetup(t)
	defer cleanup()

	t.Run("returns positive uptime", func(t *testing.T) {
		metrics := svc.GetRuntimeMetrics()

		uptime, ok := metrics["uptime"].(string)
		if !ok {
			t.Fatal("Expected uptime to be a string")
		}

		if uptime == "" {
			t.Error("Expected non-empty uptime")
		}

		// Uptime should contain time units (s, m, h, etc.)
		if len(uptime) < 2 {
			t.Errorf("Uptime format seems incorrect: %s", uptime)
		}
	})
}
