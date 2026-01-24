package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/response"
)

// ============ Error Code Constants Tests ============

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ErrCodeUnauthorized", ErrCodeUnauthorized, response.ErrCodeUnauthorized},
		{"ErrCodeForbidden", ErrCodeForbidden, response.ErrCodeForbidden},
		{"ErrCodeNotFound", ErrCodeNotFound, response.ErrCodeNotFound},
		{"ErrCodeBadRequest", ErrCodeBadRequest, response.ErrCodeBadRequest},
		{"ErrCodeConflict", ErrCodeConflict, response.ErrCodeConflict},
		{"ErrCodeInternalError", ErrCodeInternalError, response.ErrCodeInternalError},
		{"ErrCodeValidation", ErrCodeValidation, response.ErrCodeValidation},
		{"ErrCodeRateLimited", ErrCodeRateLimited, response.ErrCodeRateLimited},
		{"ErrCodeTokenExpired", ErrCodeTokenExpired, response.ErrCodeTokenExpired},
		{"ErrCodeTokenInvalid", ErrCodeTokenInvalid, response.ErrCodeTokenInvalid},
		{"ErrCodeEmailNotVerified", ErrCodeEmailNotVerified, response.ErrCodeEmailNotVerified},
		{"ErrCodeAccountInactive", ErrCodeAccountInactive, response.ErrCodeAccountInactive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

// ============ WriteJSON Tests ============

func TestWriteJSON_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
	}{
		{"200 with map", http.StatusOK, map[string]string{"key": "value"}},
		{"201 with struct", http.StatusCreated, struct{ Name string }{"test"}},
		{"400 with string", http.StatusBadRequest, "error message"},
		{"500 with nil", http.StatusInternalServerError, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, tt.statusCode, tt.data)

			if w.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.statusCode)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want 'application/json'", ct)
			}
		})
	}
}

// ============ WriteSuccess Tests ============

func TestWriteSuccess_Variations(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    interface{}
	}{
		{"with message and data", "Success", map[string]string{"id": "123"}},
		{"with empty message", "", map[string]int{"count": 42}},
		{"with nil data", "Created", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteSuccess(w, tt.message, tt.data)

			if w.Code != http.StatusOK {
				t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

// ============ WriteError Tests ============

func TestWriteError_Codes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
	}{
		{"bad request", http.StatusBadRequest, ErrCodeBadRequest, "Invalid input"},
		{"not found", http.StatusNotFound, ErrCodeNotFound, "Resource not found"},
		{"internal error", http.StatusInternalServerError, ErrCodeInternalError, "Server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			WriteError(w, r, tt.statusCode, tt.code, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.statusCode)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if errObj, ok := resp["error"].(map[string]interface{}); ok {
				if errObj["code"] != tt.code {
					t.Errorf("error code = %v, want %q", errObj["code"], tt.code)
				}
			}
		})
	}
}

// ============ WriteBadRequest Tests ============

func TestWriteBadRequest_Status(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", nil)

	WriteBadRequest(w, r, "Invalid request body")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ WriteUnauthorized Tests ============

func TestWriteUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/protected", nil)

	WriteUnauthorized(w, r, "Authentication required")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ WriteForbidden Tests ============

func TestWriteForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/admin", nil)

	WriteForbidden(w, r, "Access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// ============ WriteNotFound Tests ============

func TestWriteNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/users/999", nil)

	WriteNotFound(w, r, "User not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ============ WriteConflict Tests ============

func TestWriteConflict(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/users", nil)

	WriteConflict(w, r, "Email already exists")

	if w.Code != http.StatusConflict {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusConflict)
	}
}

// ============ WriteInternalError Tests ============

func TestWriteInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/data", nil)

	WriteInternalError(w, r, "Database connection failed")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// ============ WriteValidationError Tests ============

func TestWriteValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", nil)

	WriteValidationError(w, r, "Email is required")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ WriteRateLimited Tests ============

func TestWriteRateLimited(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/data", nil)

	WriteRateLimited(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

// ============ Response Content-Type Tests ============

func TestResponseContentType(t *testing.T) {
	tests := []struct {
		name string
		fn   func(w http.ResponseWriter, r *http.Request)
	}{
		{"WriteBadRequest", func(w http.ResponseWriter, r *http.Request) { WriteBadRequest(w, r, "error") }},
		{"WriteUnauthorized", func(w http.ResponseWriter, r *http.Request) { WriteUnauthorized(w, r, "error") }},
		{"WriteForbidden", func(w http.ResponseWriter, r *http.Request) { WriteForbidden(w, r, "error") }},
		{"WriteNotFound", func(w http.ResponseWriter, r *http.Request) { WriteNotFound(w, r, "error") }},
		{"WriteConflict", func(w http.ResponseWriter, r *http.Request) { WriteConflict(w, r, "error") }},
		{"WriteInternalError", func(w http.ResponseWriter, r *http.Request) { WriteInternalError(w, r, "error") }},
		{"WriteValidationError", func(w http.ResponseWriter, r *http.Request) { WriteValidationError(w, r, "error") }},
		{"WriteRateLimited", func(w http.ResponseWriter, r *http.Request) { WriteRateLimited(w, r) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			tt.fn(w, r)

			ct := w.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("%s Content-Type = %q, want 'application/json'", tt.name, ct)
			}
		})
	}
}
