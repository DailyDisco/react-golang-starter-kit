package observability

import (
	"context"
	"errors"
	"testing"

	"react-golang-starter/internal/contextkeys"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// ============ CorrelationContext Tests ============

func TestCorrelationContext_Structure(t *testing.T) {
	cc := CorrelationContext{
		RequestID:     "req-123",
		TraceID:       "trace-456",
		UserID:        42,
		SessionID:     "sess-789",
		SentryEventID: "sentry-abc",
	}

	if cc.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want %q", cc.RequestID, "req-123")
	}
	if cc.TraceID != "trace-456" {
		t.Errorf("TraceID = %q, want %q", cc.TraceID, "trace-456")
	}
	if cc.UserID != 42 {
		t.Errorf("UserID = %d, want %d", cc.UserID, 42)
	}
	if cc.SessionID != "sess-789" {
		t.Errorf("SessionID = %q, want %q", cc.SessionID, "sess-789")
	}
	if cc.SentryEventID != "sentry-abc" {
		t.Errorf("SentryEventID = %q, want %q", cc.SentryEventID, "sentry-abc")
	}
}

func TestGetCorrelation_EmptyContext(t *testing.T) {
	ctx := context.Background()
	cc := GetCorrelation(ctx)

	if cc == nil {
		t.Fatal("GetCorrelation() returned nil")
	}

	if cc.RequestID != "" {
		t.Errorf("RequestID = %q, want empty", cc.RequestID)
	}
	if cc.TraceID != "" {
		t.Errorf("TraceID = %q, want empty", cc.TraceID)
	}
	if cc.UserID != 0 {
		t.Errorf("UserID = %d, want 0", cc.UserID)
	}
}

func TestGetCorrelation_WithRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "test-request-id")
	cc := GetCorrelation(ctx)

	if cc.RequestID != "test-request-id" {
		t.Errorf("RequestID = %q, want %q", cc.RequestID, "test-request-id")
	}
	// TraceID defaults to RequestID
	if cc.TraceID != "test-request-id" {
		t.Errorf("TraceID = %q, want %q", cc.TraceID, "test-request-id")
	}
}

func TestGetCorrelation_WithUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.UserIDKey, uint(123))
	cc := GetCorrelation(ctx)

	if cc.UserID != 123 {
		t.Errorf("UserID = %d, want 123", cc.UserID)
	}
}

func TestGetCorrelation_WithAllValues(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.RequestIDKey, "req-all")
	ctx = context.WithValue(ctx, contextkeys.UserIDKey, uint(999))

	cc := GetCorrelation(ctx)

	if cc.RequestID != "req-all" {
		t.Errorf("RequestID = %q, want %q", cc.RequestID, "req-all")
	}
	if cc.TraceID != "req-all" {
		t.Errorf("TraceID = %q, want %q", cc.TraceID, "req-all")
	}
	if cc.UserID != 999 {
		t.Errorf("UserID = %d, want 999", cc.UserID)
	}
}

func TestGetCorrelation_WrongType(t *testing.T) {
	// Use wrong type for request ID
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, 12345) // int instead of string
	cc := GetCorrelation(ctx)

	// Should return empty string since type assertion fails
	if cc.RequestID != "" {
		t.Errorf("RequestID = %q, want empty (wrong type)", cc.RequestID)
	}
}

// ============ WithCorrelation Tests ============

func TestWithCorrelation_EmptyContext(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	newLogger := WithCorrelation(ctx, logger)

	// Should return a valid logger (doesn't panic)
	if newLogger.GetLevel() != zerolog.Disabled {
		// Nop logger level may vary, just check it's valid
	}
}

func TestWithCorrelation_WithAllFields(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.RequestIDKey, "req-corr-test")
	ctx = context.WithValue(ctx, contextkeys.UserIDKey, uint(42))

	logger := zerolog.Nop()
	newLogger := WithCorrelation(ctx, logger)

	// Logger should be valid and not panic
	newLogger.Info().Msg("test message")
}

func TestLogWithCorrelation(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "log-test")

	logger := LogWithCorrelation(ctx)

	// Should return a valid logger
	logger.Info().Msg("test message")
}

// ============ Metrics Variable Tests ============

func TestHTTPMetricsExist(t *testing.T) {
	// Verify HTTP metrics are initialized
	if HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal is nil")
	}
	if HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration is nil")
	}
	if HTTPRequestsInFlight == nil {
		t.Error("HTTPRequestsInFlight is nil")
	}
}

func TestDBMetricsExist(t *testing.T) {
	if DBQueryDuration == nil {
		t.Error("DBQueryDuration is nil")
	}
	if DBConnectionsActive == nil {
		t.Error("DBConnectionsActive is nil")
	}
	if DBConnectionsIdle == nil {
		t.Error("DBConnectionsIdle is nil")
	}
}

func TestWebSocketMetricsExist(t *testing.T) {
	if WSConnectionsActive == nil {
		t.Error("WSConnectionsActive is nil")
	}
	if WSMessagesTotal == nil {
		t.Error("WSMessagesTotal is nil")
	}
}

