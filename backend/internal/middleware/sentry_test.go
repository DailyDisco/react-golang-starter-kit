package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/getsentry/sentry-go"
)

// ============ DefaultSentryConfig Tests ============

func TestDefaultSentryConfig(t *testing.T) {
	config := DefaultSentryConfig()

	if config.Enabled {
		t.Error("DefaultSentryConfig().Enabled = true, want false")
	}

	if config.DSN != "" {
		t.Errorf("DefaultSentryConfig().DSN = %q, want empty", config.DSN)
	}

	if config.Environment != "development" {
		t.Errorf("DefaultSentryConfig().Environment = %q, want %q", config.Environment, "development")
	}

	if config.Release != "1.0.0" {
		t.Errorf("DefaultSentryConfig().Release = %q, want %q", config.Release, "1.0.0")
	}

	if config.SampleRate != 1.0 {
		t.Errorf("DefaultSentryConfig().SampleRate = %f, want %f", config.SampleRate, 1.0)
	}

	if config.Debug {
		t.Error("DefaultSentryConfig().Debug = true, want false")
	}
}

// ============ LoadSentryConfig Tests ============

func TestLoadSentryConfig_Defaults(t *testing.T) {
	// Clear all relevant env vars
	envVars := []string{
		"SENTRY_DSN",
		"SENTRY_ENVIRONMENT",
		"APP_ENV",
		"APP_VERSION",
		"SENTRY_DEBUG",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	config := LoadSentryConfig()

	if config.Enabled {
		t.Error("LoadSentryConfig() should not be enabled without DSN")
	}

	if config.Environment != "development" {
		t.Errorf("LoadSentryConfig().Environment = %q, want %q", config.Environment, "development")
	}
}

func TestLoadSentryConfig_WithDSN(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://test@sentry.io/123")

	config := LoadSentryConfig()

	if !config.Enabled {
		t.Error("LoadSentryConfig() should be enabled with DSN")
	}

	if config.DSN != "https://test@sentry.io/123" {
		t.Errorf("LoadSentryConfig().DSN = %q, want %q", config.DSN, "https://test@sentry.io/123")
	}
}

func TestLoadSentryConfig_WithSentryEnvironment(t *testing.T) {
	t.Setenv("SENTRY_ENVIRONMENT", "staging")

	config := LoadSentryConfig()

	if config.Environment != "staging" {
		t.Errorf("LoadSentryConfig().Environment = %q, want %q", config.Environment, "staging")
	}
}

func TestLoadSentryConfig_WithAppEnvFallback(t *testing.T) {
	os.Unsetenv("SENTRY_ENVIRONMENT")
	t.Setenv("APP_ENV", "production")

	config := LoadSentryConfig()

	if config.Environment != "production" {
		t.Errorf("LoadSentryConfig().Environment = %q, want %q", config.Environment, "production")
	}
}

func TestLoadSentryConfig_SentryEnvironmentTakesPrecedence(t *testing.T) {
	t.Setenv("SENTRY_ENVIRONMENT", "staging")
	t.Setenv("APP_ENV", "production")

	config := LoadSentryConfig()

	if config.Environment != "staging" {
		t.Errorf("SENTRY_ENVIRONMENT should take precedence, got %q", config.Environment)
	}
}

func TestLoadSentryConfig_WithAppVersion(t *testing.T) {
	t.Setenv("APP_VERSION", "2.0.0")

	config := LoadSentryConfig()

	if config.Release != "2.0.0" {
		t.Errorf("LoadSentryConfig().Release = %q, want %q", config.Release, "2.0.0")
	}
}

func TestLoadSentryConfig_WithDebug(t *testing.T) {
	t.Setenv("SENTRY_DEBUG", "true")

	config := LoadSentryConfig()

	if !config.Debug {
		t.Error("LoadSentryConfig().Debug = false, want true")
	}
}

func TestLoadSentryConfig_DebugFalse(t *testing.T) {
	t.Setenv("SENTRY_DEBUG", "false")

	config := LoadSentryConfig()

	if config.Debug {
		t.Error("LoadSentryConfig().Debug = true, want false")
	}
}

func TestLoadSentryConfig_ProductionSampleRate(t *testing.T) {
	t.Setenv("SENTRY_ENVIRONMENT", "production")

	config := LoadSentryConfig()

	if config.SampleRate != 0.1 {
		t.Errorf("LoadSentryConfig().SampleRate = %f, want %f for production", config.SampleRate, 0.1)
	}
}

func TestLoadSentryConfig_NonProductionSampleRate(t *testing.T) {
	t.Setenv("SENTRY_ENVIRONMENT", "staging")

	config := LoadSentryConfig()

	if config.SampleRate != 1.0 {
		t.Errorf("LoadSentryConfig().SampleRate = %f, want %f for non-production", config.SampleRate, 1.0)
	}
}

// ============ InitSentry Tests ============

func TestInitSentry_Disabled(t *testing.T) {
	config := &SentryConfig{
		Enabled: false,
		DSN:     "",
	}

	err := InitSentry(config)

	if err != nil {
		t.Errorf("InitSentry() returned error for disabled config: %v", err)
	}
}

func TestInitSentry_NoDSN(t *testing.T) {
	config := &SentryConfig{
		Enabled: true,
		DSN:     "",
	}

	err := InitSentry(config)

	if err != nil {
		t.Errorf("InitSentry() returned error for empty DSN: %v", err)
	}
}

// ============ SentryMiddleware Tests ============

func TestSentryMiddleware_Disabled(t *testing.T) {
	config := &SentryConfig{
		Enabled: false,
	}

	handlerCalled := false
	handler := SentryMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("SentryMiddleware did not call next handler when disabled")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("SentryMiddleware() status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestSentryMiddleware_PassesToNextHandler(t *testing.T) {
	config := &SentryConfig{
		Enabled: true,
	}

	handlerCalled := false
	handler := SentryMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("SentryMiddleware did not call next handler")
	}

	if rr.Code != http.StatusTeapot {
		t.Errorf("SentryMiddleware() status = %d, want %d", rr.Code, http.StatusTeapot)
	}
}

func TestSentryMiddleware_RecoversPanic(t *testing.T) {
	config := &SentryConfig{
		Enabled: true,
	}

	handler := SentryMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	// Should not panic - middleware recovers
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("SentryMiddleware() after panic status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestSentryMiddleware_WithRequestID(t *testing.T) {
	config := &SentryConfig{
		Enabled: true,
	}

	handlerCalled := false
	handler := SentryMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), RequestIDKey, "test-request-id")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("SentryMiddleware did not call next handler with request ID")
	}
}

