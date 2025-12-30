// Package observability provides correlation, tracing, and metrics utilities.
package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Database metrics
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"operation"},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// WebSocket metrics
	WSConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	WSMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_total",
			Help: "Total WebSocket messages",
		},
		[]string{"type", "direction"}, // direction: inbound, outbound
	)

	// Authentication metrics
	AuthenticationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "authentications_total",
			Help: "Total authentication attempts",
		},
		[]string{"result", "method"}, // result: success, failure; method: password, oauth, refresh
	)

	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of users registered",
		},
	)

	// Cache metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total cache hits",
		},
		[]string{"cache"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total cache misses",
		},
		[]string{"cache"},
	)

	// Job queue metrics
	JobsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_processed_total",
			Help: "Total jobs processed",
		},
		[]string{"job_type", "status"}, // status: success, failure
	)

	JobsQueuedGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jobs_queued",
			Help: "Number of jobs currently queued",
		},
		[]string{"job_type"},
	)

	JobProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_processing_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: []float64{.1, .5, 1, 2.5, 5, 10, 30, 60, 120},
		},
		[]string{"job_type"},
	)

	// File storage metrics
	FilesUploadedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "files_uploaded_total",
			Help: "Total files uploaded",
		},
	)

	FilesUploadedBytes = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "files_uploaded_bytes_total",
			Help: "Total bytes uploaded",
		},
	)

	// Business metrics
	ActiveSubscriptionsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_subscriptions",
			Help: "Number of active subscriptions by plan",
		},
		[]string{"plan"},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordDBQuery records a database query metric
func RecordDBQuery(operation string, duration float64) {
	DBQueryDuration.WithLabelValues(operation).Observe(duration)
}

// RecordAuthentication records an authentication attempt
func RecordAuthentication(result, method string) {
	AuthenticationsTotal.WithLabelValues(result, method).Inc()
}

// RecordWSConnection updates WebSocket connection count
func RecordWSConnection(delta int) {
	WSConnectionsActive.Add(float64(delta))
}

// RecordWSMessage records a WebSocket message
func RecordWSMessage(msgType, direction string) {
	WSMessagesTotal.WithLabelValues(msgType, direction).Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cache string) {
	CacheHitsTotal.WithLabelValues(cache).Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cache string) {
	CacheMissesTotal.WithLabelValues(cache).Inc()
}

// RecordJobProcessed records a processed job
func RecordJobProcessed(jobType, status string, duration float64) {
	JobsProcessedTotal.WithLabelValues(jobType, status).Inc()
	JobProcessingDuration.WithLabelValues(jobType).Observe(duration)
}

// RecordFileUpload records a file upload
func RecordFileUpload(bytes int64) {
	FilesUploadedTotal.Inc()
	FilesUploadedBytes.Add(float64(bytes))
}
