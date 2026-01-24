package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/services"
	"react-golang-starter/internal/stripe"
)

// OrgHandler handles organization-related HTTP requests
type OrgHandler struct {
	orgService *services.OrgService
}

// NewOrgHandler creates a new organization handler
func NewOrgHandler(orgService *services.OrgService) *OrgHandler {
	return &OrgHandler{orgService: orgService}
}

// CreateOrganizationRequest represents the request body for creating an organization
type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}

// UpdateOrganizationRequest represents the request body for updating an organization
type UpdateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// InviteMemberRequest represents the request body for inviting a member
type InviteMemberRequest struct {
	Email string                  `json:"email" validate:"required,email"`
	Role  models.OrganizationRole `json:"role" validate:"required,oneof=admin member"`
}

// UpdateMemberRoleRequest represents the request body for updating a member's role
type UpdateMemberRoleRequest struct {
	Role models.OrganizationRole `json:"role" validate:"required,oneof=owner admin member"`
}

// OrganizationResponse represents an organization in API responses
type OrganizationResponse struct {
	ID        uint                    `json:"id"`
	Name      string                  `json:"name"`
	Slug      string                  `json:"slug"`
	Plan      models.OrganizationPlan `json:"plan"`
	CreatedAt string                  `json:"created_at"`
	Role      models.OrganizationRole `json:"role,omitempty"` // User's role in this org
}

// MemberResponse represents a member in API responses
type MemberResponse struct {
	ID       uint                    `json:"id"`
	UserID   uint                    `json:"user_id"`
	Email    string                  `json:"email"`
	Name     string                  `json:"name"`
	Role     models.OrganizationRole `json:"role"`
	Status   models.MemberStatus     `json:"status"`
	JoinedAt *string                 `json:"joined_at,omitempty"`
}

// InvitationResponse represents an invitation in API responses
type InvitationResponse struct {
	ID        uint                    `json:"id"`
	Email     string                  `json:"email"`
	Role      models.OrganizationRole `json:"role"`
	InvitedBy string                  `json:"invited_by"`
	ExpiresAt string                  `json:"expires_at"`
	CreatedAt string                  `json:"created_at"`
}

// ListOrganizations returns all organizations the user is a member of
// @Summary List user's organizations
// @Description Get all organizations the authenticated user belongs to
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse{data=[]OrganizationResponse}
// @Failure 401 {object} models.ErrorResponse
// @Router /organizations [get]
func (h *OrgHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok || user == nil {
		WriteUnauthorized(w, r, "Unauthorized")
		return
	}

	// Use single query to get orgs with roles (no N+1)
	orgsWithRoles, err := h.orgService.GetUserOrganizationsWithRoles(r.Context(), user.ID)
	if err != nil {
		WriteInternalError(w, r, "Failed to fetch organizations")
		return
	}

	// Build response
	response := make([]OrganizationResponse, 0, len(orgsWithRoles))
	for _, owr := range orgsWithRoles {
		response = append(response, OrganizationResponse{
			ID:        owr.Organization.ID,
			Name:      owr.Organization.Name,
			Slug:      owr.Organization.Slug,
			Plan:      owr.Organization.Plan,
			CreatedAt: owr.Organization.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Role:      owr.Role,
		})
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
}

// CreateOrganization creates a new organization
// @Summary Create organization
// @Description Create a new organization with the authenticated user as owner
// @Tags organizations
// @Accept json
// @Produce json
// @Param request body CreateOrganizationRequest true "Organization details"
// @Success 201 {object} models.SuccessResponse{data=OrganizationResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /organizations [post]
func (h *OrgHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok || user == nil {
		WriteUnauthorized(w, r, "Unauthorized")
		return
	}

	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	org, err := h.orgService.CreateOrganization(r.Context(), user.ID, req.Name, req.Slug)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidSlug):
			WriteBadRequest(w, r, "Invalid slug format. Use lowercase letters, numbers, and hyphens only.")
		case errors.Is(err, services.ErrOrgSlugTaken):
			WriteConflict(w, r, "Organization slug is already taken")
		default:
			WriteInternalError(w, r, "Failed to create organization")
		}
		return
	}

	response := OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Plan:      org.Plan,
		CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Role:      models.OrgRoleOwner,
	}

	WriteJSON(w, http.StatusCreated, models.SuccessResponse{Success: true, Data: response})
}

// GetOrganization returns organization details
// @Summary Get organization
// @Description Get organization details by slug
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse{data=OrganizationResponse}
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /organizations/{orgSlug} [get]
func (h *OrgHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())
	membership := auth.GetMembershipFromContext(r.Context())

	if org == nil || membership == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	response := OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Plan:      org.Plan,
		CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Role:      membership.Role,
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
}

