package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
)

// ============ getEnvSafe Tests ============

func TestGetEnvSafe_WithValue(t *testing.T) {
	// Use a key name that doesn't contain any sensitive keywords
	key := "TEST_DEBUG_PORT"
	t.Setenv(key, "test_value")

	result := getEnvSafe(key, "fallback")
	if result != "test_value" {
		t.Errorf("getEnvSafe() = %q, want %q", result, "test_value")
	}
}

func TestGetEnvSafe_WithFallback(t *testing.T) {
	key := "TEST_GET_ENV_SAFE_NONEXISTENT"
	os.Unsetenv(key)

	result := getEnvSafe(key, "fallback")
	if result != "fallback" {
		t.Errorf("getEnvSafe() = %q, want %q", result, "fallback")
	}
}

func TestGetEnvSafe_SensitiveKeyRedacted(t *testing.T) {
	key := "TEST_DATABASE_PASSWORD"
	t.Setenv(key, "super_secret")

	result := getEnvSafe(key, "fallback")
	if result != "[redacted]" {
		t.Errorf("getEnvSafe() with sensitive key = %q, want %q", result, "[redacted]")
	}
}

func TestGetEnvSafe_SensitiveKeyVariants(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"API_KEY", "TEST_API_KEY"},
		{"SECRET", "TEST_SECRET_VALUE"},
		{"TOKEN", "TEST_AUTH_TOKEN"},
		{"PASSWORD", "TEST_DB_PASSWORD"},
		{"CREDENTIAL", "TEST_CREDENTIAL"},
		{"PRIVATE", "TEST_PRIVATE_KEY"},
		{"AUTH", "TEST_AUTH_HEADER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, "sensitive_value")

			result := getEnvSafe(tt.key, "fallback")
			if result != "[redacted]" {
				t.Errorf("getEnvSafe(%s) = %q, want [redacted]", tt.key, result)
			}
		})
	}
}

// ============ isSensitiveKey Tests ============

func TestIsSensitiveKey_SensitiveKeys(t *testing.T) {
	sensitiveKeys := []string{
		"DATABASE_PASSWORD",
		"JWT_SECRET",
		"API_KEY",
		"AUTH_TOKEN",
		"PRIVATE_KEY",
		"CREDENTIAL_FILE",
		"secret_value",
		"my_token",
	}

	for _, key := range sensitiveKeys {
		t.Run(key, func(t *testing.T) {
			if !isSensitiveKey(key) {
				t.Errorf("isSensitiveKey(%q) = false, want true", key)
			}
		})
	}
}

func TestIsSensitiveKey_NonSensitiveKeys(t *testing.T) {
	nonSensitiveKeys := []string{
		"API_HOST",
		"API_PORT",
		"DB_HOST",
		"LOG_LEVEL",
		"GO_ENV",
		"DEBUG",
		"CACHE_ENABLED",
	}

	for _, key := range nonSensitiveKeys {
		t.Run(key, func(t *testing.T) {
			if isSensitiveKey(key) {
				t.Errorf("isSensitiveKey(%q) = true, want false", key)
			}
		})
	}
}

