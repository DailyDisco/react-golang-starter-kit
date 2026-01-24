package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"
)

// ============ UsageHandler Creation Tests ============

func TestNewUsageHandler(t *testing.T) {
	handler := NewUsageHandler(nil)
	if handler == nil {
		t.Error("NewUsageHandler() returned nil")
	}
}

// ============ GetCurrentUsage Tests ============

func TestUsageHandler_GetCurrentUsage_Unauthorized(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/usage", nil)
	w := httptest.NewRecorder()

	handler.GetCurrentUsage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetCurrentUsage() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestUsageHandler_GetCurrentUsage_ZeroUserID(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/usage", nil)
	w := httptest.NewRecorder()

	// Set user with ID 0 in context (should still be unauthorized)
	user := &models.User{ID: 0, Email: "test@example.com"}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.GetCurrentUsage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetCurrentUsage() with userID=0 status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ GetUsageHistory Tests ============

func TestUsageHandler_GetUsageHistory_Unauthorized(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/usage/history", nil)
	w := httptest.NewRecorder()

	handler.GetUsageHistory(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUsageHistory() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ GetAlerts Tests ============

func TestUsageHandler_GetAlerts_Unauthorized(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/usage/alerts", nil)
	w := httptest.NewRecorder()

	handler.GetAlerts(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetAlerts() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ AcknowledgeAlert Tests ============

func TestUsageHandler_AcknowledgeAlert_Unauthorized(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/usage/alerts/1/acknowledge", nil)
	w := httptest.NewRecorder()

	handler.AcknowledgeAlert(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AcknowledgeAlert() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestUsageHandler_AcknowledgeAlert_MissingAlertID(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/usage/alerts//acknowledge", nil)
	w := httptest.NewRecorder()

	// Add user to context using SetUserContext which sets all required context keys
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.AcknowledgeAlert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("AcknowledgeAlert() without alert ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUsageHandler_AcknowledgeAlert_InvalidAlertID(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/usage/alerts/abc/acknowledge", nil)
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()

	// Add user to context using SetUserContext which sets all required context keys
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.AcknowledgeAlert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("AcknowledgeAlert() with invalid alert ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ RecordUsage Tests ============

func TestUsageHandler_RecordUsage_Unauthorized(t *testing.T) {
	handler := NewUsageHandler(nil)

	body := `{"event_type": "test", "resource": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/usage/record", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RecordUsage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RecordUsage() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestUsageHandler_RecordUsage_InvalidJSON(t *testing.T) {
	handler := NewUsageHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/usage/record", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add user to context using SetUserContext which sets all required context keys
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.RecordUsage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RecordUsage() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUsageHandler_RecordUsage_MissingEventType(t *testing.T) {
	handler := NewUsageHandler(nil)

	body := `{"resource": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/usage/record", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add user to context using SetUserContext which sets all required context keys
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.RecordUsage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RecordUsage() without event_type status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUsageHandler_RecordUsage_MissingResource(t *testing.T) {
	handler := NewUsageHandler(nil)

	body := `{"event_type": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/usage/record", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add user to context using SetUserContext which sets all required context keys
	user := &models.User{ID: 1, Email: "test@example.com"}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.RecordUsage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RecordUsage() without resource status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUsageHandler_RecordUsage_TableDriven(t *testing.T) {
	handler := NewUsageHandler(nil)

	tests := []struct {
		name     string
		body     map[string]interface{}
		wantCode int
	}{
		{
			name:     "missing event_type",
			body:     map[string]interface{}{"resource": "test"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing resource",
			body:     map[string]interface{}{"event_type": "test"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty event_type",
			body:     map[string]interface{}{"event_type": "", "resource": "test"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty resource",
			body:     map[string]interface{}{"event_type": "test", "resource": ""},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/usage/record", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Add user to context using SetUserContext which sets all required context keys
			user := &models.User{ID: 1, Email: "test@example.com"}
			ctx := auth.SetUserContext(req.Context(), user)
			req = req.WithContext(ctx)

			handler.RecordUsage(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("RecordUsage() %s status = %v, want %v", tt.name, w.Code, tt.wantCode)
			}
		})
	}
}
