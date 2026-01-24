package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// ============ Settings Handlers Tests ============

func TestGetSettingsByCategory_MissingCategory(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/settings/", nil)
	w := httptest.NewRecorder()

	// Create router to handle URL params
	r := chi.NewRouter()
	r.Get("/settings/{category}", GetSettingsByCategory)

	// Request without category param
	req = httptest.NewRequest(http.MethodGet, "/settings/", nil)
	w = httptest.NewRecorder()

	// Use empty string as category
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("category", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	GetSettingsByCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetSettingsByCategory() with empty category status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateSetting_MissingKey(t *testing.T) {
	body := `{"value": "test"}`
	req := httptest.NewRequest(http.MethodPut, "/settings/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Use empty string as key
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateSetting(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateSetting() with empty key status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateSetting_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/settings/test-key", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("key", "test-key")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateSetting(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateSetting() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ Email Settings Tests ============

func TestUpdateEmailSettings_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/settings/email", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateEmailSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateEmailSettings() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestTestEmailSettings_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/settings/email/test", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	TestEmailSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("TestEmailSettings() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ Security Settings Tests ============

func TestUpdateSecuritySettings_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/settings/security", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateSecuritySettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateSecuritySettings() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ Site Settings Tests ============

func TestUpdateSiteSettings_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/settings/site", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateSiteSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateSiteSettings() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ IP Blocklist Tests ============

func TestBlockIP_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ip-blocklist", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	BlockIP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("BlockIP() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestBlockIP_MissingIPAddress(t *testing.T) {
	body := `{"reason": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/ip-blocklist", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	BlockIP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("BlockIP() without IP address status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUnblockIP_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/ip-blocklist/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UnblockIP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UnblockIP() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ Announcement Tests ============

func TestCreateAnnouncement_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/announcements", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateAnnouncement(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateAnnouncement() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCreateAnnouncement_MissingTitleAndMessage(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "missing title",
			body: map[string]interface{}{"message": "test message"},
		},
		{
			name: "missing message",
			body: map[string]interface{}{"title": "test title"},
		},
		{
			name: "both empty",
			body: map[string]interface{}{"title": "", "message": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/announcements", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			CreateAnnouncement(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("CreateAnnouncement() %s status = %v, want %v", tt.name, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestUpdateAnnouncement_InvalidID(t *testing.T) {
	body := `{"title": "test"}`
	req := httptest.NewRequest(http.MethodPut, "/announcements/invalid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateAnnouncement(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateAnnouncement() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateAnnouncement_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/announcements/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateAnnouncement(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateAnnouncement() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestDeleteAnnouncement_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/announcements/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	DeleteAnnouncement(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DeleteAnnouncement() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestDismissAnnouncement_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/announcements/invalid/dismiss", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	DismissAnnouncement(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DismissAnnouncement() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestDismissAnnouncement_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/dismiss", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// No user in context
	DismissAnnouncement(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DismissAnnouncement() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestMarkAnnouncementRead_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/announcements/invalid/read", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	MarkAnnouncementRead(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("MarkAnnouncementRead() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestMarkAnnouncementRead_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/read", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// No user in context
	MarkAnnouncementRead(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("MarkAnnouncementRead() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUnreadModalAnnouncements_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/announcements/unread-modals", nil)
	w := httptest.NewRecorder()

	// No user in context
	GetUnreadModalAnnouncements(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUnreadModalAnnouncements() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ Email Template Tests ============

func TestGetEmailTemplate_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/email-templates/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	GetEmailTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetEmailTemplate() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateEmailTemplate_InvalidID(t *testing.T) {
	body := `{"subject": "test"}`
	req := httptest.NewRequest(http.MethodPut, "/email-templates/invalid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateEmailTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateEmailTemplate() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestUpdateEmailTemplate_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/email-templates/1", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateEmailTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateEmailTemplate() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestPreviewEmailTemplate_InvalidID(t *testing.T) {
	body := `{"variables": {}}`
	req := httptest.NewRequest(http.MethodPost, "/email-templates/invalid/preview", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	PreviewEmailTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("PreviewEmailTemplate() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestPreviewEmailTemplate_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/email-templates/1/preview", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	PreviewEmailTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("PreviewEmailTemplate() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// Note: GetChangelog tests require integration test setup with database
// since the handler relies on settingsService global variable.
// Input validation for pagination is simple (uses defaults for invalid values)
// and doesn't warrant unit tests without service mocking.
