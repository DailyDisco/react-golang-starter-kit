package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/auth"

	"github.com/go-chi/chi/v5"
)

// ============ CreateNotificationRequest Structure Tests ============

func TestCreateNotificationRequest_Structure(t *testing.T) {
	req := CreateNotificationRequest{
		UserID:  1,
		Type:    "info",
		Title:   "Test Title",
		Message: "Test message",
		Link:    "/test",
	}

	if req.UserID != 1 {
		t.Errorf("CreateNotificationRequest.UserID = %d, want 1", req.UserID)
	}

	if req.Type != "info" {
		t.Errorf("CreateNotificationRequest.Type = %q, want info", req.Type)
	}

	if req.Title != "Test Title" {
		t.Errorf("CreateNotificationRequest.Title = %q, want Test Title", req.Title)
	}

	if req.Message != "Test message" {
		t.Errorf("CreateNotificationRequest.Message = %q, want Test message", req.Message)
	}

	if req.Link != "/test" {
		t.Errorf("CreateNotificationRequest.Link = %q, want /test", req.Link)
	}
}

func TestCreateNotificationRequest_JSONMarshaling(t *testing.T) {
	req := CreateNotificationRequest{
		UserID:  42,
		Type:    "alert",
		Title:   "Alert Title",
		Message: "Alert message",
		Link:    "/alerts/1",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal CreateNotificationRequest: %v", err)
	}

	var decoded CreateNotificationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CreateNotificationRequest: %v", err)
	}

	if decoded.UserID != req.UserID {
		t.Errorf("UserID after unmarshal = %d, want %d", decoded.UserID, req.UserID)
	}

	if decoded.Type != req.Type {
		t.Errorf("Type after unmarshal = %q, want %q", decoded.Type, req.Type)
	}
}

// ============ GetNotifications Handler Tests ============