// UpdateOrganization updates organization details
// @Summary Update organization
// @Description Update organization details (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Param request body UpdateOrganizationRequest true "Updated details"
// @Success 200 {object} models.SuccessResponse{data=OrganizationResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug} [put]
func (h *OrgHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())
	membership := auth.GetMembershipFromContext(r.Context())

	if org == nil || membership == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if err := h.orgService.UpdateOrganization(r.Context(), org, req.Name); err != nil {
		WriteInternalError(w, r, "Failed to update organization")
		return
	}

	response := OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Plan:      org.Plan,
		CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Role:      membership.Role,
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
}

// DeleteOrganization deletes an organization
// @Summary Delete organization
// @Description Delete an organization (owner only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug} [delete]
func (h *OrgHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	if err := h.orgService.DeleteOrganization(r.Context(), org); err != nil {
		WriteInternalError(w, r, "Failed to delete organization")
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]string{"message": "Organization deleted successfully"}})
}

// ListMembers returns all members of an organization
// @Summary List organization members
// @Description Get all members of an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse{data=[]MemberResponse}
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/members [get]
func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	members, err := h.orgService.GetMembers(r.Context(), org.ID)
	if err != nil {
		WriteInternalError(w, r, "Failed to fetch members")
		return
	}

	response := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		var joinedAt *string
		if m.AcceptedAt != nil {
			t := m.AcceptedAt.Format("2006-01-02T15:04:05Z")
			joinedAt = &t
		}

		email := ""
		name := ""
		if m.User != nil {
			email = m.User.Email
			name = m.User.Name
		}

		response = append(response, MemberResponse{
			ID:       m.ID,
			UserID:   m.UserID,
			Email:    email,
			Name:     name,
			Role:     m.Role,
			Status:   m.Status,
			JoinedAt: joinedAt,
		})
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
}

// InviteMember invites a new member to the organization
// @Summary Invite member
// @Description Send an invitation to join the organization (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Param request body InviteMemberRequest true "Invitation details"
// @Success 201 {object} models.SuccessResponse{data=InvitationResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/members/invite [post]
func (h *OrgHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	org := auth.GetOrganizationFromContext(r.Context())

	if !ok || org == nil || user == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Cannot invite as owner
	if req.Role == models.OrgRoleOwner {
		WriteBadRequest(w, r, "Cannot invite as owner. Use role transfer instead.")
		return
	}

	invitation, err := h.orgService.CreateInvitation(r.Context(), org.ID, user.ID, req.Email, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrAlreadyMember):
			WriteConflict(w, r, "User is already a member")
		case errors.Is(err, services.ErrInvitationEmailTaken):
			WriteConflict(w, r, "An invitation for this email already exists")
		default:
			WriteInternalError(w, r, "Failed to create invitation")
		}
		return
	}

	response := InvitationResponse{
		ID:        invitation.ID,
		Email:     invitation.Email,
		Role:      invitation.Role,
		InvitedBy: user.Email,
		ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt: invitation.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	WriteJSON(w, http.StatusCreated, models.SuccessResponse{Success: true, Data: response})
}

// UpdateMemberRole updates a member's role
// @Summary Update member role
// @Description Update a member's role in the organization (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Param userId path int true "User ID"
// @Param request body UpdateMemberRoleRequest true "New role"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/members/{userId}/role [put]
func (h *OrgHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	org := auth.GetOrganizationFromContext(r.Context())
	membership := auth.GetMembershipFromContext(r.Context())

	if !ok || org == nil || user == nil || membership == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	userIDStr := r.PathValue("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid user ID")
		return
	}

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Only owners can promote to owner
	if req.Role == models.OrgRoleOwner && membership.Role != models.OrgRoleOwner {
		WriteForbidden(w, r, "Only owners can promote to owner")
		return
	}

	err = h.orgService.UpdateMemberRole(r.Context(), org.ID, uint(targetUserID), user.ID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotMember):
			WriteNotFound(w, r, "Member not found")
		case errors.Is(err, services.ErrCannotChangeOwnRole):
			WriteBadRequest(w, r, "Cannot change your own role")
		case errors.Is(err, services.ErrMustHaveOwner):
			WriteBadRequest(w, r, "Organization must have at least one owner")
		default:
			WriteInternalError(w, r, "Failed to update role")
		}
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]string{"message": "Role updated successfully"}})
}

