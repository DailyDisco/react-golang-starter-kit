package response

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetCachePublic(t *testing.T) {
	w := httptest.NewRecorder()
	SetCachePublic(w, 300)

	assert.Equal(t, "public, max-age=300", w.Header().Get("Cache-Control"))
}

func TestSetCachePrivate(t *testing.T) {
	w := httptest.NewRecorder()
	SetCachePrivate(w, 60)

	assert.Equal(t, "private, max-age=60", w.Header().Get("Cache-Control"))
}

func TestSetNoCache(t *testing.T) {
	w := httptest.NewRecorder()
	SetNoCache(w)

	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))
}

func TestGenerateETag(t *testing.T) {
	data := []byte(`{"key":"value"}`)
	etag := GenerateETag(data)

	// Should be quoted hex string
	assert.Regexp(t, `^"[a-f0-9]{32}"$`, etag)

	// Same data should produce same ETag
	etag2 := GenerateETag(data)
	assert.Equal(t, etag, etag2)

	// Different data should produce different ETag
	etag3 := GenerateETag([]byte(`{"key":"different"}`))
	assert.NotEqual(t, etag, etag3)
}

func TestSetETag_NoMatch(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	data := []byte(`{"test":"data"}`)
	notModified := SetETag(w, r, data)

	assert.False(t, notModified)
	assert.NotEmpty(t, w.Header().Get("ETag"))
}

func TestSetETag_Match(t *testing.T) {
	data := []byte(`{"test":"data"}`)
	etag := GenerateETag(data)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("If-None-Match", etag)

	notModified := SetETag(w, r, data)

	assert.True(t, notModified)
}

func TestSetETag_NoMatchDifferent(t *testing.T) {
	data := []byte(`{"test":"data"}`)
	oldEtag := GenerateETag([]byte(`{"old":"data"}`))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("If-None-Match", oldEtag)

	notModified := SetETag(w, r, data)

	assert.False(t, notModified)
}

func TestSetETagFromStruct(t *testing.T) {
	type TestData struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	data := TestData{Key: "test", Value: 42}
	notModified := SetETagFromStruct(w, r, data)

	assert.False(t, notModified)
	assert.NotEmpty(t, w.Header().Get("ETag"))
}

func TestSetLastModified(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	SetLastModified(w, now)

	assert.Equal(t, "Sun, 15 Jun 2025 10:30:00 GMT", w.Header().Get("Last-Modified"))
}

func TestCheckLastModified_NotModified(t *testing.T) {
	lastMod := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("If-Modified-Since", "Sun, 15 Jun 2025 10:30:00 GMT") // After lastMod

	notModified := CheckLastModified(r, lastMod)
	assert.True(t, notModified)
}

func TestCheckLastModified_Modified(t *testing.T) {
	lastMod := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("If-Modified-Since", "Sun, 15 Jun 2025 10:30:00 GMT") // Before lastMod

	notModified := CheckLastModified(r, lastMod)
	assert.False(t, notModified)
}

func TestSendNotModified(t *testing.T) {
	w := httptest.NewRecorder()
	SendNotModified(w)

	assert.Equal(t, http.StatusNotModified, w.Code)
}

func TestApplyCacheConfig_Public(t *testing.T) {
	w := httptest.NewRecorder()
	ApplyCacheConfig(w, CachePublic5Min)

	assert.Equal(t, "public, max-age=300", w.Header().Get("Cache-Control"))
}

func TestApplyCacheConfig_Private(t *testing.T) {
	w := httptest.NewRecorder()
	ApplyCacheConfig(w, CachePrivate1Min)

	assert.Equal(t, "private, max-age=60", w.Header().Get("Cache-Control"))
}

func TestApplyCacheConfig_NoCache(t *testing.T) {
	w := httptest.NewRecorder()
	ApplyCacheConfig(w, CacheNone)

	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
}

func TestSetVary(t *testing.T) {
	w := httptest.NewRecorder()
	SetVary(w, "Accept", "Authorization")

	// Headers can have multiple values
	values := w.Header().Values("Vary")
	assert.Contains(t, values, "Accept")
	assert.Contains(t, values, "Authorization")
}

func TestSetCacheHeaders_Full(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	data := []byte(`{"test":"data"}`)
	lastMod := time.Now().Add(-1 * time.Hour)

	notModified := SetCacheHeaders(w, r, CachePrivate5Min, data, &lastMod)

	assert.False(t, notModified)
	assert.Equal(t, "private, max-age=300", w.Header().Get("Cache-Control"))
	assert.NotEmpty(t, w.Header().Get("ETag"))
	assert.NotEmpty(t, w.Header().Get("Last-Modified"))
	assert.NotEmpty(t, w.Header().Get("Expires"))
}

func TestSetCacheHeaders_NotModifiedByETag(t *testing.T) {
	data := []byte(`{"test":"data"}`)
	etag := GenerateETag(data)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("If-None-Match", etag)

	notModified := SetCacheHeaders(w, r, CachePrivate5Min, data, nil)

	assert.True(t, notModified)
}

func TestCacheControlFromSeconds(t *testing.T) {
	tests := []struct {
		seconds  int
		public   bool
		expected string
	}{
		{300, true, "public, max-age=300"},
		{60, false, "private, max-age=60"},
		{3600, true, "public, max-age=3600"},
	}

	for _, tt := range tests {
		result := CacheControlFromSeconds(tt.seconds, tt.public)
		assert.Equal(t, tt.expected, result)
	}
}
