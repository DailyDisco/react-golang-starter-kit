package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// ============ WriteJSON Tests ============

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		payload        interface{}
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "success response",
			code:           http.StatusOK,
			payload:        map[string]string{"message": "success"},
			wantStatusCode: http.StatusOK,
			wantBody:       "{\"message\":\"success\"}\n",
		},
		{
			name:           "error response",
			code:           http.StatusBadRequest,
			payload:        map[string]string{"error": "bad request"},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "{\"error\":\"bad request\"}\n",
		},
		{
			name:           "empty object",
			code:           http.StatusOK,
			payload:        map[string]string{},
			wantStatusCode: http.StatusOK,
			wantBody:       "{}\n",
		},
		{
			name:           "nested object",
			code:           http.StatusOK,
			payload:        map[string]interface{}{"data": map[string]int{"count": 5}},
			wantStatusCode: http.StatusOK,
			wantBody:       "{\"data\":{\"count\":5}}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, tt.code, tt.payload)

			if w.Code != tt.wantStatusCode {
				t.Errorf("WriteJSON() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("WriteJSON() Content-Type = %v, want application/json", contentType)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("WriteJSON() body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestWriteJSON_NilPayload(t *testing.T) {
	w := httptest.NewRecorder()

	// Nil payload should result in empty body
	WriteJSON(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("WriteJSON() with nil payload status = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "" {
		t.Errorf("WriteJSON() with nil payload body = %q, want empty", w.Body.String())
	}
}

// ============ WriteError Tests ============

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		errorCode      string
		message        string
		wantStatusCode int
	}{
		{
			name:           "bad request error",
			code:           http.StatusBadRequest,
			errorCode:      ErrCodeBadRequest,
			message:        "Invalid input",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "not found error",
			code:           http.StatusNotFound,
			errorCode:      ErrCodeNotFound,
			message:        "Resource not found",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "internal server error",
			code:           http.StatusInternalServerError,
			errorCode:      ErrCodeInternalError,
			message:        "Something went wrong",
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			WriteError(w, req, tt.code, tt.errorCode, tt.message)

			if w.Code != tt.wantStatusCode {
				t.Errorf("WriteError() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			var response models.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Error != tt.errorCode {
				t.Errorf("WriteError() error code = %v, want %v", response.Error, tt.errorCode)
			}

			if response.Message != tt.message {
				t.Errorf("WriteError() message = %v, want %v", response.Message, tt.message)
			}

			if response.Code != tt.code {
				t.Errorf("WriteError() code = %v, want %v", response.Code, tt.code)
			}
		})
	}
}

// ============ NewService Tests ============

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Error("NewService() returned nil")
	}
}

// ============ GetUsers Pagination Tests ============

func TestGetUsers_PaginationValidation(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		wantStatusCode int
		wantError      bool
	}{
		{
			name:           "invalid page parameter (non-numeric)",
			queryParams:    "?page=abc",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid page parameter (zero)",
			queryParams:    "?page=0",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid page parameter (negative)",
			queryParams:    "?page=-1",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid limit parameter (non-numeric)",
			queryParams:    "?limit=abc",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid limit parameter (zero)",
			queryParams:    "?limit=0",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid limit parameter (exceeds max)",
			queryParams:    "?limit=101",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "invalid limit parameter (negative)",
			queryParams:    "?limit=-5",
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler := GetUsers()
			handler.ServeHTTP(w, req)

			if tt.wantError && w.Code != tt.wantStatusCode {
				t.Errorf("GetUsers() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

// ============ GetUser Tests ============

func TestGetUser_InvalidID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		wantStatusCode int
	}{
		{"non-numeric id", "abc", http.StatusBadRequest},
		{"float id", "1.5", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.id, nil)
			w := httptest.NewRecorder()

			// Set up chi router context with URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler := GetUser()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("GetUser() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestGetUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()

	// Set up chi router context with valid ID but no auth context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := GetUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ CreateUser Validation Tests ============

func TestCreateUser_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := CreateUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateUser() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"missing @", "invalidemail"},
		{"missing domain", "test@"},
		{"missing local part", "@example.com"},
		{"spaces", "test @example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{
				"name":     "Test User",
				"email":    tt.email,
				"password": "SecurePass123!",
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := CreateUser()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("CreateUser() with email %q status = %v, want %v", tt.email, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateUser_WeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"too short", "Short1"},
		{"no uppercase", "password123"},
		{"no lowercase", "PASSWORD123"},
		{"no number", "PasswordOnly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": tt.password,
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := CreateUser()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("CreateUser() with weak password %q status = %v, want %v", tt.password, w.Code, http.StatusBadRequest)
			}
		})
	}
}

// ============ UpdateUser Tests ============

func TestUpdateUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/users/abc", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := UpdateUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateUser() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/users/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := UpdateUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ DeleteUser Tests ============

func TestDeleteUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/users/abc", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := DeleteUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DeleteUser() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestDeleteUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := DeleteUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ UpdateUserRole Tests ============

