package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_GeneratesNewID(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request ID is in context
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("expected request ID in context, got empty string")
		}

		// Verify it's a valid UUID format (36 chars with hyphens)
		if len(requestID) != 36 {
			t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(requestID), requestID)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify response header is set
	responseRequestID := rec.Header().Get(RequestIDHeader)
	if responseRequestID == "" {
		t.Error("expected X-Request-ID header in response, got empty string")
	}
}

func TestRequestIDMiddleware_UsesExistingHeader(t *testing.T) {
	existingID := "test-request-id-12345"

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID != existingID {
			t.Errorf("expected request ID %q, got %q", existingID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify response header matches the provided ID
	responseRequestID := rec.Header().Get(RequestIDHeader)
	if responseRequestID != existingID {
		t.Errorf("expected response header %q, got %q", existingID, responseRequestID)
	}
}

func TestGetRequestID_ReturnsEmptyForMissingContext(t *testing.T) {
	ctx := context.Background()
	requestID := GetRequestID(ctx)

	if requestID != "" {
		t.Errorf("expected empty string for context without request ID, got %q", requestID)
	}
}

func TestGetRequestID_ReturnsValueFromContext(t *testing.T) {
	expectedID := "my-test-id"
	ctx := context.WithValue(context.Background(), RequestIDKey, expectedID)

	requestID := GetRequestID(ctx)
	if requestID != expectedID {
		t.Errorf("expected %q, got %q", expectedID, requestID)
	}
}

func TestGetRequestIDFromRequest(t *testing.T) {
	expectedID := "request-based-id"

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), RequestIDKey, expectedID)
	req = req.WithContext(ctx)

	requestID := GetRequestIDFromRequest(req)
	if requestID != expectedID {
		t.Errorf("expected %q, got %q", expectedID, requestID)
	}
}

func TestRequestIDMiddleware_SetsResponseHeader(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Response header should be set before the handler writes
	if rec.Header().Get(RequestIDHeader) == "" {
		t.Error("expected X-Request-ID response header to be set")
	}
}

func TestRequestIDMiddleware_ContextPassedToHandler(t *testing.T) {
	var capturedRequestID string

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// The captured ID should match the response header
	responseID := rec.Header().Get(RequestIDHeader)
	if capturedRequestID != responseID {
		t.Errorf("context request ID %q doesn't match response header %q", capturedRequestID, responseID)
	}
}

func TestRequestIDKey_TypeSafety(t *testing.T) {
	// Ensure the context key type prevents collisions
	ctx := context.WithValue(context.Background(), "request_id", "wrong-type")
	requestID := GetRequestID(ctx)

	// Should return empty because the key type is different
	if requestID != "" {
		t.Errorf("expected empty string when key type doesn't match, got %q", requestID)
	}

	// Now use the correct key type
	ctx = context.WithValue(context.Background(), RequestIDKey, "correct-id")
	requestID = GetRequestID(ctx)

	if requestID != "correct-id" {
		t.Errorf("expected 'correct-id', got %q", requestID)
	}
}
