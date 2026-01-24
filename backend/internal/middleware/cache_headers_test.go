package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============ DefaultCacheHeadersConfig Tests ============

func TestDefaultCacheHeadersConfig(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	if config == nil {
		t.Fatal("DefaultCacheHeadersConfig() returned nil")
	}

	if !config.Enabled {
		t.Error("DefaultCacheHeadersConfig().Enabled = false, want true")
	}

	if len(config.Rules) == 0 {
		t.Error("DefaultCacheHeadersConfig().Rules is empty")
	}
}

func TestDefaultCacheHeadersConfig_Rules(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	expectedPrefixes := []string{
		"/api/health",
		"/api/v1/changelog",
		"/api/admin/",
		"/api/v1/auth/",
		"/api/v1/billing/",
		"/api/v1/files/",
		"/api/v1/settings/",
		"/api/v1/organizations/",
		"/api/",
	}

	if len(config.Rules) != len(expectedPrefixes) {
		t.Errorf("Number of rules = %d, want %d", len(config.Rules), len(expectedPrefixes))
	}

	for i, prefix := range expectedPrefixes {
		if i >= len(config.Rules) {
			break
		}
		if config.Rules[i].PathPrefix != prefix {
			t.Errorf("Rules[%d].PathPrefix = %q, want %q", i, config.Rules[i].PathPrefix, prefix)
		}
	}
}

func TestDefaultCacheHeadersConfig_AdminNoCache(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	// Find admin rule
	var adminRule *CacheRule
	for _, rule := range config.Rules {
		if rule.PathPrefix == "/api/admin/" {
			adminRule = &rule
			break
		}
	}

	if adminRule == nil {
		t.Fatal("Admin rule not found")
	}

	if !adminRule.NoCache {
		t.Error("Admin rule NoCache = false, want true")
	}

	if !adminRule.NoStore {
		t.Error("Admin rule NoStore = false, want true")
	}

	if !adminRule.MustRevalidate {
		t.Error("Admin rule MustRevalidate = false, want true")
	}
}

func TestDefaultCacheHeadersConfig_AuthNoCache(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	var authRule *CacheRule
	for _, rule := range config.Rules {
		if rule.PathPrefix == "/api/v1/auth/" {
			authRule = &rule
			break
		}
	}

	if authRule == nil {
		t.Fatal("Auth rule not found")
	}

	if !authRule.NoCache {
		t.Error("Auth rule NoCache = false, want true")
	}

	if !authRule.NoStore {
		t.Error("Auth rule NoStore = false, want true")
	}
}

// ============ LoadCacheHeadersConfig Tests ============

func TestLoadCacheHeadersConfig(t *testing.T) {
	config := LoadCacheHeadersConfig()

	if config == nil {
		t.Fatal("LoadCacheHeadersConfig() returned nil")
	}
}

func TestLoadCacheHeadersConfig_FromEnv(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   bool
	}{
		{"enabled true", "true", true},
		{"enabled TRUE", "TRUE", true},
		{"enabled false", "false", false},
		{"enabled 0", "0", false},
		{"empty defaults to true", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CACHE_HEADERS_ENABLED", tt.envVal)
			config := LoadCacheHeadersConfig()
			if config.Enabled != tt.want {
				t.Errorf("LoadCacheHeadersConfig().Enabled = %v, want %v", config.Enabled, tt.want)
			}
		})
	}
}

// ============ CacheHeaders Middleware Tests ============

func TestCacheHeaders_NilConfig(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheHeaders(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCacheHeaders_Disabled(t *testing.T) {
	config := &CacheHeadersConfig{Enabled: false}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheHeaders(config)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "" {
		t.Errorf("Cache-Control = %q, want empty (disabled)", cacheControl)
	}
}

func TestCacheHeaders_OnlyGETAndHEAD(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			middleware := CacheHeaders(config)
			req := httptest.NewRequest(method, "/api/health", nil)
			w := httptest.NewRecorder()

			middleware(handler).ServeHTTP(w, req)

			cacheControl := w.Header().Get("Cache-Control")
			if cacheControl != "" {
				t.Errorf("Cache-Control for %s = %q, want empty", method, cacheControl)
			}
		})
	}
}

func TestCacheHeaders_GETRequest(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheHeaders(config)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl == "" {
		t.Error("Cache-Control is empty for GET request")
	}
}

func TestCacheHeaders_HEADRequest(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheHeaders(config)
	req := httptest.NewRequest(http.MethodHead, "/api/health", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl == "" {
		t.Error("Cache-Control is empty for HEAD request")
	}
}

func TestCacheHeaders_HealthEndpoint(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheHeaders(config)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=5" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "public, max-age=5")
	}
}

func TestCacheHeaders_AdminEndpoint(t *testing.T) {
	config := DefaultCacheHeadersConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CacheHeaders(config)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	// Should contain no-cache, no-store, must-revalidate
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "no-cache, no-store, must-revalidate")
	}

	pragma := w.Header().Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Pragma = %q, want %q", pragma, "no-cache")
	}

	expires := w.Header().Get("Expires")
	if expires != "0" {
		t.Errorf("Expires = %q, want %q", expires, "0")
	}
}

// ============ applyCacheRule Tests ============

func TestApplyCacheRule_PublicCache(t *testing.T) {
	w := httptest.NewRecorder()
	rule := CacheRule{MaxAge: 300, Public: true}

	applyCacheRule(w, rule)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=300" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "public, max-age=300")
	}
}

