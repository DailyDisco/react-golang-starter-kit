package response

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// SetCachePublic sets Cache-Control header for publicly cacheable responses
func SetCachePublic(w http.ResponseWriter, maxAge int) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
}

// SetCachePrivate sets Cache-Control header for user-specific cacheable responses
func SetCachePrivate(w http.ResponseWriter, maxAge int) {
	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", maxAge))
}

// SetNoCache sets headers to prevent caching
func SetNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetETag generates an ETag from data and checks If-None-Match header
// Returns true if the client already has the current version (304 should be sent)
func SetETag(w http.ResponseWriter, r *http.Request, data []byte) bool {
	etag := GenerateETag(data)
	w.Header().Set("ETag", etag)

	// Check if client has matching ETag
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
		return true
	}
	return false
}

// SetETagFromStruct generates an ETag from a struct and checks If-None-Match
// Returns true if client already has current version
func SetETagFromStruct(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	bytes, err := json.Marshal(data)
	if err != nil {
		return false
	}
	return SetETag(w, r, bytes)
}

// GenerateETag creates an ETag string from data
func GenerateETag(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf(`"%x"`, hash)
}

// SetLastModified sets the Last-Modified header
func SetLastModified(w http.ResponseWriter, t time.Time) {
	w.Header().Set("Last-Modified", t.UTC().Format(http.TimeFormat))
}

// CheckLastModified checks If-Modified-Since header
// Returns true if the resource hasn't been modified (304 should be sent)
func CheckLastModified(r *http.Request, lastModified time.Time) bool {
	if ims := r.Header.Get("If-Modified-Since"); ims != "" {
		if t, err := http.ParseTime(ims); err == nil {
			// Truncate to seconds for comparison (HTTP dates don't have sub-second precision)
			if !lastModified.Truncate(time.Second).After(t.Truncate(time.Second)) {
				return true
			}
		}
	}
	return false
}

// SendNotModified sends a 304 Not Modified response
func SendNotModified(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotModified)
}

// CacheConfig holds caching configuration for a response
type CacheConfig struct {
	MaxAge  int  // seconds
	Public  bool // true for public, false for private
	NoCache bool // if true, prevents caching
}

// Common cache configurations
var (
	CachePublic5Min   = CacheConfig{MaxAge: 300, Public: true}
	CachePublic1Hour  = CacheConfig{MaxAge: 3600, Public: true}
	CachePrivate1Min  = CacheConfig{MaxAge: 60, Public: false}
	CachePrivate5Min  = CacheConfig{MaxAge: 300, Public: false}
	CachePrivate1Hour = CacheConfig{MaxAge: 3600, Public: false}
	CacheNone         = CacheConfig{NoCache: true}
)

// ApplyCacheConfig applies a CacheConfig to the response
func ApplyCacheConfig(w http.ResponseWriter, config CacheConfig) {
	if config.NoCache {
		SetNoCache(w)
		return
	}
	if config.Public {
		SetCachePublic(w, config.MaxAge)
	} else {
		SetCachePrivate(w, config.MaxAge)
	}
}

// SetVary sets the Vary header to indicate which request headers affect caching
func SetVary(w http.ResponseWriter, headers ...string) {
	for _, h := range headers {
		w.Header().Add("Vary", h)
	}
}

// SetExpiresFromMaxAge sets the Expires header based on max-age
func SetExpiresFromMaxAge(w http.ResponseWriter, maxAge int) {
	expires := time.Now().Add(time.Duration(maxAge) * time.Second)
	w.Header().Set("Expires", expires.UTC().Format(http.TimeFormat))
}

// SetCacheHeaders is a convenience function that sets common cache headers
// It sets Cache-Control, and optionally ETag and Last-Modified
func SetCacheHeaders(w http.ResponseWriter, r *http.Request, config CacheConfig, data []byte, lastModified *time.Time) bool {
	// Apply cache control
	ApplyCacheConfig(w, config)

	// If no caching, we're done
	if config.NoCache {
		return false
	}

	// Set and check ETag if data is provided
	if data != nil {
		if SetETag(w, r, data) {
			return true // Client has current version
		}
	}

	// Set and check Last-Modified if provided
	if lastModified != nil {
		SetLastModified(w, *lastModified)
		if CheckLastModified(r, *lastModified) {
			return true // Client has current version
		}
	}

	// Set Expires header
	if !config.NoCache {
		SetExpiresFromMaxAge(w, config.MaxAge)
	}

	return false
}

// CacheControlFromSeconds builds a Cache-Control header value
func CacheControlFromSeconds(seconds int, public bool) string {
	visibility := "private"
	if public {
		visibility = "public"
	}
	return visibility + ", max-age=" + strconv.Itoa(seconds)
}
