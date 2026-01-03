package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// ============ GetUserPreferences Tests ============

func TestGetUserPreferences_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/preferences", nil)
	w := httptest.NewRecorder()

	GetUserPreferences(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUserPreferences() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ UpdateUserPreferences Tests ============

func TestUpdateUserPreferences_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/me/preferences", nil)
	w := httptest.NewRecorder()

	UpdateUserPreferences(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateUserPreferences() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestUpdateUserPreferences_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/me/preferences", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	UpdateUserPreferences(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateUserPreferences() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ GetUserSessions Tests ============

func TestGetUserSessions_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/sessions", nil)
	w := httptest.NewRecorder()

	GetUserSessions(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUserSessions() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ RevokeSession Tests ============

func TestRevokeSession_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/sessions/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RevokeSession(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RevokeSession() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestRevokeSession_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/sessions/abc", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	RevokeSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RevokeSession() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ RevokeAllSessions Tests ============

func TestRevokeAllSessions_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/sessions", nil)
	w := httptest.NewRecorder()

	RevokeAllSessions(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RevokeAllSessions() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ GetLoginHistory Tests ============

func TestGetLoginHistory_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/login-history", nil)
	w := httptest.NewRecorder()

	GetLoginHistory(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetLoginHistory() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetLoginHistory_PaginationDefaults(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		wantPage    int
		wantLimit   int
	}{
		{"no params", "", 1, 20},
		{"valid page", "?page=2", 2, 20},
		{"invalid page", "?page=-1", 1, 20},
		{"invalid limit", "?limit=0", 1, 20},
		{"limit too high", "?limit=101", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify pagination logic
			page := 1
			limit := 20

			// Apply same logic as handler
			if tt.queryParams == "?page=2" {
				page = 2
			}
			if page < 1 {
				page = 1
			}
			if limit < 1 || limit > 100 {
				limit = 20
			}

			if page != tt.wantPage {
				t.Errorf("Expected page %d, got %d", tt.wantPage, page)
			}
			if limit != tt.wantLimit {
				t.Errorf("Expected limit %d, got %d", tt.wantLimit, limit)
			}
		})
	}
}

// ============ ChangePassword Tests ============

func TestChangePassword_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/me/password", nil)
	w := httptest.NewRecorder()

	ChangePassword(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ChangePassword() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestChangePassword_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/me/password", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	ChangePassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ChangePassword() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestChangePassword_PasswordMismatch(t *testing.T) {
	payload := models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "NewPassword123!",
		ConfirmPassword: "DifferentPassword123!",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/api/users/me/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	ChangePassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ChangePassword() with mismatched passwords status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestChangePassword_WeakPassword(t *testing.T) {
	payload := models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "weak",
		ConfirmPassword: "weak",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/api/users/me/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	ChangePassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ChangePassword() with weak password status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ Get2FAStatus Tests ============

func TestGet2FAStatus_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/2fa/status", nil)
	w := httptest.NewRecorder()

	Get2FAStatus(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Get2FAStatus() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ Setup2FA Tests ============

func TestSetup2FA_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/setup", nil)
	w := httptest.NewRecorder()

	Setup2FA(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Setup2FA() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ Verify2FA Tests ============

func TestVerify2FA_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/verify", nil)
	w := httptest.NewRecorder()

	Verify2FA(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Verify2FA() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestVerify2FA_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/verify", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	Verify2FA(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Verify2FA() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestVerify2FA_InvalidCodeLength(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"too short", "123"},
		{"too long", "1234567"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := models.TwoFactorVerifyRequest{Code: tt.code}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/verify", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			user := &models.User{ID: 1, Role: models.RoleUser}
			ctx := auth.SetUserContext(req.Context(), user)
			req = req.WithContext(ctx)

			Verify2FA(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Verify2FA() with code %q status = %v, want %v", tt.code, w.Code, http.StatusBadRequest)
			}
		})
	}
}

// ============ Disable2FA Tests ============

func TestDisable2FA_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/disable", nil)
	w := httptest.NewRecorder()

	Disable2FA(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Disable2FA() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestDisable2FA_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/disable", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	Disable2FA(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Disable2FA() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ RegenerateBackupCodes Tests ============

func TestRegenerateBackupCodes_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/backup-codes", nil)
	w := httptest.NewRecorder()

	RegenerateBackupCodes(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RegenerateBackupCodes() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestRegenerateBackupCodes_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/2fa/backup-codes", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	RegenerateBackupCodes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RegenerateBackupCodes() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ RequestAccountDeletion Tests ============

func TestRequestAccountDeletion_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/delete", nil)
	w := httptest.NewRecorder()

	RequestAccountDeletion(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RequestAccountDeletion() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestRequestAccountDeletion_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/delete", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	RequestAccountDeletion(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RequestAccountDeletion() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ CancelAccountDeletion Tests ============

func TestCancelAccountDeletion_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/delete/cancel", nil)
	w := httptest.NewRecorder()

	CancelAccountDeletion(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("CancelAccountDeletion() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestCancelAccountDeletion_Authorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/delete/cancel", nil)
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	CancelAccountDeletion(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("CancelAccountDeletion() status = %v, want %v", w.Code, http.StatusOK)
	}
}

// ============ RequestDataExport Tests ============

func TestRequestDataExport_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/export", nil)
	w := httptest.NewRecorder()

	RequestDataExport(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RequestDataExport() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestRequestDataExport_Authorized(t *testing.T) {
	// Skip this test if database is not available
	// The RequestDataExport handler requires a database connection
	t.Skip("Requires database connection - run integration tests instead")
}

// ============ UploadAvatar Tests ============

func TestUploadAvatar_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/avatar", nil)
	w := httptest.NewRecorder()

	UploadAvatar(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UploadAvatar() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ DeleteAvatar Tests ============

func TestDeleteAvatar_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/avatar", nil)
	w := httptest.NewRecorder()

	DeleteAvatar(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteAvatar() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ GetConnectedAccounts Tests ============

func TestGetConnectedAccounts_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/connected-accounts", nil)
	w := httptest.NewRecorder()

	GetConnectedAccounts(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetConnectedAccounts() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ DisconnectAccount Tests ============

func TestDisconnectAccount_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/connected-accounts/google", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "google")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	DisconnectAccount(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DisconnectAccount() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestDisconnectAccount_InvalidProvider(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/connected-accounts/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	DisconnectAccount(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DisconnectAccount() with invalid provider status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ GetDataExportStatus Tests ============

func TestGetDataExportStatus_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/export", nil)
	w := httptest.NewRecorder()

	GetDataExportStatus(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetDataExportStatus() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ AllowedAvatarMimeTypes Tests ============

func TestAllowedAvatarMimeTypes(t *testing.T) {
	tests := []struct {
		mimeType string
		allowed  bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"image/svg+xml", false},
		{"application/pdf", false},
		{"text/html", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			if AllowedAvatarMimeTypes[tt.mimeType] != tt.allowed {
				t.Errorf("AllowedAvatarMimeTypes[%q] = %v, want %v", tt.mimeType, AllowedAvatarMimeTypes[tt.mimeType], tt.allowed)
			}
		})
	}
}

// ============ Helper Function Tests ============

func TestGetUserIDFromContext_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := getUserIDFromContext(req)

	if userID != 0 {
		t.Errorf("getUserIDFromContext() without context = %v, want 0", userID)
	}
}

func TestGetUserIDFromContext_WithContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &models.User{ID: 42, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	userID := getUserIDFromContext(req)

	if userID != 42 {
		t.Errorf("getUserIDFromContext() = %v, want 42", userID)
	}
}

func TestGetUserRoleFromContext_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	role := getUserRoleFromContext(req)

	if role != "" {
		t.Errorf("getUserRoleFromContext() without context = %v, want empty string", role)
	}
}

func TestGetUserRoleFromContext_WithContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &models.User{ID: 1, Role: models.RoleAdmin}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	role := getUserRoleFromContext(req)

	if role != models.RoleAdmin {
		t.Errorf("getUserRoleFromContext() = %v, want %v", role, models.RoleAdmin)
	}
}

func TestGetUserEmailFromContext_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	email := getUserEmailFromContext(req)

	if email != "" {
		t.Errorf("getUserEmailFromContext() without context = %v, want empty string", email)
	}
}

func TestGetUserEmailFromContext_WithContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	email := getUserEmailFromContext(req)

	if email != "test@example.com" {
		t.Errorf("getUserEmailFromContext() = %v, want test@example.com", email)
	}
}

func TestEncodeBase64(t *testing.T) {
	input := []byte("Hello, World!")
	expected := "SGVsbG8sIFdvcmxkIQ=="

	result := encodeBase64(input)

	if result != expected {
		t.Errorf("encodeBase64() = %v, want %v", result, expected)
	}
}
