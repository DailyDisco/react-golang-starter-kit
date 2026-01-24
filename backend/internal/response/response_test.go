package response

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/contextkeys"
	"react-golang-starter/internal/models"
)

// ============ Constants Tests ============

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"unauthorized", ErrCodeUnauthorized, "UNAUTHORIZED"},
		{"forbidden", ErrCodeForbidden, "FORBIDDEN"},
		{"not found", ErrCodeNotFound, "NOT_FOUND"},
		{"bad request", ErrCodeBadRequest, "BAD_REQUEST"},
		{"conflict", ErrCodeConflict, "CONFLICT"},
		{"internal error", ErrCodeInternalError, "INTERNAL_ERROR"},
		{"validation", ErrCodeValidation, "VALIDATION_ERROR"},
		{"rate limited", ErrCodeRateLimited, "RATE_LIMITED"},
		{"token expired", ErrCodeTokenExpired, "TOKEN_EXPIRED"},
		{"token invalid", ErrCodeTokenInvalid, "TOKEN_INVALID"},
		{"email not verified", ErrCodeEmailNotVerified, "EMAIL_NOT_VERIFIED"},
		{"account inactive", ErrCodeAccountInactive, "ACCOUNT_INACTIVE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("constant = %q, want %q", tt.constant, tt.want)
			}
		})
	}
}

// ============ getRequestID Tests ============

func TestGetRequestID_WithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "test-request-123")
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	id := getRequestID(req)
	if id != "test-request-123" {
		t.Errorf("getRequestID() = %q, want %q", id, "test-request-123")
	}
}

func TestGetRequestID_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	id := getRequestID(req)
	if id != "" {
		t.Errorf("getRequestID() = %q, want empty string", id)
	}
}

func TestGetRequestID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, 12345)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	id := getRequestID(req)
	if id != "" {
		t.Errorf("getRequestID() = %q, want empty string for wrong type", id)
	}
}

// ============ JSON Tests ============

func TestJSON_WithData(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("result[key] = %q, want %q", result["key"], "value")
	}
}

func TestJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusNoContent, nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNoContent)
	}

	if w.Body.Len() != 0 {
		t.Errorf("body length = %d, want 0", w.Body.Len())
	}
}

func TestJSON_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			JSON(w, tt.statusCode, map[string]string{})

			if w.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.statusCode)
			}
		})
	}
}

// ============ Success Tests ============

func TestSuccess_BasicResponse(t *testing.T) {
	w := httptest.NewRecorder()

	Success(w, "Operation successful", map[string]int{"count": 42})

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result.Success {
		t.Error("Success = false, want true")
	}

	if result.Message != "Operation successful" {
		t.Errorf("Message = %q, want %q", result.Message, "Operation successful")
	}
}

func TestSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()

	Success(w, "Success with no data", nil)

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var result models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Data != nil {
		t.Errorf("Data = %v, want nil", result.Data)
	}
}

// ============ Error Tests ============

func TestError_BasicResponse(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	Error(w, req, http.StatusBadRequest, ErrCodeBadRequest, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var result models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Error != ErrCodeBadRequest {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeBadRequest)
	}

	if result.Message != "Invalid input" {
		t.Errorf("Message = %q, want %q", result.Message, "Invalid input")
	}

	if result.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want %d", result.Code, http.StatusBadRequest)
	}
}

func TestError_WithRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "req-abc-123")
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	Error(w, req, http.StatusInternalServerError, ErrCodeInternalError, "Something went wrong")

	var result models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.RequestID != "req-abc-123" {
		t.Errorf("RequestID = %q, want %q", result.RequestID, "req-abc-123")
	}
}

// ============ Common Error Response Tests ============

func TestBadRequest_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	BadRequest(w, req, "Bad request message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeBadRequest {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeBadRequest)
	}
}

func TestUnauthorized_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Unauthorized(w, req, "Not authenticated")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeUnauthorized {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeUnauthorized)
	}
}

func TestForbidden_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Forbidden(w, req, "Access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusForbidden)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeForbidden {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeForbidden)
	}
}

func TestNotFound_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	NotFound(w, req, "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeNotFound {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeNotFound)
	}
}

func TestConflict_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	Conflict(w, req, "Resource conflict")

	if w.Code != http.StatusConflict {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusConflict)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeConflict {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeConflict)
	}
}

func TestInternalError_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	InternalError(w, req, "Internal error message")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeInternalError {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeInternalError)
	}
}

func TestValidationError_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ValidationError(w, req, "Validation failed")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeValidation {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeValidation)
	}
}

func TestValidationErrorWithDetails_Response(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "req-123")
	req := httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)

	details := []models.FieldError{
		{Field: "email", Message: "Invalid email format"},
		{Field: "password", Message: "Password too short"},
	}

	ValidationErrorWithDetails(w, req, "Validation failed", details)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var result models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Error != ErrCodeValidation {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeValidation)
	}

	if result.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want %q", result.RequestID, "req-123")
	}

	if len(result.Details) != 2 {
		t.Errorf("Details length = %d, want 2", len(result.Details))
	}
}

func TestRateLimited_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	RateLimited(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusTooManyRequests)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeRateLimited {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeRateLimited)
	}
}

func TestTokenExpired_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	TokenExpired(w, req, "Token has expired")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeTokenExpired {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeTokenExpired)
	}
}

func TestTokenInvalid_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	TokenInvalid(w, req, "Token is invalid")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeTokenInvalid {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeTokenInvalid)
	}
}

func TestEmailNotVerified_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	EmailNotVerified(w, req, "Please verify your email")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeEmailNotVerified {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeEmailNotVerified)
	}
}

func TestAccountInactive_Response(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	AccountInactive(w, req, "Account is inactive")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeAccountInactive {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeAccountInactive)
	}
}
