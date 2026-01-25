package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============ RecoveryMiddleware Tests ============

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Body.String() != "OK" {
		t.Errorf("body = %q, want 'OK'", w.Body.String())
	}
}

func TestRecoveryMiddleware_RecoverFromPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	// Should return 500 Internal Server Error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	// Should return JSON error response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, hasError := response["error"]; !hasError {
		t.Error("response should have 'error' field")
	}
}

func TestRecoveryMiddleware_RecoverFromNilPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic (even with nil panic value)
	// Note: panic(nil) may or may not be recovered depending on Go version
	// In Go 1.21+, panic(nil) does not trigger recover()
	handler.ServeHTTP(w, req)
}

func TestRecoveryMiddleware_RecoverFromErrorPanic(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("database connection failed")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestRecoveryMiddleware_DoesNotExposeInternalDetails(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("super secret internal error with password=123")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	body := w.Body.String()

	// Should not expose internal error details
	if contains(body, "secret") || contains(body, "password") {
		t.Error("response should not expose internal error details")
	}
}

// ============ RecoveryMiddlewareWithCallback Tests ============

func TestRecoveryMiddlewareWithCallback_NoPanic(t *testing.T) {
	callbackCalled := false
	handler := RecoveryMiddlewareWithCallback(func(r *http.Request, err interface{}, stack []byte) {
		callbackCalled = true
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if callbackCalled {
		t.Error("callback should not be called when there's no panic")
	}
}

func TestRecoveryMiddlewareWithCallback_CallsCallback(t *testing.T) {
	var capturedErr interface{}
	var capturedStack []byte
	var capturedRequest *http.Request

	handler := RecoveryMiddlewareWithCallback(func(r *http.Request, err interface{}, stack []byte) {
		capturedErr = err
		capturedStack = stack
		capturedRequest = r
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("callback test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/callback-test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedErr == nil {
		t.Error("callback should receive the panic error")
	}

	if capturedErr != "callback test panic" {
		t.Errorf("capturedErr = %v, want 'callback test panic'", capturedErr)
	}

	if len(capturedStack) == 0 {
		t.Error("callback should receive stack trace")
	}

	if capturedRequest == nil {
		t.Error("callback should receive the request")
	}

	if capturedRequest.URL.Path != "/callback-test" {
		t.Errorf("request path = %q, want '/callback-test'", capturedRequest.URL.Path)
	}
}

func TestRecoveryMiddlewareWithCallback_NilCallback(t *testing.T) {
	handler := RecoveryMiddlewareWithCallback(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("nil callback test")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic even with nil callback
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestRecoveryMiddlewareWithCallback_ReturnsErrorResponse(t *testing.T) {
	handler := RecoveryMiddlewareWithCallback(func(r *http.Request, err interface{}, stack []byte) {
		// Custom handling - just count or log
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("response test")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	// Should still return JSON error response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
}

// ============ Edge Cases ============

func TestRecoveryMiddleware_StackContainsUsefulInfo(t *testing.T) {
	var capturedStack []byte

	handler := RecoveryMiddlewareWithCallback(func(r *http.Request, err interface{}, stack []byte) {
		capturedStack = stack
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("stack trace test")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	stackStr := string(capturedStack)

	// Stack should contain goroutine info
	if !contains(stackStr, "goroutine") {
		t.Error("stack trace should contain 'goroutine'")
	}
}

func TestRecoveryMiddleware_MultiplePanicsHandledIndependently(t *testing.T) {
	callCount := 0
	handler := RecoveryMiddlewareWithCallback(func(r *http.Request, err interface{}, stack []byte) {
		callCount++
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("multi panic test")
	}))

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/test1", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	// Second request
	req2 := httptest.NewRequest(http.MethodGet, "/test2", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if callCount != 2 {
		t.Errorf("callback called %d times, want 2", callCount)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