// RemoveMember removes a member from the organization
// @Summary Remove member
// @Description Remove a member from the organization (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Param userId path int true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/members/{userId} [delete]
func (h *OrgHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	userIDStr := r.PathValue("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid user ID")
		return
	}

	err = h.orgService.RemoveMember(r.Context(), org.ID, uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotMember):
			WriteNotFound(w, r, "Member not found")
		case errors.Is(err, services.ErrCannotRemoveOwner):
			WriteBadRequest(w, r, "Cannot remove the only owner")
		default:
			WriteInternalError(w, r, "Failed to remove member")
		}
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]string{"message": "Member removed successfully"}})
}

// ListInvitations returns pending invitations for an organization
// @Summary List pending invitations
// @Description Get all pending invitations for an organization (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse{data=[]InvitationResponse}
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/invitations [get]
func (h *OrgHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	invitations, err := h.orgService.GetPendingInvitations(r.Context(), org.ID)
	if err != nil {
		WriteInternalError(w, r, "Failed to fetch invitations")
		return
	}

	response := make([]InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		invitedBy := ""
		if inv.InvitedByUser != nil {
			invitedBy = inv.InvitedByUser.Email
		}
		response = append(response, InvitationResponse{
			ID:        inv.ID,
			Email:     inv.Email,
			Role:      inv.Role,
			InvitedBy: invitedBy,
			ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
}

// CancelInvitation cancels a pending invitation
// @Summary Cancel invitation
// @Description Cancel a pending invitation (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Param invitationId path int true "Invitation ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/invitations/{invitationId} [delete]
func (h *OrgHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	invIDStr := r.PathValue("invitationId")
	invID, err := strconv.ParseUint(invIDStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid invitation ID")
		return
	}

	err = h.orgService.CancelInvitation(r.Context(), uint(invID), org.ID)
	if err != nil {
		if errors.Is(err, services.ErrInvitationNotFound) {
			WriteNotFound(w, r, "Invitation not found")
			return
		}
		WriteInternalError(w, r, "Failed to cancel invitation")
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]string{"message": "Invitation cancelled successfully"}})
}

// AcceptInvitation accepts an invitation to join an organization
// @Summary Accept invitation
// @Description Accept an invitation to join an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param token query string true "Invitation token"
// @Success 200 {object} models.SuccessResponse{data=OrganizationResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /invitations/accept [post]
func (h *OrgHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok || user == nil {
		WriteUnauthorized(w, r, "Unauthorized")
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		WriteBadRequest(w, r, "Invitation token required")
		return
	}

	member, err := h.orgService.AcceptInvitation(r.Context(), token, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvitationNotFound):
			WriteNotFound(w, r, "Invitation not found")
		case errors.Is(err, services.ErrInvitationExpired):
			WriteBadRequest(w, r, "Invitation has expired")
		case errors.Is(err, services.ErrInvitationAccepted):
			WriteBadRequest(w, r, "Invitation has already been accepted")
		case errors.Is(err, services.ErrAlreadyMember):
			WriteConflict(w, r, "You are already a member of this organization")
		default:
			WriteInternalError(w, r, "Failed to accept invitation")
		}
		return
	}

	// Fetch the organization details using the member's organization ID
	org, err := h.orgService.GetOrganizationByID(r.Context(), member.OrganizationID)
	if err == nil && org != nil {
		response := OrganizationResponse{
			ID:        org.ID,
			Name:      org.Name,
			Slug:      org.Slug,
			Plan:      org.Plan,
			CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Role:      member.Role,
		}
		WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]string{"message": "Invitation accepted successfully"}})
}

// LeaveOrganization allows a user to leave an organization
// @Summary Leave organization
// @Description Leave an organization (non-owners only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/leave [post]
func (h *OrgHandler) LeaveOrganization(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	org := auth.GetOrganizationFromContext(r.Context())
	membership := auth.GetMembershipFromContext(r.Context())

	if !ok || org == nil || user == nil || membership == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	if membership.Role == models.OrgRoleOwner {
		WriteBadRequest(w, r, "Owners cannot leave. Transfer ownership first.")
		return
	}

	err := h.orgService.RemoveMember(r.Context(), org.ID, user.ID)
	if err != nil {
		WriteInternalError(w, r, "Failed to leave organization")
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]string{"message": "Left organization successfully"}})
}

// OrgBillingResponse represents organization billing information
type OrgBillingResponse struct {
	Plan             models.OrganizationPlan      `json:"plan"`
	HasSubscription  bool                         `json:"has_subscription"`
	Subscription     *models.SubscriptionResponse `json:"subscription,omitempty"`
	SeatLimit        int                          `json:"seat_limit"`
	SeatCount        int64                        `json:"seat_count"`
	StripeCustomerID *string                      `json:"stripe_customer_id,omitempty"`
}

