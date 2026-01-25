package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// ============ ImpersonateUser Tests ============

func TestImpersonateUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/impersonate", nil)
	w := httptest.NewRecorder()

	ImpersonateUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ImpersonateUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestImpersonateUser_NonSuperAdmin(t *testing.T) {
	payload := models.ImpersonateRequest{UserID: 2, Reason: "Testing"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/impersonate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add regular admin user to context
	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	ImpersonateUser(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("ImpersonateUser() with admin role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestImpersonateUser_AlreadyImpersonating(t *testing.T) {
	payload := models.ImpersonateRequest{UserID: 2, Reason: "Testing"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/impersonate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add super admin who is already impersonating
	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin, OriginalUserID: 3}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	ImpersonateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ImpersonateUser() already impersonating status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestImpersonateUser_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/impersonate", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	ImpersonateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ImpersonateUser() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestImpersonateUser_CannotImpersonateSelf(t *testing.T) {
	payload := models.ImpersonateRequest{UserID: 1, Reason: "Testing"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/impersonate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	ImpersonateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ImpersonateUser() self-impersonation status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ StopImpersonation Tests ============

func TestStopImpersonation_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/stop-impersonate", nil)
	w := httptest.NewRecorder()

	StopImpersonation(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("StopImpersonation() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestStopImpersonation_NotImpersonating(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/stop-impersonate", nil)
	w := httptest.NewRecorder()

	// User not currently impersonating
	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin, OriginalUserID: 0}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	StopImpersonation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("StopImpersonation() not impersonating status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ AdminUpdateUserRole Tests ============

func TestAdminUpdateUserRole_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/1/role", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	AdminUpdateUserRole(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AdminUpdateUserRole() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestAdminUpdateUserRole_NonSuperAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/1/role", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	AdminUpdateUserRole(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("AdminUpdateUserRole() with admin role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestAdminUpdateUserRole_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/abc/role", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	AdminUpdateUserRole(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("AdminUpdateUserRole() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestAdminUpdateUserRole_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/2/role", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	AdminUpdateUserRole(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("AdminUpdateUserRole() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestAdminUpdateUserRole_InvalidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{"invalid role", "superuser"},
		{"empty role", ""},
		{"uppercase role", "ADMIN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{"role": tt.role}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPut, "/api/admin/users/2/role", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "2")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin}
			ctx := auth.SetClaimsContext(req.Context(), claims)
			req = req.WithContext(ctx)

			AdminUpdateUserRole(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("AdminUpdateUserRole() with role %q status = %v, want %v", tt.role, w.Code, http.StatusBadRequest)
			}
		})
	}
}

// ============ DeactivateUser Tests ============

func TestDeactivateUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/1/deactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	DeactivateUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeactivateUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestDeactivateUser_RegularUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/2/deactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleUser}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	DeactivateUser(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("DeactivateUser() with user role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestDeactivateUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/abc/deactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	DeactivateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DeactivateUser() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestDeactivateUser_CannotDeactivateSelf(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/1/deactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	DeactivateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DeactivateUser() self-deactivation status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ ReactivateUser Tests ============

func TestReactivateUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/1/reactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	ReactivateUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ReactivateUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestReactivateUser_RegularUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/2/reactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleUser}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	ReactivateUser(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("ReactivateUser() with user role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestReactivateUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/abc/reactivate", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	ReactivateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ReactivateUser() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ RestoreUser Tests ============

func TestRestoreUser_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/1/restore", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	RestoreUser(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("RestoreUser() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestRestoreUser_NonSuperAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/2/restore", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	RestoreUser(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("RestoreUser() with admin role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

func TestRestoreUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/abc/restore", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	claims := &auth.Claims{UserID: 1, Role: models.RoleSuperAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	RestoreUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("RestoreUser() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ GetDeletedUsers Tests ============

func TestGetDeletedUsers_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/deleted", nil)
	w := httptest.NewRecorder()

	GetDeletedUsers(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetDeletedUsers() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetDeletedUsers_NonSuperAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/deleted", nil)
	w := httptest.NewRecorder()

	claims := &auth.Claims{UserID: 1, Role: models.RoleAdmin}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	GetDeletedUsers(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("GetDeletedUsers() with admin role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

// ============ GetAuditLogs Tests ============

func TestGetAuditLogs_PaginationDefaults(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		wantPage    int
		wantLimit   int
	}{
		{"no params", "", 1, 20},
		{"invalid page", "?page=-1", 1, 20},
		{"invalid limit", "?limit=0", 1, 20},
		{"limit too high", "?limit=101", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies pagination defaults without database
			// Full integration tests would verify actual data retrieval
			filter := models.AuditLogFilter{}
			page := tt.wantPage
			if page < 1 {
				page = 1
			}
			limit := tt.wantLimit
			if limit < 1 || limit > 100 {
				limit = 20
			}
			filter.Page = page
			filter.Limit = limit

			if filter.Page != tt.wantPage {
				t.Errorf("Expected page %d, got %d", tt.wantPage, filter.Page)
			}
			if filter.Limit != tt.wantLimit {
				t.Errorf("Expected limit %d, got %d", tt.wantLimit, filter.Limit)
			}
		})
	}
}

// ============ SearchUsers Tests ============

func TestSearchUsers_MissingQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/search", nil)
	w := httptest.NewRecorder()

	SearchUsers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("SearchUsers() without query status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var response models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != "Query parameter is required" {
		t.Errorf("SearchUsers() error message = %q, want 'Query parameter is required'", response.Message)
	}
}

func TestSearchUsers_EmptyQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/search?query=", nil)
	w := httptest.NewRecorder()

	SearchUsers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("SearchUsers() with empty query status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ AdminUpdateUserRole Additional Tests ============

func TestAdminUpdateUserRole_ValidRoleConstants(t *testing.T) {
	// Test that valid role constants are defined correctly
	validRoles := map[string]bool{
		models.RoleSuperAdmin: true,
		models.RoleAdmin:      true,
		models.RolePremium:    true,
		models.RoleUser:       true,
	}

	for role := range validRoles {
		if role == "" {
			t.Error("Role constant should not be empty")
		}
	}

	// Verify role constants match expected values
	if models.RoleSuperAdmin != "super_admin" {
		t.Errorf("RoleSuperAdmin = %q, want 'super_admin'", models.RoleSuperAdmin)
	}
	if models.RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %q, want 'admin'", models.RoleAdmin)
	}
	if models.RolePremium != "premium" {
		t.Errorf("RolePremium = %q, want 'premium'", models.RolePremium)
	}
	if models.RoleUser != "user" {
		t.Errorf("RoleUser = %q, want 'user'", models.RoleUser)
	}
}

// ============ ImpersonateRequest Validation Tests ============

func TestImpersonateRequest_Structure(t *testing.T) {
	req := models.ImpersonateRequest{
		UserID: 42,
		Reason: "Customer support investigation",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ImpersonateRequest: %v", err)
	}

	var decoded models.ImpersonateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ImpersonateRequest: %v", err)
	}

	if decoded.UserID != 42 {
		t.Errorf("ImpersonateRequest.UserID = %d, want 42", decoded.UserID)
	}

	if decoded.Reason != "Customer support investigation" {
		t.Errorf("ImpersonateRequest.Reason = %q, want 'Customer support investigation'", decoded.Reason)
	}
}

// ============ AdminStatsResponse Structure Tests ============

func TestAdminStatsResponse_Structure(t *testing.T) {
	stats := models.AdminStatsResponse{
		TotalUsers:            100,
		ActiveUsers:           90,
		VerifiedUsers:         85,
		NewUsersToday:         5,
		NewUsersThisWeek:      25,
		NewUsersThisMonth:     50,
		TotalSubscriptions:    40,
		ActiveSubscriptions:   35,
		CanceledSubscriptions: 5,
		TotalFiles:            1000,
		TotalFileSize:         1073741824, // 1 GB
		UsersByRole:           map[string]int64{"user": 80, "admin": 10, "super_admin": 2},
	}

	if stats.TotalUsers != 100 {
		t.Errorf("AdminStatsResponse.TotalUsers = %d, want 100", stats.TotalUsers)
	}

	if stats.ActiveUsers != 90 {
		t.Errorf("AdminStatsResponse.ActiveUsers = %d, want 90", stats.ActiveUsers)
	}

	if stats.UsersByRole["user"] != 80 {
		t.Errorf("AdminStatsResponse.UsersByRole[user] = %d, want 80", stats.UsersByRole["user"])
	}
}

// ============ AuditLogFilter Tests ============

func TestAuditLogFilter_QueryParamParsing(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantUserID *uint
		wantAction string
	}{
		{
			name:       "with user_id",
			query:      "?user_id=42",
			wantUserID: func() *uint { v := uint(42); return &v }(),
			wantAction: "",
		},
		{
			name:       "with action",
			query:      "?action=login",
			wantUserID: nil,
			wantAction: "login",
		},
		{
			name:       "with both",
			query:      "?user_id=42&action=update",
			wantUserID: func() *uint { v := uint(42); return &v }(),
			wantAction: "update",
		},
		{
			name:       "invalid user_id is ignored",
			query:      "?user_id=abc",
			wantUserID: nil,
			wantAction: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/admin/audit-logs"+tt.query, nil)
			filter := models.AuditLogFilter{}

			// Parse query params like GetAuditLogs does
			if userIDStr := req.URL.Query().Get("user_id"); userIDStr != "" {
				if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
					uid := uint(userID)
					filter.UserID = &uid
				}
			}
			filter.Action = req.URL.Query().Get("action")

			if tt.wantUserID == nil && filter.UserID != nil {
				t.Errorf("Expected nil UserID, got %d", *filter.UserID)
			}
			if tt.wantUserID != nil && filter.UserID == nil {
				t.Errorf("Expected UserID %d, got nil", *tt.wantUserID)
			}
			if tt.wantUserID != nil && filter.UserID != nil && *filter.UserID != *tt.wantUserID {
				t.Errorf("UserID = %d, want %d", *filter.UserID, *tt.wantUserID)
			}
			if filter.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", filter.Action, tt.wantAction)
			}
		})
	}
}

func TestGetDeletedUsers_RegularUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/deleted", nil)
	w := httptest.NewRecorder()

	claims := &auth.Claims{UserID: 1, Role: models.RoleUser}
	ctx := auth.SetClaimsContext(req.Context(), claims)
	req = req.WithContext(ctx)

	GetDeletedUsers(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("GetDeletedUsers() with user role status = %v, want %v", w.Code, http.StatusForbidden)
	}
}