func TestIsSensitiveKey_CaseInsensitive(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"password", true},
		{"PASSWORD", true},
		{"Password", true},
		{"PaSsWoRd", true},
		{"secret", true},
		{"SECRET", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isSensitiveKey(tt.key); got != tt.want {
				t.Errorf("isSensitiveKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// ============ NewDebugHandlers Tests ============

func TestNewDebugHandlers(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	if h == nil {
		t.Fatal("NewDebugHandlers() returned nil")
	}

	if h.router == nil {
		t.Error("NewDebugHandlers() did not set router")
	}

	if h.startTime.IsZero() {
		t.Error("NewDebugHandlers() did not set startTime")
	}
}

// ============ GetConfig Handler Tests ============

func TestDebugHandlers_GetConfig(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	// Set some env vars for testing
	t.Setenv("API_PORT", "9999")
	t.Setenv("GO_ENV", "test")

	req := httptest.NewRequest(http.MethodGet, "/debug/config", nil)
	w := httptest.NewRecorder()

	h.GetConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetConfig() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that data exists
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing 'data' field")
	}

	// Check server config
	server, ok := data["server"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing 'server' config")
	}

	if server["port"] != "9999" {
		t.Errorf("server.port = %v, want 9999", server["port"])
	}

	if server["env"] != "test" {
		t.Errorf("server.env = %v, want test", server["env"])
	}

	// Check database password is redacted
	db, ok := data["database"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing 'database' config")
	}

	if db["password"] != "[redacted]" {
		t.Errorf("database.password = %v, want [redacted]", db["password"])
	}
}

func TestDebugHandlers_GetConfig_AllSections(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/config", nil)
	w := httptest.NewRecorder()

	h.GetConfig(w, req)

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data := resp["data"].(map[string]interface{})

	expectedSections := []string{"server", "database", "rate_limit", "logging", "features"}
	for _, section := range expectedSections {
		if _, ok := data[section]; !ok {
			t.Errorf("GetConfig() missing section %q", section)
		}
	}
}

// ============ GetStats Handler Tests ============

func TestDebugHandlers_GetStats(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/stats", nil)
	w := httptest.NewRecorder()

	h.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetStats() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing 'data' field")
	}

	// Check expected fields
	expectedFields := []string{
		"uptime",
		"go_version",
		"num_cpu",
		"num_goroutines",
		"alloc_mb",
		"total_alloc_mb",
		"sys_mb",
		"num_gc",
	}

	for _, field := range expectedFields {
		if _, ok := data[field]; !ok {
			t.Errorf("GetStats() missing field %q", field)
		}
	}
}

func TestDebugHandlers_GetStats_NumCPUPositive(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/stats", nil)
	w := httptest.NewRecorder()

	h.GetStats(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	data := resp["data"].(map[string]interface{})
	numCPU := data["num_cpu"].(float64)

	if numCPU < 1 {
		t.Errorf("GetStats() num_cpu = %v, should be >= 1", numCPU)
	}
}

func TestDebugHandlers_GetStats_NumGoroutinesPositive(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/stats", nil)
	w := httptest.NewRecorder()

	h.GetStats(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	data := resp["data"].(map[string]interface{})
	numGoroutines := data["num_goroutines"].(float64)

	if numGoroutines < 1 {
		t.Errorf("GetStats() num_goroutines = %v, should be >= 1", numGoroutines)
	}
}

// ============ TriggerGC Handler Tests ============

func TestDebugHandlers_TriggerGC(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodPost, "/debug/gc", nil)
	w := httptest.NewRecorder()

	h.TriggerGC(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("TriggerGC() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing 'data' field")
	}

	// Check expected fields
	expectedFields := []string{"freed_mb", "heap_before_mb", "heap_after_mb"}
	for _, field := range expectedFields {
		if _, ok := data[field]; !ok {
			t.Errorf("TriggerGC() missing field %q", field)
		}
	}
}

// ============ GetRoutes Handler Tests ============

func TestDebugHandlers_GetRoutes(t *testing.T) {
	router := chi.NewRouter()

	// Add some test routes
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {})
	router.Post("/test", func(w http.ResponseWriter, r *http.Request) {})
	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {})

	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", nil)
	w := httptest.NewRecorder()

	h.GetRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetRoutes() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Response missing 'data' field")
	}

	if _, ok := data["count"]; !ok {
		t.Error("GetRoutes() missing 'count' field")
	}

	if _, ok := data["routes"]; !ok {
		t.Error("GetRoutes() missing 'routes' field")
	}

	routes := data["routes"].([]interface{})
	if len(routes) < 3 {
		t.Errorf("GetRoutes() returned %d routes, want at least 3", len(routes))
	}
}

