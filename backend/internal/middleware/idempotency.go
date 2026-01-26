// Package middleware provides HTTP middleware for the application.
package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/cache"

	"github.com/rs/zerolog/log"
)

// IdempotencyConfig holds configuration for the idempotency middleware.
type IdempotencyConfig struct {
	// Enabled determines if idempotency checking is enabled
	Enabled bool
	// TTL is how long to cache idempotency responses (default: 24 hours)
	TTL time.Duration
	// HeaderName is the name of the idempotency key header
	HeaderName string
}

// IdempotencyResponse stores a cached response for an idempotency key.
type IdempotencyResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	CreatedAt  time.Time         `json:"created_at"`
}

// LoadIdempotencyConfig loads configuration from environment variables.
func LoadIdempotencyConfig() *IdempotencyConfig {
	ttl := 24 * time.Hour
	if envTTL := os.Getenv("IDEMPOTENCY_TTL"); envTTL != "" {
		if d, err := time.ParseDuration(envTTL); err == nil {
			ttl = d
		}
	}

	return &IdempotencyConfig{
		Enabled:    os.Getenv("IDEMPOTENCY_ENABLED") != "false",
		TTL:        ttl,
		HeaderName: getEnvOrDefault("IDEMPOTENCY_HEADER", "Idempotency-Key"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// IdempotencyMiddleware prevents duplicate processing of POST/PUT/PATCH requests.
// If a request with the same idempotency key is received within the TTL,
// the cached response is returned instead of processing the request again.
func IdempotencyMiddleware(config *IdempotencyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to state-changing methods
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			// Skip if idempotency is disabled or cache is unavailable
			if !config.Enabled || !cache.IsAvailable() {
				next.ServeHTTP(w, r)
				return
			}

			// Get idempotency key from header
			idempotencyKey := r.Header.Get(config.HeaderName)
			if idempotencyKey == "" {
				// No idempotency key - process normally
				next.ServeHTTP(w, r)
				return
			}

			// Get user ID for scoping (if authenticated)
			var userID uint
			if user, ok := auth.GetUserFromContext(r.Context()); ok {
				userID = user.ID
			}

			// Build cache key
			cacheKey := buildIdempotencyCacheKey(idempotencyKey, userID, r.URL.Path)

			// Check for existing response
			ctx := r.Context()
			existingResponse, err := getIdempotencyResponse(ctx, cacheKey)
			if err == nil && existingResponse != nil {
				// Verify request hash matches (same request body)
				requestHash := hashRequest(r)
				storedHashBytes, _ := cache.Get(ctx, cacheKey+":hash")
				storedHash := string(storedHashBytes)
				if storedHash == requestHash {
					log.Debug().
						Str("idempotency_key", idempotencyKey).
						Str("path", r.URL.Path).
						Msg("returning cached idempotent response")

					// Return cached response
					w.Header().Set("X-Idempotent-Response", "true")
					for key, value := range existingResponse.Headers {
						w.Header().Set(key, value)
					}
					w.WriteHeader(existingResponse.StatusCode)
					w.Write(existingResponse.Body)
					return
				}

				// Different request with same key - conflict
				log.Warn().
					Str("idempotency_key", idempotencyKey).
					Str("path", r.URL.Path).
					Msg("idempotency key reused with different request")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "Idempotency key conflict",
					"message": "This idempotency key was already used for a different request",
				})
				return
			}

			// Capture the response
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			// Process the request
			next.ServeHTTP(recorder, r)

			// Store the response if successful (2xx or 4xx client errors)
			if recorder.statusCode >= 200 && recorder.statusCode < 500 {
				response := &IdempotencyResponse{
					StatusCode: recorder.statusCode,
					Headers:    extractCacheableHeaders(recorder.Header()),
					Body:       recorder.body.Bytes(),
					CreatedAt:  time.Now(),
				}

				// Store response and request hash
				if err := storeIdempotencyResponse(ctx, cacheKey, response, config.TTL); err != nil {
					log.Warn().Err(err).Str("key", idempotencyKey).Msg("failed to cache idempotency response")
				}

				requestHash := hashRequest(r)
				cache.Set(ctx, cacheKey+":hash", []byte(requestHash), config.TTL)
			}
		})
	}
}

// buildIdempotencyCacheKey creates a cache key for an idempotency entry.
func buildIdempotencyCacheKey(key string, userID uint, path string) string {
	// Include user ID to scope keys per user
	return "idempotency:" + key + ":" + strconv.FormatUint(uint64(userID), 10) + ":" + path
}

// hashRequest creates a hash of the request body for verification.
func hashRequest(r *http.Request) string {
	if r.Body == nil {
		return "empty"
	}

	// Read body and restore it
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "error"
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// extractCacheableHeaders extracts headers that should be cached.
func extractCacheableHeaders(headers http.Header) map[string]string {
	cacheable := make(map[string]string)
	cacheableKeys := []string{
		"Content-Type",
		"X-Request-Id",
		"Location",
	}

	for _, key := range cacheableKeys {
		if value := headers.Get(key); value != "" {
			cacheable[key] = value
		}
	}

	return cacheable
}

// getIdempotencyResponse retrieves a cached idempotency response.
func getIdempotencyResponse(ctx context.Context, key string) (*IdempotencyResponse, error) {
	data, err := cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var response IdempotencyResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// storeIdempotencyResponse stores an idempotency response in cache.
func storeIdempotencyResponse(ctx context.Context, key string, response *IdempotencyResponse, ttl time.Duration) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return cache.Set(ctx, key, data, ttl)
}

// responseRecorder captures the response for caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	written    bool
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if !r.written {
		r.statusCode = statusCode
		r.ResponseWriter.WriteHeader(statusCode)
		r.written = true
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter for middleware compatibility.
func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// ShouldApplyIdempotency determines if idempotency should be applied to a path.
// Returns true for paths that perform state-changing operations.
func ShouldApplyIdempotency(path string) bool {
	// Apply to payment-related endpoints
	if strings.Contains(path, "/billing/") || strings.Contains(path, "/payment") {
		return true
	}

	// Apply to user creation/updates
	if strings.HasPrefix(path, "/api/v1/users") {
		return true
	}

	// Apply to organization operations
	if strings.Contains(path, "/organizations") {
		return true
	}

	// Apply to file uploads
	if strings.Contains(path, "/files") {
		return true
	}

	return false
}
