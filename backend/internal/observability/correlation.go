// Package observability provides correlation and tracing utilities for logging and error tracking.
package observability

import (
	"context"

	"react-golang-starter/internal/contextkeys"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// CorrelationContext holds correlation data for request tracing.
type CorrelationContext struct {
	// RequestID is the unique identifier for this request
	RequestID string `json:"request_id,omitempty"`

	// TraceID is used for distributed tracing (optional)
	TraceID string `json:"trace_id,omitempty"`

	// UserID is the authenticated user's ID (if any)
	UserID uint `json:"user_id,omitempty"`

	// SessionID is the user's session ID (if any)
	SessionID string `json:"session_id,omitempty"`

	// SentryEventID is the Sentry event ID if an error was captured
	SentryEventID string `json:"sentry_event_id,omitempty"`
}

// GetCorrelation extracts correlation context from the request context.
func GetCorrelation(ctx context.Context) *CorrelationContext {
	cc := &CorrelationContext{}

	// Extract request ID
	if requestID, ok := ctx.Value(contextkeys.RequestIDKey).(string); ok {
		cc.RequestID = requestID
		cc.TraceID = requestID // Use request ID as trace ID by default
	}

	// Extract user ID if available (using the contextkeys.UserIDKey type)
	if userID, ok := ctx.Value(contextkeys.UserIDKey).(uint); ok {
		cc.UserID = userID
	}

	return cc
}

// WithCorrelation returns a logger with correlation context attached.
func WithCorrelation(ctx context.Context, logger zerolog.Logger) zerolog.Logger {
	cc := GetCorrelation(ctx)

	l := logger.With()

	if cc.RequestID != "" {
		l = l.Str("request_id", cc.RequestID)
	}
	if cc.TraceID != "" {
		l = l.Str("trace_id", cc.TraceID)
	}
	if cc.UserID != 0 {
		l = l.Uint("user_id", cc.UserID)
	}
	if cc.SessionID != "" {
		l = l.Str("session_id", cc.SessionID)
	}

	return l.Logger()
}

// LogWithCorrelation returns a logger with correlation from the given context.
func LogWithCorrelation(ctx context.Context) zerolog.Logger {
	return WithCorrelation(ctx, log.Logger)
}

// CaptureError captures an error to Sentry with correlation context.
// Returns the Sentry event ID for reference.
func CaptureError(ctx context.Context, err error, extras map[string]interface{}) string {
	cc := GetCorrelation(ctx)

	// Get or create hub from context
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}

	// Set correlation tags
	hub.Scope().SetTag("request_id", cc.RequestID)
	hub.Scope().SetTag("trace_id", cc.TraceID)

	if cc.UserID != 0 {
		hub.Scope().SetUser(sentry.User{
			ID: string(rune(cc.UserID)),
		})
	}

	// Add extras
	for k, v := range extras {
		hub.Scope().SetExtra(k, v)
	}

	// Capture the error
	eventID := hub.CaptureException(err)
	if eventID != nil {
		return string(*eventID)
	}
	return ""
}

// CaptureMessage captures a message to Sentry with correlation context.
func CaptureMessage(ctx context.Context, message string, level sentry.Level, extras map[string]interface{}) string {
	cc := GetCorrelation(ctx)

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}

	hub.Scope().SetTag("request_id", cc.RequestID)
	hub.Scope().SetTag("trace_id", cc.TraceID)
	hub.Scope().SetLevel(level)

	for k, v := range extras {
		hub.Scope().SetExtra(k, v)
	}

	eventID := hub.CaptureMessage(message)
	if eventID != nil {
		return string(*eventID)
	}
	return ""
}

// LogAndCapture logs an error and captures it to Sentry, returning the event ID.
func LogAndCapture(ctx context.Context, err error, message string, extras map[string]interface{}) string {
	logger := LogWithCorrelation(ctx)

	// Log the error
	event := logger.Error().Err(err)
	for k, v := range extras {
		event = event.Interface(k, v)
	}
	event.Msg(message)

	// Capture to Sentry
	eventID := CaptureError(ctx, err, extras)

	// Log the Sentry event ID for correlation
	if eventID != "" {
		logger.Debug().Str("sentry_event_id", eventID).Msg("error captured to Sentry")
	}

	return eventID
}

// AddBreadcrumb adds a breadcrumb to the current Sentry scope.
func AddBreadcrumb(ctx context.Context, category string, message string, data map[string]interface{}) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: category,
		Message:  message,
		Data:     data,
		Level:    sentry.LevelInfo,
	}, nil)
}

// StartSpan starts a new Sentry span for performance monitoring.
// Returns the span and a cleanup function that should be deferred.
func StartSpan(ctx context.Context, operation string, description string) (*sentry.Span, func()) {
	span := sentry.StartSpan(ctx, operation)
	span.Description = description

	cc := GetCorrelation(ctx)
	if cc.RequestID != "" {
		span.SetTag("request_id", cc.RequestID)
	}

	return span, func() {
		span.Finish()
	}
}