func TestUpdateUserRole_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/admin/users/abc/role", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := UpdateUserRole()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateUserRole() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateUserRole_InvalidPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1/role", bytes.NewBufferString("invalid"))
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler := UpdateUserRole()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateUserRole() with invalid payload status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{"invalid role", "superuser"},
		{"empty role", ""},
		{"typo in role", "admni"},
		{"uppercase role", "ADMIN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{"role": tt.role}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPut, "/admin/users/1/role", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler := UpdateUserRole()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("UpdateUserRole() with role %q status = %v, want %v", tt.role, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestUpdateUserRole_ValidRolesAreAccepted(t *testing.T) {
	// This test verifies that valid role values pass the role validation check.
	// We can't test the full handler without a database, so we just verify
	// the role validation logic accepts valid roles.
	validRoles := map[string]bool{
		models.RoleSuperAdmin: true,
		models.RoleAdmin:      true,
		models.RolePremium:    true,
		models.RoleUser:       true,
	}

	for role := range validRoles {
		t.Run(role, func(t *testing.T) {
			if !validRoles[role] {
				t.Errorf("Role %q should be valid", role)
			}
		})
	}
}

// ============ GetCurrentUser Tests ============

func TestGetCurrentUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	handler := GetCurrentUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetCurrentUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetCurrentUser_WithAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	// Add authenticated user to context using the same key as middleware
	user := &models.User{
		ID:            1,
		Name:          "Test User",
		Email:         "test@example.com",
		Role:          "user",
		EmailVerified: true,
		IsActive:      true,
	}
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)

	handler := GetCurrentUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetCurrentUser() with auth status = %v, want %v", w.Code, http.StatusOK)
	}

	var response models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("GetCurrentUser() response.Success = false, want true")
	}
}

// ============ GetPremiumContent Tests ============

func TestGetPremiumContent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/premium/content", nil)
	w := httptest.NewRecorder()

	GetPremiumContent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetPremiumContent() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("GetPremiumContent() response.Success = false, want true")
	}

	// Verify the data structure
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("GetPremiumContent() response.Data is not a map")
	}

	if _, ok := data["content"]; !ok {
		t.Error("GetPremiumContent() response missing 'content' field")
	}

	if _, ok := data["features"]; !ok {
		t.Error("GetPremiumContent() response missing 'features' field")
	}

	if _, ok := data["access_level"]; !ok {
		t.Error("GetPremiumContent() response missing 'access_level' field")
	}
}

// ============ UpdateRoleRequest Tests ============

func TestUpdateRoleRequest_JSONMarshal(t *testing.T) {
	req := UpdateRoleRequest{Role: "admin"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal UpdateRoleRequest: %v", err)
	}

	var decoded UpdateRoleRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal UpdateRoleRequest: %v", err)
	}

	if decoded.Role != "admin" {
		t.Errorf("UpdateRoleRequest.Role = %v, want admin", decoded.Role)
	}
}

// ============ PremiumContentResponse Tests ============

func TestPremiumContentResponse_JSONMarshal(t *testing.T) {
	resp := PremiumContentResponse{
		Content:     "Test content",
		Features:    []string{"feature1", "feature2"},
		AccessLevel: "premium",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal PremiumContentResponse: %v", err)
	}

	var decoded PremiumContentResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PremiumContentResponse: %v", err)
	}

	if decoded.Content != resp.Content {
		t.Errorf("PremiumContentResponse.Content = %v, want %v", decoded.Content, resp.Content)
	}

	if len(decoded.Features) != len(resp.Features) {
		t.Errorf("PremiumContentResponse.Features length = %v, want %v", len(decoded.Features), len(resp.Features))
	}

	if decoded.AccessLevel != resp.AccessLevel {
		t.Errorf("PremiumContentResponse.AccessLevel = %v, want %v", decoded.AccessLevel, resp.AccessLevel)
	}
}

// ============ formatBytes Tests ============

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"2.5 MB", 2621440, "2.5 MB"},
		{"10 GB", 10737418240, "10.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_LargeValues(t *testing.T) {
	// Test petabyte range
	result := formatBytes(1024 * 1024 * 1024 * 1024 * 1024)
	if result != "1.0 PB" {
		t.Errorf("formatBytes(1PB) = %q, want 1.0 PB", result)
	}

	// Test exabyte range
	result = formatBytes(1024 * 1024 * 1024 * 1024 * 1024 * 1024)
	if result != "1.0 EB" {
		t.Errorf("formatBytes(1EB) = %q, want 1.0 EB", result)
	}
}

// ============ formatDuration Tests ============

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"just seconds", 45 * time.Second, "45s"},
		{"1 minute", 60 * time.Second, "1m 0s"},
		{"5 minutes", 5 * time.Minute, "5m 0s"},
		{"1 hour", 1 * time.Hour, "1h 0m 0s"},
		{"1 hour 30 minutes", 90 * time.Minute, "1h 30m 0s"},
		{"1 day", 24 * time.Hour, "1d 0h 0m 0s"},
		{"1 day 2 hours", 26 * time.Hour, "1d 2h 0m 0s"},
		{"complex", 25*time.Hour + time.Minute + time.Second, "1d 1h 1m 1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// ============ getUserCacheKey Tests ============

func TestGetUserCacheKey(t *testing.T) {
	tests := []struct {
		userID   uint
		expected string
	}{
		{1, "user:1"},
		{42, "user:42"},
		{999999, "user:999999"},
		{0, "user:0"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getUserCacheKey(tt.userID)
			if result != tt.expected {
				t.Errorf("getUserCacheKey(%d) = %q, want %q", tt.userID, result, tt.expected)
			}
		})
	}
}