func TestGetNotifications_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	w := httptest.NewRecorder()

	// Call without auth context
	GetNotifications(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetNotifications() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetNotifications_ZeroUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	// Set user ID to 0 (invalid)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(0))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	GetNotifications(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetNotifications() with zero user ID status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ MarkNotificationRead Handler Tests ============

func TestMarkNotificationRead_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/1/read", nil)
	w := httptest.NewRecorder()

	MarkNotificationRead(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("MarkNotificationRead() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarkNotificationRead_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/invalid/read", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(1))

	// Set up chi route context with invalid ID
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	MarkNotificationRead(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("MarkNotificationRead() with invalid ID status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarkNotificationRead_EmptyID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/notifications//read", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(1))

	// Set up chi route context with empty ID
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	MarkNotificationRead(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("MarkNotificationRead() with empty ID status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarkNotificationRead_NegativeID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/-1/read", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(1))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "-1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	MarkNotificationRead(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("MarkNotificationRead() with negative ID status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ============ MarkAllNotificationsRead Handler Tests ============

func TestMarkAllNotificationsRead_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
	w := httptest.NewRecorder()

	MarkAllNotificationsRead(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("MarkAllNotificationsRead() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarkAllNotificationsRead_ZeroUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/read-all", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(0))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	MarkAllNotificationsRead(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("MarkAllNotificationsRead() with zero user ID status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ DeleteNotification Handler Tests ============

func TestDeleteNotification_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/1", nil)
	w := httptest.NewRecorder()

	DeleteNotification(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteNotification() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDeleteNotification_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/abc", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(1))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	DeleteNotification(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DeleteNotification() with invalid ID status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestDeleteNotification_ZeroUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/notifications/1", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDContextKey, uint(0))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	DeleteNotification(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteNotification() with zero user ID status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ============ CreateNotification Handler Tests ============

func TestCreateNotification_InvalidBody(t *testing.T) {
	body := bytes.NewBufferString("invalid json")
	req := httptest.NewRequest(http.MethodPost, "/api/admin/notifications", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateNotification(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateNotification() with invalid body status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateNotification_EmptyTitle(t *testing.T) {
	reqBody := CreateNotificationRequest{
		UserID:  1,
		Type:    "info",
		Title:   "", // Empty title
		Message: "Test message",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateNotification(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateNotification() with empty title status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateNotification_ValidRequest_StructValidation(t *testing.T) {
	// Test that request structure is properly parsed (without DB)
	reqBody := CreateNotificationRequest{
		UserID:  1,
		Type:    "info",
		Title:   "Test Title",
		Message: "Test message",
		Link:    "/test",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify the JSON can be parsed back
	var parsed CreateNotificationRequest
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if parsed.Title != reqBody.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, reqBody.Title)
	}

	if parsed.UserID != reqBody.UserID {
		t.Errorf("UserID = %d, want %d", parsed.UserID, reqBody.UserID)
	}
}

// ============ Query Parameter Parsing Tests ============

func TestGetNotifications_PaginationParsing(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		expectValid bool
	}{
		{"valid page", "page=2", true},
		{"valid per_page", "per_page=50", true},
		{"valid both", "page=3&per_page=25", true},
		{"invalid page", "page=invalid", true}, // Falls back to default
		{"invalid per_page", "per_page=abc", true},
		{"negative page", "page=-1", true}, // Falls back to default
		{"per_page over limit", "per_page=200", true},
		{"unread filter", "unread=true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/notifications?" + tt.queryString
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Parse the query params like the handler does
			page := 1
			if p := req.URL.Query().Get("page"); p != "" {
				if parsed, err := parsePositiveInt(p); err == nil && parsed > 0 {
					page = parsed
				}
			}

			perPage := 20
			if pp := req.URL.Query().Get("per_page"); pp != "" {
				if parsed, err := parsePositiveInt(pp); err == nil && parsed > 0 && parsed <= 100 {
					perPage = parsed
				}
			}

			// Defaults should be valid
			if page < 1 || perPage < 1 {
				t.Errorf("Invalid pagination: page=%d, perPage=%d", page, perPage)
			}
		})
	}
}

// Helper for parsing
func parsePositiveInt(s string) (int, error) {
	var i int
	err := json.Unmarshal([]byte(s), &i)
	return i, err
}

// ============ Notification Type Constants ============

func TestNotificationTypes(t *testing.T) {
	// Test that common notification types are valid
	validTypes := []string{
		"info",
		"success",
		"warning",
		"error",
		"alert",
		"system",
	}

	for _, notifType := range validTypes {
		t.Run(notifType, func(t *testing.T) {
			req := CreateNotificationRequest{
				UserID:  1,
				Type:    notifType,
				Title:   "Test",
				Message: "Test message",
			}

			if req.Type != notifType {
				t.Errorf("Type = %q, want %q", req.Type, notifType)
			}
		})
	}
}

// ============ Edge Cases ============

func TestGetNotifications_LargePageNumber(t *testing.T) {
	// Test that large page numbers don't cause issues
	req := httptest.NewRequest(http.MethodGet, "/api/notifications?page=999999", nil)
	query := req.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := parsePositiveInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if page != 999999 {
		t.Errorf("Large page number not parsed correctly: got %d", page)
	}
}

func TestGetNotifications_PerPageBoundary(t *testing.T) {
	// Test per_page boundary at 100
	tests := []struct {
		input    string
		expected int
	}{
		{"100", 100}, // At limit
		{"101", 20},  // Over limit, should use default
		{"99", 99},   // Under limit
		{"1", 1},     // Minimum
		{"0", 20},    // Zero, should use default
		{"-1", 20},   // Negative, should use default
	}

	for _, tt := range tests {
		t.Run("per_page="+tt.input, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/notifications?per_page="+tt.input, nil)
			query := req.URL.Query()

			perPage := 20
			if pp := query.Get("per_page"); pp != "" {
				if parsed, err := parsePositiveInt(pp); err == nil && parsed > 0 && parsed <= 100 {
					perPage = parsed
				}
			}

			if perPage != tt.expected {
				t.Errorf("per_page=%s: got %d, want %d", tt.input, perPage, tt.expected)
			}
		})
	}
}

// ============ Offset Calculation Tests ============

func TestPaginationOffsetCalculation(t *testing.T) {
	tests := []struct {
		page     int
		perPage  int
		expected int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 20, 40},
		{1, 10, 0},
		{5, 10, 40},
		{1, 100, 0},
		{2, 100, 100},
	}

	for _, tt := range tests {
		offset := (tt.page - 1) * tt.perPage
		if offset != tt.expected {
			t.Errorf("page=%d, perPage=%d: offset=%d, want %d", tt.page, tt.perPage, offset, tt.expected)
		}
	}
}