// ============ CaptureError Tests ============

func TestCaptureError_WithoutContext(t *testing.T) {
	// This just verifies it doesn't panic when called without proper context
	err := context.DeadlineExceeded
	ctx := context.Background()

	// Should not panic
	CaptureError(err, ctx, map[string]interface{}{
		"key": "value",
	})
}

func TestCaptureError_WithExtras(t *testing.T) {
	err := context.Canceled
	ctx := context.Background()

	extras := map[string]interface{}{
		"user_id":    123,
		"operation":  "test",
		"additional": true,
	}

	// Should not panic
	CaptureError(err, ctx, extras)
}

// ============ CaptureMessage Tests ============

func TestCaptureMessage_WithoutContext(t *testing.T) {
	ctx := context.Background()

	// Should not panic with any level
	CaptureMessage("test message", sentry.LevelInfo, ctx)
}

// ============ SetUser Tests ============

func TestSetUser_WithoutSentryContext(t *testing.T) {
	ctx := context.Background()

	// Should not panic when called without Sentry context
	SetUser(ctx, "123", "test@example.com", "testuser")
}

// ============ AddBreadcrumb Tests ============

func TestAddBreadcrumb_WithoutSentryContext(t *testing.T) {
	ctx := context.Background()

	// Should not panic when called without Sentry context
	AddBreadcrumb(ctx, "test", "test message", map[string]interface{}{
		"key": "value",
	})
}

func TestAddBreadcrumb_WithNilData(t *testing.T) {
	ctx := context.Background()

	// Should not panic with nil data
	AddBreadcrumb(ctx, "test", "test message", nil)
}