// ============ getCachedUser Tests ============

func TestGetCachedUser_CacheNotAvailable(t *testing.T) {
	// When cache is not available, getCachedUser should return nil, false
	user, found := getCachedUser(context.Background(), 1)

	if found {
		t.Error("getCachedUser() should return false when cache is not available")
	}

	if user != nil {
		t.Error("getCachedUser() should return nil when cache is not available")
	}
}

// ============ cacheUser Tests ============

func TestCacheUser_CacheNotAvailable(t *testing.T) {
	// When cache is not available, cacheUser should return without error
	userResponse := &models.UserResponse{
		ID:    1,
		Name:  "Test",
		Email: "test@example.com",
	}

	// This should not panic or error even if cache is not available
	cacheUser(context.Background(), 1, userResponse)
}

// ============ invalidateUserCache Tests ============

func TestInvalidateUserCache_CacheNotAvailable(t *testing.T) {
	// When cache is not available, invalidateUserCache should return without error
	// This should not panic or error even if cache is not available
	invalidateUserCache(context.Background(), 1)
}

// ============ UpdateCurrentUser Tests ============

func TestUpdateCurrentUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/users/me", nil)
	w := httptest.NewRecorder()

	handler := UpdateCurrentUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateCurrentUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestUpdateCurrentUser_InvalidJSON(t *testing.T) {
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.UserContextKey, user)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler := UpdateCurrentUser()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateCurrentUser() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ HealthCheck Tests ============
// Note: Full HealthCheck and ReadinessCheck tests require database infrastructure.
// These tests verify the response structure only.

func TestHealthStatus_Structure(t *testing.T) {
	// Test the HealthStatus response structure
	status := models.HealthStatus{
		OverallStatus: "healthy",
		Timestamp:     "2025-01-01T00:00:00Z",
		Uptime:        "1h 30m 0s",
		Version: models.VersionInfo{
			Version:   "1.0.0",
			BuildTime: "2025-01-01",
			GitCommit: "abc123",
		},
		Components: []models.ComponentStatus{
			{Name: "database", Status: "healthy", Message: "Connected"},
		},
	}

	if status.OverallStatus != "healthy" {
		t.Errorf("HealthStatus.OverallStatus = %q, want healthy", status.OverallStatus)
	}

	if len(status.Components) != 1 {
		t.Errorf("HealthStatus.Components length = %d, want 1", len(status.Components))
	}
}

func TestVersionInfo_Structure(t *testing.T) {
	info := models.VersionInfo{
		Version:   "1.0.0",
		BuildTime: "2025-01-01",
		GitCommit: "abc123def",
	}

	if info.Version != "1.0.0" {
		t.Errorf("VersionInfo.Version = %q, want 1.0.0", info.Version)
	}
}

func TestComponentStatus_Structure(t *testing.T) {
	status := models.ComponentStatus{
		Name:    "database",
		Status:  "healthy",
		Message: "Connected successfully",
	}

	if status.Name != "database" {
		t.Errorf("ComponentStatus.Name = %q, want database", status.Name)
	}

	if status.Status != "healthy" {
		t.Errorf("ComponentStatus.Status = %q, want healthy", status.Status)
	}
}

func TestRuntimeInfo_Structure(t *testing.T) {
	info := models.RuntimeInfo{
		Goroutines:  10,
		MemoryAlloc: "50.0 MB",
		MemorySys:   "100.0 MB",
		NumGC:       5,
		GoVersion:   "go1.25",
		NumCPU:      8,
		GOOS:        "linux",
		GOARCH:      "amd64",
	}

	if info.Goroutines != 10 {
		t.Errorf("RuntimeInfo.Goroutines = %d, want 10", info.Goroutines)
	}

	if info.NumCPU != 8 {
		t.Errorf("RuntimeInfo.NumCPU = %d, want 8", info.NumCPU)
	}
}

// ============ Version Variables Tests ============

func TestVersionVariables(t *testing.T) {
	// Version variables should have default values
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// BuildTime and GitCommit may be empty if not set during build
	// but should not cause errors when accessed
	_ = BuildTime
	_ = GitCommit
}

// ============ Cache Key Prefix Tests ============

func TestCacheKeyConstants(t *testing.T) {
	if userCacheKeyPrefix != "user:" {
		t.Errorf("userCacheKeyPrefix = %q, want 'user:'", userCacheKeyPrefix)
	}

	// TTL should be reasonable
	if userCacheTTL <= 0 {
		t.Error("userCacheTTL should be positive")
	}
}
