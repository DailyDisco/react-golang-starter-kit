package middleware

import (
	"context"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
)

// SentryConfig holds configuration for Sentry error tracking
type SentryConfig struct {
	Enabled     bool
	DSN         string
	Environment string
	Release     string
	SampleRate  float64
	Debug       bool
}

// DefaultSentryConfig returns a default Sentry configuration
func DefaultSentryConfig() *SentryConfig {
	return &SentryConfig{
		Enabled:     false,
		DSN:         "",
		Environment: "development",
		Release:     "1.0.0",
		SampleRate:  1.0,
		Debug:       false,
	}
}

// LoadSentryConfig loads Sentry configuration from environment variables
func LoadSentryConfig() *SentryConfig {
	config := DefaultSentryConfig()

	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		config.DSN = dsn
		config.Enabled = true
	}

	if env := os.Getenv("SENTRY_ENVIRONMENT"); env != "" {
		config.Environment = env
	} else if env := os.Getenv("APP_ENV"); env != "" {
		config.Environment = env
	}

	if release := os.Getenv("APP_VERSION"); release != "" {
		config.Release = release
	}

	if os.Getenv("SENTRY_DEBUG") == "true" {
		config.Debug = true
	}

	// In production, reduce sample rate if needed
	if config.Environment == "production" {
		config.SampleRate = 0.1
	}

	return config
}

// InitSentry initializes the Sentry SDK
// Returns an error if initialization fails
func InitSentry(config *SentryConfig) error {
	if !config.Enabled || config.DSN == "" {
		log.Debug().Msg("Sentry DSN not configured, skipping initialization")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		TracesSampleRate: config.SampleRate,
		Debug:            config.Debug,

		// Configure which errors to ignore
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Filter out certain error types if needed
			if hint.OriginalException != nil {
				// Example: skip context canceled errors
				if err, ok := hint.OriginalException.(error); ok {
					if err == context.Canceled {
						return nil
					}
				}
			}
			return event
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Sentry")
		return err
	}

	log.Info().
		Str("environment", config.Environment).
		Str("release", config.Release).
		Msg("Sentry initialized")

	return nil
}

// FlushSentry flushes any buffered events to Sentry
// Call this before the application exits
func FlushSentry(timeout time.Duration) {
	sentry.Flush(timeout)
}

// SentryMiddleware returns a middleware that captures panics and reports them to Sentry
func SentryMiddleware(config *SentryConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Create a new hub for this request
			hub := sentry.CurrentHub().Clone()
			ctx := sentry.SetHubOnContext(r.Context(), hub)

			// Configure the scope with request info
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetRequest(r)
				scope.SetTag("method", r.Method)
				scope.SetTag("path", r.URL.Path)

				// Add request ID if available
				if reqID := r.Context().Value(RequestIDKey); reqID != nil {
					if id, ok := reqID.(string); ok {
						scope.SetTag("request_id", id)
					}
				}
			})

			// Recover from panics and report to Sentry
			defer func() {
				if err := recover(); err != nil {
					hub.RecoverWithContext(ctx, err)

					// Also log the panic locally
					log.Error().
						Interface("panic", err).
						Str("stack", string(debug.Stack())).
						Msg("Recovered from panic")

					// Return 500 error to client
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			// Use the context with Sentry hub
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// CaptureError captures an error and sends it to Sentry
func CaptureError(err error, ctx context.Context, extras map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			for key, value := range extras {
				scope.SetExtra(key, value)
			}
			hub.CaptureException(err)
		})
	} else {
		// Fallback to global hub if no context hub
		sentry.WithScope(func(scope *sentry.Scope) {
			for key, value := range extras {
				scope.SetExtra(key, value)
			}
			sentry.CaptureException(err)
		})
	}
}

// CaptureMessage captures a message and sends it to Sentry
func CaptureMessage(message string, level sentry.Level, ctx context.Context) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureMessage(message)
	} else {
		sentry.CaptureMessage(message)
	}
}

// SetUser sets the user context for Sentry events
func SetUser(ctx context.Context, id string, email string, username string) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       id,
				Email:    email,
				Username: username,
			})
		})
	}
}

// AddBreadcrumb adds a breadcrumb to the current scope
func AddBreadcrumb(ctx context.Context, category string, message string, data map[string]interface{}) {
	breadcrumb := &sentry.Breadcrumb{
		Category: category,
		Message:  message,
		Data:     data,
		Level:    sentry.LevelInfo,
	}

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(breadcrumb, nil)
	} else {
		sentry.AddBreadcrumb(breadcrumb)
	}
}
