package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/contextkeys"
	"react-golang-starter/internal/models"
)

// ============ RegisterUser Tests ============

func TestRegisterUser_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	RegisterUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "BAD_REQUEST" {
		t.Errorf("expected error code 'BAD_REQUEST', got '%s'", response.Error)
	}
}

func TestRegisterUser_InvalidEmail(t *testing.T) {
	reqBody := models.RegisterRequest{
		Name:     "Test User",
		Email:    "invalid-email",
		Password: "SecurePass123!",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	RegisterUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "BAD_REQUEST" {
		t.Errorf("expected error code 'BAD_REQUEST', got '%s'", response.Error)
	}
}

func TestRegisterUser_WeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"too short", "abc"},
		{"no uppercase", "password123"},
		{"no lowercase", "PASSWORD123"},
		{"no digit", "PasswordOnly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: tt.password,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			RegisterUser(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}
}

// ============ LoginUser Tests ============

func TestLoginUser_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	LoginUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "BAD_REQUEST" {
		t.Errorf("expected error code 'BAD_REQUEST', got '%s'", response.Error)
	}
}

// ============ VerifyEmail Tests ============

func TestVerifyEmail_MissingToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/auth/verify-email", nil)
	rec := httptest.NewRecorder()

	VerifyEmail(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "Verification token is required" {
		t.Errorf("expected message 'Verification token is required', got '%s'", response.Message)
	}
}

// ============ RequestPasswordReset Tests ============

func TestRequestPasswordReset_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	RequestPasswordReset(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// ============ ResetPassword Tests ============

func TestResetPassword_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/reset-password/confirm", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	reqBody := models.PasswordResetConfirm{
		Token:    "some-token",
		Password: "weak",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/auth/reset-password/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// ============ RefreshAccessToken Tests ============

func TestRefreshAccessToken_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	RefreshAccessToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRefreshAccessToken_MissingToken(t *testing.T) {
	reqBody := models.RefreshTokenRequest{
		RefreshToken: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	RefreshAccessToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "Refresh token is required" {
		t.Errorf("expected message 'Refresh token is required', got '%s'", response.Message)
	}
}

// ============ GetCurrentUser Tests ============

func TestGetCurrentUser_NoContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	GetCurrentUser(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "UNAUTHORIZED" {
		t.Errorf("expected error code 'UNAUTHORIZED', got '%s'", response.Error)
	}
}

// ============ Response Format Tests ============

func TestErrorResponse_IncludesRequestID(t *testing.T) {
	// Test that error responses include request_id field when context has it
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID to context (simulating middleware)
	expectedRequestID := "test-request-id-12345"
	ctx := context.WithValue(req.Context(), contextkeys.RequestIDKey, expectedRequestID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	RegisterUser(rec, req)

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// request_id field should exist and match the context value
	requestID, exists := response["request_id"]
	if !exists {
		t.Error("expected response to include 'request_id' field")
	}
	if requestID != expectedRequestID {
		t.Errorf("expected request_id %q, got %q", expectedRequestID, requestID)
	}
}

func TestErrorResponse_IncludesErrorCode(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	RegisterUser(rec, req)

	var response models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error == "" {
		t.Error("expected error code to be set")
	}

	if response.Code == 0 {
		t.Error("expected HTTP status code to be set")
	}

	if response.Message == "" {
		t.Error("expected error message to be set")
	}
}