// OrgCheckoutRequest represents the request body for org checkout
type OrgCheckoutRequest struct {
	PriceID string `json:"price_id" validate:"required"`
}

// GetOrganizationBilling returns the organization's billing information
// @Summary Get organization billing
// @Description Get organization billing and subscription details (admin+ only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse{data=OrgBillingResponse}
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/billing [get]
func (h *OrgHandler) GetOrganizationBilling(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	// Get member count
	memberCount, err := h.orgService.GetMemberCount(r.Context(), org.ID)
	if err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to get member count")
		memberCount = 0
	}

	response := OrgBillingResponse{
		Plan:             org.Plan,
		HasSubscription:  org.HasSubscription(),
		SeatLimit:        org.GetSeatLimit(),
		SeatCount:        memberCount,
		StripeCustomerID: org.StripeCustomerID,
	}

	// Get subscription if exists
	sub, err := h.orgService.GetOrganizationSubscription(r.Context(), org.ID)
	if err == nil && sub != nil {
		subResponse := sub.ToSubscriptionResponse()
		response.Subscription = &subResponse
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: response})
}

// CreateOrganizationCheckout creates a Stripe checkout session for the organization
// @Summary Create organization checkout session
// @Description Create a Stripe checkout session for organization subscription (owner only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Param request body OrgCheckoutRequest true "Checkout request"
// @Success 200 {object} models.SuccessResponse{data=models.CheckoutSessionResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/billing/checkout [post]
func (h *OrgHandler) CreateOrganizationCheckout(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	org := auth.GetOrganizationFromContext(r.Context())

	if !ok || org == nil || user == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	svc := stripe.GetService()
	if !svc.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Billing is not configured")
		return
	}

	var req OrgCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	if req.PriceID == "" {
		WriteBadRequest(w, r, "Price ID is required")
		return
	}

	// Get or create Stripe customer for org
	var customerID string
	if org.StripeCustomerID != nil && *org.StripeCustomerID != "" {
		customerID = *org.StripeCustomerID
	} else {
		// Create customer using the org owner's info but with org metadata
		newCustomerID, err := svc.CreateCustomer(r.Context(), user)
		if err != nil {
			log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to create stripe customer for org")
			WriteInternalError(w, r, "Failed to create billing account")
			return
		}
		customerID = newCustomerID

		// Save customer ID to organization
		if err := h.orgService.SetOrganizationStripeCustomer(r.Context(), org.ID, customerID); err != nil {
			log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to save stripe customer ID")
		}
	}

	// Get config for URLs
	stripeConfig := stripe.LoadConfig()

	// Create checkout session
	session, err := svc.CreateCheckoutSession(r.Context(), customerID, req.PriceID, stripeConfig.SuccessURL, stripeConfig.CancelURL)
	if err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to create checkout session")
		WriteInternalError(w, r, "Failed to create checkout session")
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: models.CheckoutSessionResponse{
		SessionID: session.ID,
		URL:       session.URL,
	}})
}

// CreateOrganizationBillingPortal creates a Stripe billing portal session
// @Summary Create organization billing portal
// @Description Create a Stripe billing portal session for subscription management (owner only)
// @Tags organizations
// @Accept json
// @Produce json
// @Param orgSlug path string true "Organization slug"
// @Success 200 {object} models.SuccessResponse{data=models.PortalSessionResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /organizations/{orgSlug}/billing/portal [post]
func (h *OrgHandler) CreateOrganizationBillingPortal(w http.ResponseWriter, r *http.Request) {
	org := auth.GetOrganizationFromContext(r.Context())

	if org == nil {
		WriteNotFound(w, r, "Organization not found")
		return
	}

	svc := stripe.GetService()
	if !svc.IsAvailable() {
		WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Billing is not configured")
		return
	}

	if org.StripeCustomerID == nil || *org.StripeCustomerID == "" {
		WriteBadRequest(w, r, "No billing account found for this organization")
		return
	}

	// Get config for return URL
	stripeConfig := stripe.LoadConfig()

	// Create portal session
	session, err := svc.CreatePortalSession(r.Context(), *org.StripeCustomerID, stripeConfig.PortalReturnURL)
	if err != nil {
		log.Error().Err(err).Uint("org_id", org.ID).Msg("failed to create portal session")
		WriteInternalError(w, r, "Failed to create billing portal")
		return
	}

	WriteJSON(w, http.StatusOK, models.SuccessResponse{Success: true, Data: models.PortalSessionResponse{
		URL: session.URL,
	}})
}