func TestApplyCacheRule_PrivateCache(t *testing.T) {
	w := httptest.NewRecorder()
	rule := CacheRule{MaxAge: 60, Public: false}

	applyCacheRule(w, rule)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "private, max-age=60" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "private, max-age=60")
	}
}

func TestApplyCacheRule_NoCache(t *testing.T) {
	w := httptest.NewRecorder()
	rule := CacheRule{NoCache: true}

	applyCacheRule(w, rule)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "no-cache")
	}

	pragma := w.Header().Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Pragma = %q, want %q", pragma, "no-cache")
	}

	expires := w.Header().Get("Expires")
	if expires != "0" {
		t.Errorf("Expires = %q, want %q", expires, "0")
	}
}

func TestApplyCacheRule_NoStore(t *testing.T) {
	w := httptest.NewRecorder()
	rule := CacheRule{NoStore: true}

	applyCacheRule(w, rule)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-store" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "no-store")
	}
}

func TestApplyCacheRule_MustRevalidate(t *testing.T) {
	w := httptest.NewRecorder()
	rule := CacheRule{MaxAge: 60, Public: true, MustRevalidate: true}

	applyCacheRule(w, rule)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=60, must-revalidate" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "public, max-age=60, must-revalidate")
	}
}

func TestApplyCacheRule_NoCacheNoStoreMustRevalidate(t *testing.T) {
	w := httptest.NewRecorder()
	rule := CacheRule{NoCache: true, NoStore: true, MustRevalidate: true}

	applyCacheRule(w, rule)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "no-cache, no-store, must-revalidate")
	}
}

// ============ WithCacheRule Tests ============

func TestWithCacheRule_GET(t *testing.T) {
	rule := CacheRule{MaxAge: 3600, Public: true}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := WithCacheRule(rule)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "public, max-age=3600")
	}
}

func TestWithCacheRule_HEAD(t *testing.T) {
	rule := CacheRule{MaxAge: 3600, Public: true}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := WithCacheRule(rule)
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "public, max-age=3600")
	}
}

func TestWithCacheRule_POST(t *testing.T) {
	rule := CacheRule{MaxAge: 3600, Public: true}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := WithCacheRule(rule)
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "" {
		t.Errorf("Cache-Control for POST = %q, want empty", cacheControl)
	}
}

// ============ NoCacheHeaders Tests ============

func TestNoCacheHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NoCacheHeaders()
	req := httptest.NewRequest(http.MethodGet, "/sensitive", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache, no-store, must-revalidate" {
		t.Errorf("Cache-Control = %q, want %q", cacheControl, "no-cache, no-store, must-revalidate")
	}

	pragma := w.Header().Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Pragma = %q, want %q", pragma, "no-cache")
	}
}

// ============ Common Cache Rules Tests ============

func TestCommonCacheRules(t *testing.T) {
	tests := []struct {
		name   string
		rule   CacheRule
		maxAge int
		public bool
	}{
		{"CacheRulePublic5Min", CacheRulePublic5Min, 300, true},
		{"CacheRulePublic1Hour", CacheRulePublic1Hour, 3600, true},
		{"CacheRulePrivate1Min", CacheRulePrivate1Min, 60, false},
		{"CacheRulePrivate5Min", CacheRulePrivate5Min, 300, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rule.MaxAge != tt.maxAge {
				t.Errorf("%s.MaxAge = %d, want %d", tt.name, tt.rule.MaxAge, tt.maxAge)
			}
			if tt.rule.Public != tt.public {
				t.Errorf("%s.Public = %v, want %v", tt.name, tt.rule.Public, tt.public)
			}
		})
	}
}

func TestCacheRuleNoCache(t *testing.T) {
	if !CacheRuleNoCache.NoCache {
		t.Error("CacheRuleNoCache.NoCache = false, want true")
	}
	if !CacheRuleNoCache.NoStore {
		t.Error("CacheRuleNoCache.NoStore = false, want true")
	}
	if !CacheRuleNoCache.MustRevalidate {
		t.Error("CacheRuleNoCache.MustRevalidate = false, want true")
	}
}

// ============ CacheRule Structure Tests ============

func TestCacheRule_Structure(t *testing.T) {
	rule := CacheRule{
		PathPrefix:     "/api/test",
		MaxAge:         600,
		Public:         true,
		NoCache:        false,
		NoStore:        false,
		MustRevalidate: true,
	}

	if rule.PathPrefix != "/api/test" {
		t.Errorf("PathPrefix = %q, want %q", rule.PathPrefix, "/api/test")
	}
	if rule.MaxAge != 600 {
		t.Errorf("MaxAge = %d, want %d", rule.MaxAge, 600)
	}
	if !rule.Public {
		t.Error("Public = false, want true")
	}
	if !rule.MustRevalidate {
		t.Error("MustRevalidate = false, want true")
	}
}

// ============ CacheHeadersConfig Structure Tests ============

func TestCacheHeadersConfig_Structure(t *testing.T) {
	config := CacheHeadersConfig{
		Enabled: true,
		Rules: []CacheRule{
			{PathPrefix: "/test", MaxAge: 60},
		},
	}

	if !config.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(config.Rules) != 1 {
		t.Errorf("Rules length = %d, want 1", len(config.Rules))
	}
}