func TestDebugHandlers_GetRoutes_EmptyRouter(t *testing.T) {
	router := chi.NewRouter()
	h := NewDebugHandlers(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", nil)
	w := httptest.NewRecorder()

	h.GetRoutes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetRoutes() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	data := resp["data"].(map[string]interface{})
	count := data["count"].(float64)

	// Empty router should have 0 routes
	if count != 0 {
		t.Errorf("GetRoutes() count = %v, want 0 for empty router", count)
	}
}

// ============ RegisterRoutes Tests ============

func TestDebugHandlers_RegisterRoutes(t *testing.T) {
	mainRouter := chi.NewRouter()
	h := NewDebugHandlers(mainRouter)

	testRouter := chi.NewRouter()
	h.RegisterRoutes(testRouter)

	// Check that routes are registered by walking the router
	var routes []string
	chi.Walk(testRouter, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes = append(routes, method+" "+route)
		return nil
	})

	expectedRoutes := []string{
		"GET /debug/config",
		"GET /debug/routes",
		"GET /debug/stats",
		"POST /debug/gc",
	}

	for _, expected := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("RegisterRoutes() missing route %q", expected)
		}
	}
}

// ============ SanitizedConfig Structure Tests ============

func TestSanitizedConfig_Structure(t *testing.T) {
	config := SanitizedConfig{
		Server: map[string]interface{}{
			"port": "8080",
		},
		Database: map[string]interface{}{
			"host": "localhost",
		},
		RateLimit: map[string]interface{}{
			"enabled": "true",
		},
		Logging: map[string]interface{}{
			"level": "info",
		},
		Features: map[string]interface{}{
			"cache_enabled": "false",
		},
	}

	if config.Server["port"] != "8080" {
		t.Errorf("SanitizedConfig.Server.port = %v, want 8080", config.Server["port"])
	}

	if config.Database["host"] != "localhost" {
		t.Errorf("SanitizedConfig.Database.host = %v, want localhost", config.Database["host"])
	}
}

// ============ RuntimeStats Structure Tests ============

func TestRuntimeStats_Structure(t *testing.T) {
	stats := RuntimeStats{
		Uptime:        "1h30m0s",
		GoVersion:     "go1.25",
		NumCPU:        4,
		NumGoroutines: 10,
		AllocMB:       50.5,
		TotalAllocMB:  100.0,
		SysMB:         200.0,
		NumGC:         5,
		HeapObjectsK:  1.5,
		HeapAllocMB:   45.0,
		HeapInUseMB:   48.0,
	}

	if stats.NumCPU != 4 {
		t.Errorf("RuntimeStats.NumCPU = %d, want 4", stats.NumCPU)
	}

	if stats.NumGoroutines != 10 {
		t.Errorf("RuntimeStats.NumGoroutines = %d, want 10", stats.NumGoroutines)
	}
}

// ============ PoolStats Structure Tests ============

func TestPoolStats_Structure(t *testing.T) {
	stats := PoolStats{
		OpenConnections: 10,
		InUse:           5,
		Idle:            5,
		WaitCount:       0,
		WaitDuration:    "0s",
		MaxOpenConns:    100,
	}

	if stats.OpenConnections != 10 {
		t.Errorf("PoolStats.OpenConnections = %d, want 10", stats.OpenConnections)
	}

	if stats.InUse != 5 {
		t.Errorf("PoolStats.InUse = %d, want 5", stats.InUse)
	}
}

// ============ RouteInfo Structure Tests ============

func TestRouteInfo_Structure(t *testing.T) {
	info := RouteInfo{
		Method:  "GET",
		Pattern: "/api/users",
	}

	if info.Method != "GET" {
		t.Errorf("RouteInfo.Method = %q, want GET", info.Method)
	}

	if info.Pattern != "/api/users" {
		t.Errorf("RouteInfo.Pattern = %q, want /api/users", info.Pattern)
	}
}

func TestRouteInfo_JSONMarshaling(t *testing.T) {
	info := RouteInfo{
		Method:  "POST",
		Pattern: "/api/create",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal RouteInfo: %v", err)
	}

	var decoded RouteInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal RouteInfo: %v", err)
	}

	if decoded.Method != info.Method {
		t.Errorf("RouteInfo.Method after unmarshal = %q, want %q", decoded.Method, info.Method)
	}

	if decoded.Pattern != info.Pattern {
		t.Errorf("RouteInfo.Pattern after unmarshal = %q, want %q", decoded.Pattern, info.Pattern)
	}
}