func TestAuthMetricsExist(t *testing.T) {
	if AuthenticationsTotal == nil {
		t.Error("AuthenticationsTotal is nil")
	}
	if UsersRegisteredTotal == nil {
		t.Error("UsersRegisteredTotal is nil")
	}
}

func TestCacheMetricsExist(t *testing.T) {
	if CacheHitsTotal == nil {
		t.Error("CacheHitsTotal is nil")
	}
	if CacheMissesTotal == nil {
		t.Error("CacheMissesTotal is nil")
	}
}

func TestJobMetricsExist(t *testing.T) {
	if JobsProcessedTotal == nil {
		t.Error("JobsProcessedTotal is nil")
	}
	if JobsQueuedGauge == nil {
		t.Error("JobsQueuedGauge is nil")
	}
	if JobProcessingDuration == nil {
		t.Error("JobProcessingDuration is nil")
	}
}

func TestFileMetricsExist(t *testing.T) {
	if FilesUploadedTotal == nil {
		t.Error("FilesUploadedTotal is nil")
	}
	if FilesUploadedBytes == nil {
		t.Error("FilesUploadedBytes is nil")
	}
}

func TestBusinessMetricsExist(t *testing.T) {
	if ActiveSubscriptionsGauge == nil {
		t.Error("ActiveSubscriptionsGauge is nil")
	}
}

// ============ Metric Recording Functions Tests ============

func TestRecordHTTPRequest(t *testing.T) {
	// Should not panic
	RecordHTTPRequest("GET", "/api/users", "200", 0.150)
	RecordHTTPRequest("POST", "/api/users", "201", 0.250)
	RecordHTTPRequest("DELETE", "/api/users/1", "404", 0.050)
}

func TestRecordDBQuery(t *testing.T) {
	// Should not panic
	RecordDBQuery("select", 0.005)
	RecordDBQuery("insert", 0.010)
	RecordDBQuery("update", 0.008)
	RecordDBQuery("delete", 0.003)
}

func TestRecordAuthentication(t *testing.T) {
	// Should not panic
	RecordAuthentication("success", "password")
	RecordAuthentication("failure", "password")
	RecordAuthentication("success", "oauth")
	RecordAuthentication("success", "refresh")
}

func TestRecordWSConnection(t *testing.T) {
	// Should not panic
	RecordWSConnection(1)  // Connect
	RecordWSConnection(-1) // Disconnect
	RecordWSConnection(0)  // No change
}

func TestRecordWSMessage(t *testing.T) {
	// Should not panic
	RecordWSMessage("chat", "inbound")
	RecordWSMessage("chat", "outbound")
	RecordWSMessage("notification", "outbound")
}

func TestRecordCacheHit(t *testing.T) {
	// Should not panic
	RecordCacheHit("user")
	RecordCacheHit("session")
	RecordCacheHit("org")
}

func TestRecordCacheMiss(t *testing.T) {
	// Should not panic
	RecordCacheMiss("user")
	RecordCacheMiss("session")
	RecordCacheMiss("org")
}

func TestRecordJobProcessed(t *testing.T) {
	// Should not panic
	RecordJobProcessed("send_email", "success", 1.5)
	RecordJobProcessed("send_email", "failure", 0.5)
	RecordJobProcessed("data_export", "success", 30.0)
}

func TestRecordFileUpload(t *testing.T) {
	// Should not panic
	RecordFileUpload(1024)     // 1 KB
	RecordFileUpload(1048576)  // 1 MB
	RecordFileUpload(10485760) // 10 MB
}

// ============ Metric Labels Tests ============

func TestHTTPRequestsTotal_Labels(t *testing.T) {
	// Verify the metric has the expected labels
	desc := HTTPRequestsTotal.WithLabelValues("GET", "/test", "200").Desc()
	if desc == nil {
		t.Error("HTTPRequestsTotal descriptor is nil")
	}
}

func TestHTTPRequestDuration_Buckets(t *testing.T) {
	// The histogram should use default buckets
	// Test by observing a value - should not panic
	HTTPRequestDuration.WithLabelValues("GET", "/test").Observe(0.1)
}

func TestDBQueryDuration_CustomBuckets(t *testing.T) {
	// DB query duration uses custom buckets for faster queries
	// Test by observing a value - should not panic
	DBQueryDuration.WithLabelValues("select").Observe(0.005)
}

func TestJobProcessingDuration_CustomBuckets(t *testing.T) {
	// Job processing uses custom buckets for longer operations
	// Test by observing a value - should not panic
	JobProcessingDuration.WithLabelValues("email").Observe(5.0)
}

// ============ Sentry Function Tests ============

func TestCaptureError_NilHub(t *testing.T) {
	ctx := context.Background()
	err := errors.New("test error")

	// Should not panic even without Sentry initialized
	eventID := CaptureError(ctx, err, nil)
	// Event ID will be empty without Sentry
	_ = eventID
}

func TestCaptureError_WithExtras(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.RequestIDKey, "req-capture")
	ctx = context.WithValue(ctx, contextkeys.UserIDKey, uint(42))

	err := errors.New("test error with extras")
	extras := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	// Should not panic
	_ = CaptureError(ctx, err, extras)
}

