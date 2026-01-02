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
)

// Helper functions to set context values for testing
func setUserInTestContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, auth.UserContextKey, user)
}

func setOrganizationInTestContext(ctx context.Context, org *models.Organization) context.Context {
	return context.WithValue(ctx, auth.OrganizationContextKey, org)
}

func setMembershipInTestContext(ctx context.Context, membership *models.OrganizationMember) context.Context {
	return context.WithValue(ctx, auth.MembershipContextKey, membership)
}

// ============ OrgHandler Creation Tests ============

func TestNewOrgHandler(t *testing.T) {
	handler := NewOrgHandler(nil)
	if handler == nil {
		t.Error("NewOrgHandler() returned nil")
	}
}

// ============ ListOrganizations Tests ============

func TestOrgHandler_ListOrganizations_Unauthorized(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	w := httptest.NewRecorder()

	handler.ListOrganizations(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ListOrganizations() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ CreateOrganization Tests ============

func TestOrgHandler_CreateOrganization_Unauthorized(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/organizations", nil)
	w := httptest.NewRecorder()

	handler.CreateOrganization(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("CreateOrganization() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestOrgHandler_CreateOrganization_InvalidJSON(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add user to context
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleUser}
	ctx := setUserInTestContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.CreateOrganization(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateOrganization() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ GetOrganization Tests ============

func TestOrgHandler_GetOrganization_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/test-org", nil)
	w := httptest.NewRecorder()

	// No org in context means not found
	handler.GetOrganization(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetOrganization() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ UpdateOrganization Tests ============

func TestOrgHandler_UpdateOrganization_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	payload := UpdateOrganizationRequest{Name: "New Name"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/organizations/test-org", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateOrganization(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateOrganization() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestOrgHandler_UpdateOrganization_InvalidJSON(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPut, "/organizations/test-org", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add org and membership to context
	org := &models.Organization{ID: 1, Name: "Test Org", Slug: "test-org"}
	membership := &models.OrganizationMember{Role: models.OrgRoleOwner}
	ctx := setOrganizationInTestContext(req.Context(), org)
	ctx = setMembershipInTestContext(ctx, membership)
	req = req.WithContext(ctx)

	handler.UpdateOrganization(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateOrganization() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ DeleteOrganization Tests ============

func TestOrgHandler_DeleteOrganization_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/test-org", nil)
	w := httptest.NewRecorder()

	handler.DeleteOrganization(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("DeleteOrganization() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ ListMembers Tests ============

func TestOrgHandler_ListMembers_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/test-org/members", nil)
	w := httptest.NewRecorder()

	handler.ListMembers(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("ListMembers() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ InviteMember Tests ============

func TestOrgHandler_InviteMember_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/organizations/test-org/members/invite", nil)
	w := httptest.NewRecorder()

	handler.InviteMember(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("InviteMember() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestOrgHandler_InviteMember_InvalidJSON(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/organizations/test-org/members/invite", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add user and org to context
	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleUser}
	org := &models.Organization{ID: 1, Name: "Test Org", Slug: "test-org"}
	ctx := setUserInTestContext(req.Context(), user)
	ctx = setOrganizationInTestContext(ctx, org)
	req = req.WithContext(ctx)

	handler.InviteMember(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("InviteMember() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestOrgHandler_InviteMember_CannotInviteAsOwner(t *testing.T) {
	handler := NewOrgHandler(nil)

	payload := InviteMemberRequest{
		Email: "new@example.com",
		Role:  models.OrgRoleOwner,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/organizations/test-org/members/invite", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add user and org to context
	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleUser}
	org := &models.Organization{ID: 1, Name: "Test Org", Slug: "test-org"}
	ctx := setUserInTestContext(req.Context(), user)
	ctx = setOrganizationInTestContext(ctx, org)
	req = req.WithContext(ctx)

	handler.InviteMember(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("InviteMember() as owner status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ UpdateMemberRole Tests ============

func TestOrgHandler_UpdateMemberRole_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPut, "/organizations/test-org/members/1/role", nil)
	w := httptest.NewRecorder()

	handler.UpdateMemberRole(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateMemberRole() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ RemoveMember Tests ============

func TestOrgHandler_RemoveMember_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/test-org/members/1", nil)
	w := httptest.NewRecorder()

	handler.RemoveMember(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("RemoveMember() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ ListInvitations Tests ============

func TestOrgHandler_ListInvitations_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/test-org/invitations", nil)
	w := httptest.NewRecorder()

	handler.ListInvitations(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("ListInvitations() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ CancelInvitation Tests ============

func TestOrgHandler_CancelInvitation_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/test-org/invitations/1", nil)
	w := httptest.NewRecorder()

	handler.CancelInvitation(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("CancelInvitation() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

// ============ AcceptInvitation Tests ============

func TestOrgHandler_AcceptInvitation_Unauthorized(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/invitations/accept?token=abc123", nil)
	w := httptest.NewRecorder()

	handler.AcceptInvitation(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AcceptInvitation() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestOrgHandler_AcceptInvitation_MissingToken(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/invitations/accept", nil)
	w := httptest.NewRecorder()

	// Add user to context
	user := &models.User{ID: 1, Email: "test@example.com", Role: models.RoleUser}
	ctx := setUserInTestContext(req.Context(), user)
	req = req.WithContext(ctx)

	handler.AcceptInvitation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("AcceptInvitation() without token status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ LeaveOrganization Tests ============

func TestOrgHandler_LeaveOrganization_NotFound(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/organizations/test-org/leave", nil)
	w := httptest.NewRecorder()

	handler.LeaveOrganization(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("LeaveOrganization() without org in context status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestOrgHandler_LeaveOrganization_OwnerCannotLeave(t *testing.T) {
	handler := NewOrgHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/organizations/test-org/leave", nil)
	w := httptest.NewRecorder()

	// Add user, org, and membership as owner
	user := &models.User{ID: 1, Email: "owner@example.com", Role: models.RoleUser}
	org := &models.Organization{ID: 1, Name: "Test Org", Slug: "test-org"}
	membership := &models.OrganizationMember{Role: models.OrgRoleOwner}

	ctx := setUserInTestContext(req.Context(), user)
	ctx = setOrganizationInTestContext(ctx, org)
	ctx = setMembershipInTestContext(ctx, membership)
	req = req.WithContext(ctx)

	handler.LeaveOrganization(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("LeaveOrganization() as owner status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ Request/Response Types Tests ============

func TestCreateOrganizationRequest_JSONMarshal(t *testing.T) {
	req := CreateOrganizationRequest{
		Name: "Test Org",
		Slug: "test-org",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal CreateOrganizationRequest: %v", err)
	}

	var decoded CreateOrganizationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CreateOrganizationRequest: %v", err)
	}

	if decoded.Name != req.Name {
		t.Errorf("CreateOrganizationRequest.Name = %v, want %v", decoded.Name, req.Name)
	}
	if decoded.Slug != req.Slug {
		t.Errorf("CreateOrganizationRequest.Slug = %v, want %v", decoded.Slug, req.Slug)
	}
}

func TestUpdateOrganizationRequest_JSONMarshal(t *testing.T) {
	req := UpdateOrganizationRequest{Name: "Updated Name"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal UpdateOrganizationRequest: %v", err)
	}

	var decoded UpdateOrganizationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal UpdateOrganizationRequest: %v", err)
	}

	if decoded.Name != req.Name {
		t.Errorf("UpdateOrganizationRequest.Name = %v, want %v", decoded.Name, req.Name)
	}
}

func TestInviteMemberRequest_JSONMarshal(t *testing.T) {
	req := InviteMemberRequest{
		Email: "new@example.com",
		Role:  models.OrgRoleAdmin,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal InviteMemberRequest: %v", err)
	}

	var decoded InviteMemberRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InviteMemberRequest: %v", err)
	}

	if decoded.Email != req.Email {
		t.Errorf("InviteMemberRequest.Email = %v, want %v", decoded.Email, req.Email)
	}
	if decoded.Role != req.Role {
		t.Errorf("InviteMemberRequest.Role = %v, want %v", decoded.Role, req.Role)
	}
}

func TestUpdateMemberRoleRequest_JSONMarshal(t *testing.T) {
	req := UpdateMemberRoleRequest{Role: models.OrgRoleMember}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal UpdateMemberRoleRequest: %v", err)
	}

	var decoded UpdateMemberRoleRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal UpdateMemberRoleRequest: %v", err)
	}

	if decoded.Role != req.Role {
		t.Errorf("UpdateMemberRoleRequest.Role = %v, want %v", decoded.Role, req.Role)
	}
}

// ============ Helper Functions Tests ============

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	respondWithError(w, r, http.StatusBadRequest, "Test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("respondWithError() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var response models.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != "Test error" {
		t.Errorf("respondWithError() message = %v, want %v", response.Message, "Test error")
	}
}

func TestRespondWithSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	respondWithSuccess(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("respondWithSuccess() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response models.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("respondWithSuccess() Success = false, want true")
	}
}

// ============ OrganizationResponse Tests ============

func TestOrganizationResponse_JSONMarshal(t *testing.T) {
	resp := OrganizationResponse{
		ID:        1,
		Name:      "Test Org",
		Slug:      "test-org",
		Plan:      models.OrgPlanFree,
		CreatedAt: "2024-01-01T00:00:00Z",
		Role:      models.OrgRoleOwner,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal OrganizationResponse: %v", err)
	}

	var decoded OrganizationResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal OrganizationResponse: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("OrganizationResponse.ID = %v, want %v", decoded.ID, resp.ID)
	}
	if decoded.Name != resp.Name {
		t.Errorf("OrganizationResponse.Name = %v, want %v", decoded.Name, resp.Name)
	}
	if decoded.Slug != resp.Slug {
		t.Errorf("OrganizationResponse.Slug = %v, want %v", decoded.Slug, resp.Slug)
	}
}

// ============ MemberResponse Tests ============

func TestMemberResponse_JSONMarshal(t *testing.T) {
	joinedAt := "2024-01-01T00:00:00Z"
	resp := MemberResponse{
		ID:       1,
		UserID:   1,
		Email:    "member@example.com",
		Name:     "Test Member",
		Role:     models.OrgRoleMember,
		Status:   models.MemberStatusActive,
		JoinedAt: &joinedAt,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal MemberResponse: %v", err)
	}

	var decoded MemberResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal MemberResponse: %v", err)
	}

	if decoded.Email != resp.Email {
		t.Errorf("MemberResponse.Email = %v, want %v", decoded.Email, resp.Email)
	}
	if decoded.Role != resp.Role {
		t.Errorf("MemberResponse.Role = %v, want %v", decoded.Role, resp.Role)
	}
}

// ============ InvitationResponse Tests ============

func TestInvitationResponse_JSONMarshal(t *testing.T) {
	resp := InvitationResponse{
		ID:        1,
		Email:     "invited@example.com",
		Role:      models.OrgRoleMember,
		InvitedBy: "admin@example.com",
		ExpiresAt: "2024-02-01T00:00:00Z",
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal InvitationResponse: %v", err)
	}

	var decoded InvitationResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InvitationResponse: %v", err)
	}

	if decoded.Email != resp.Email {
		t.Errorf("InvitationResponse.Email = %v, want %v", decoded.Email, resp.Email)
	}
	if decoded.InvitedBy != resp.InvitedBy {
		t.Errorf("InvitationResponse.InvitedBy = %v, want %v", decoded.InvitedBy, resp.InvitedBy)
	}
}
