package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/response"
)

// ============ Error Code Constants Tests ============

func TestAuthErrorCodeConstants(t *testing.T) {
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

// ============ writeJSON Tests ============

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
	}{
		{"200 with map", http.StatusOK, map[string]string{"key": "value"}},
		{"201 with struct", http.StatusCreated, struct{ Name string }{"test"}},
		{"400 with string", http.StatusBadRequest, "error message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tt.statusCode, tt.data)

			if w.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.statusCode)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want 'application/json'", ct)
			}
		})
	}
}

// ============ writeSuccess Tests ============

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	writeSuccess(w, "Login successful", map[string]string{"user": "test"})

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "Login successful" {
		t.Errorf("message = %v, want 'Login successful'", resp["message"])
	}
}

// ============ writeError Tests ============

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	writeError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, "Invalid credentials")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ writeBadRequest Tests ============

func TestWriteBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/register", nil)

	writeBadRequest(w, r, "Invalid request body")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ writeUnauthorized Tests ============

func TestWriteUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	writeUnauthorized(w, r, "Authentication required")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ writeForbidden Tests ============

func TestWriteForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/auth/admin", nil)

	writeForbidden(w, r, "Admin access required")

	if w.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// ============ writeNotFound Tests ============

func TestWriteNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/auth/user/999", nil)

	writeNotFound(w, r, "User not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ============ writeConflict Tests ============

func TestWriteConflict(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/register", nil)

	writeConflict(w, r, "Email already registered")

	if w.Code != http.StatusConflict {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusConflict)
	}
}

// ============ writeInternalError Tests ============

func TestWriteInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	writeInternalError(w, r, "Database connection failed")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// ============ writeValidationError Tests ============

func TestWriteValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/register", nil)

	writeValidationError(w, r, "Email is required")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ writeTokenExpired Tests ============

func TestWriteTokenExpired(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	writeTokenExpired(w, r, "Access token has expired")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ writeTokenInvalid Tests ============

func TestWriteTokenInvalid(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	writeTokenInvalid(w, r, "Invalid access token")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ writeEmailNotVerified Tests ============

func TestWriteEmailNotVerified(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	writeEmailNotVerified(w, r, "Please verify your email")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ writeAccountInactive Tests ============

func TestWriteAccountInactive(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	writeAccountInactive(w, r, "Account is deactivated")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ Content-Type Tests ============

func TestAuthResponseContentType(t *testing.T) {
	tests := []struct {
		name string
		fn   func(w http.ResponseWriter, r *http.Request)
	}{
		{"writeBadRequest", func(w http.ResponseWriter, r *http.Request) { writeBadRequest(w, r, "error") }},
		{"writeUnauthorized", func(w http.ResponseWriter, r *http.Request) { writeUnauthorized(w, r, "error") }},
		{"writeForbidden", func(w http.ResponseWriter, r *http.Request) { writeForbidden(w, r, "error") }},
		{"writeNotFound", func(w http.ResponseWriter, r *http.Request) { writeNotFound(w, r, "error") }},
		{"writeConflict", func(w http.ResponseWriter, r *http.Request) { writeConflict(w, r, "error") }},
		{"writeInternalError", func(w http.ResponseWriter, r *http.Request) { writeInternalError(w, r, "error") }},
		{"writeTokenExpired", func(w http.ResponseWriter, r *http.Request) { writeTokenExpired(w, r, "error") }},
		{"writeTokenInvalid", func(w http.ResponseWriter, r *http.Request) { writeTokenInvalid(w, r, "error") }},
		{"writeEmailNotVerified", func(w http.ResponseWriter, r *http.Request) { writeEmailNotVerified(w, r, "error") }},
		{"writeAccountInactive", func(w http.ResponseWriter, r *http.Request) { writeAccountInactive(w, r, "error") }},
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