func TestCaptureMessage_NilHub(t *testing.T) {
	ctx := context.Background()

	// Should not panic even without Sentry initialized
	eventID := CaptureMessage(ctx, "test message", "info", nil)
	_ = eventID
}

func TestCaptureMessage_WithExtras(t *testing.T) {
	ctx := context.Background()
	extras := map[string]interface{}{
		"feature": "test",
	}

	// Should not panic
	_ = CaptureMessage(ctx, "test message", "warning", extras)
}

func TestLogAndCapture(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "req-log-capture")
	err := errors.New("logged and captured error")
	extras := map[string]interface{}{
		"context": "test",
	}

	// Should not panic
	_ = LogAndCapture(ctx, err, "test message", extras)
}

func TestLogAndCapture_NilExtras(t *testing.T) {
	ctx := context.Background()
	err := errors.New("test error")

	// Should not panic with nil extras
	_ = LogAndCapture(ctx, err, "test message", nil)
}

func TestAddBreadcrumb(t *testing.T) {
	ctx := context.Background()
	data := map[string]interface{}{
		"action": "click",
	}

	// Should not panic
	AddBreadcrumb(ctx, "ui", "User clicked button", data)
}

func TestAddBreadcrumb_NilData(t *testing.T) {
	ctx := context.Background()

	// Should not panic with nil data
	AddBreadcrumb(ctx, "navigation", "User navigated to page", nil)
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()

	span, cleanup := StartSpan(ctx, "test.operation", "Test operation description")
	defer cleanup()

	if span == nil {
		t.Error("StartSpan returned nil span")
	}
}

func TestStartSpan_WithCorrelation(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "req-span")

	span, cleanup := StartSpan(ctx, "db.query", "SELECT * FROM users")
	defer cleanup()

	if span == nil {
		t.Error("StartSpan returned nil span")
	}
}

// ============ Gauge Tests ============

func TestHTTPRequestsInFlight_Gauge(t *testing.T) {
	// Test gauge operations
	HTTPRequestsInFlight.Inc()
	HTTPRequestsInFlight.Dec()
	HTTPRequestsInFlight.Set(5)
	HTTPRequestsInFlight.Add(3)
	HTTPRequestsInFlight.Sub(2)
}

func TestDBConnectionsActive_Gauge(t *testing.T) {
	DBConnectionsActive.Set(10)
	DBConnectionsActive.Inc()
	DBConnectionsActive.Dec()
}

func TestDBConnectionsIdle_Gauge(t *testing.T) {
	DBConnectionsIdle.Set(5)
	DBConnectionsIdle.Inc()
	DBConnectionsIdle.Dec()
}

func TestWSConnectionsActive_Gauge(t *testing.T) {
	WSConnectionsActive.Set(100)
	WSConnectionsActive.Add(10)
	WSConnectionsActive.Sub(5)
}

func TestJobsQueuedGauge_WithLabels(t *testing.T) {
	JobsQueuedGauge.WithLabelValues("email").Set(10)
	JobsQueuedGauge.WithLabelValues("export").Set(5)
	JobsQueuedGauge.WithLabelValues("cleanup").Set(0)
}

func TestActiveSubscriptionsGauge_WithLabels(t *testing.T) {
	ActiveSubscriptionsGauge.WithLabelValues("free").Set(1000)
	ActiveSubscriptionsGauge.WithLabelValues("pro").Set(200)
	ActiveSubscriptionsGauge.WithLabelValues("enterprise").Set(50)
}

// ============ Counter Tests ============

func TestUsersRegisteredTotal_Counter(t *testing.T) {
	// Counter can only increase
	UsersRegisteredTotal.Inc()
	UsersRegisteredTotal.Add(5)
}

func TestFilesUploadedTotal_Counter(t *testing.T) {
	FilesUploadedTotal.Inc()
	FilesUploadedTotal.Add(10)
}

func TestFilesUploadedBytes_Counter(t *testing.T) {
	FilesUploadedBytes.Add(1024)
	FilesUploadedBytes.Add(2048)
}

// ============ Integration Tests ============

func TestMetricsRegistry(t *testing.T) {
	// Verify all metrics are registered
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Should have some metrics
	if len(mfs) == 0 {
		t.Error("No metrics gathered")
	}

	// Look for our custom metrics
	metricNames := make(map[string]bool)
	for _, mf := range mfs {
		metricNames[mf.GetName()] = true
	}

	expectedMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"http_requests_in_flight",
		"db_query_duration_seconds",
		"db_connections_active",
		"db_connections_idle",
		"websocket_connections_active",
		"websocket_messages_total",
		"authentications_total",
		"users_registered_total",
		"cache_hits_total",
		"cache_misses_total",
		"jobs_processed_total",
		"jobs_queued",
		"job_processing_duration_seconds",
		"files_uploaded_total",
		"files_uploaded_bytes_total",
		"active_subscriptions",
	}

	for _, name := range expectedMetrics {
		if !metricNames[name] {
			t.Errorf("Expected metric %q not found in registry", name)
		}
	}
}
