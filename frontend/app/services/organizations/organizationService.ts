import { API_BASE_URL, authenticatedFetch, parseErrorResponse } from "../api/client";

// Types
export interface Organization {
  id: number;
  name: string;
  slug: string;
  plan: "free" | "pro" | "enterprise";
  created_at: string;
  role: "owner" | "admin" | "member";
}

export interface OrganizationMember {
  id: number;
  user_id: number;
  email: string;
  name: string;
  role: "owner" | "admin" | "member";
  status: "active" | "inactive" | "pending";
  joined_at?: string;
}

export interface OrganizationInvitation {
  id: number;
  email: string;
  role: "admin" | "member";
  invited_by: string;
  expires_at: string;
  created_at: string;
}

export interface CreateOrganizationRequest {
  name: string;
  slug: string;
}

export interface UpdateOrganizationRequest {
  name: string;
}

export interface InviteMemberRequest {
  email: string;
  role: "admin" | "member";
}

export interface UpdateMemberRoleRequest {
  role: "owner" | "admin" | "member";
}

export interface OrgBillingInfo {
  plan: "free" | "pro" | "enterprise";
  has_subscription: boolean;
  subscription?: {
    id: number;
    status: string;
    plan: string;
    current_period_start: string;
    current_period_end: string;
    cancel_at_period_end: boolean;
  };
  seat_limit: number;
  seat_count: number;
  stripe_customer_id?: string;
}

export interface CheckoutSessionResponse {
  session_id: string;
  url: string;
}

export interface PortalSessionResponse {
  url: string;
}

// API response types
interface SuccessResponse<T> {
  success: boolean;
  data: T;
}

export class OrganizationService {
  /**
   * List all organizations the current user belongs to
   */
  static async listOrganizations(): Promise<Organization[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to fetch organizations");
    }

    const data: SuccessResponse<Organization[]> = await response.json();
    return data.data || [];
  }

  /**
   * Create a new organization
   */
  static async createOrganization(req: CreateOrganizationRequest): Promise<Organization> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to create organization");
    }

    const data: SuccessResponse<Organization> = await response.json();
    return data.data;
  }

  /**
   * Get organization details by slug
   */
  static async getOrganization(slug: string): Promise<Organization> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to fetch organization");
    }

    const data: SuccessResponse<Organization> = await response.json();
    return data.data;
  }

  /**
   * Update organization details
   */
  static async updateOrganization(slug: string, req: UpdateOrganizationRequest): Promise<Organization> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to update organization");
    }

    const data: SuccessResponse<Organization> = await response.json();
    return data.data;
  }

  /**
   * Delete organization (owner only)
   */
  static async deleteOrganization(slug: string): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to delete organization");
    }
  }

  /**
   * Leave organization (non-owners only)
   */
  static async leaveOrganization(slug: string): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/leave`, {
      method: "POST",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to leave organization");
    }
  }

  /**
   * List organization members
   */
  static async listMembers(slug: string): Promise<OrganizationMember[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/members`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to fetch members");
    }

    const data: SuccessResponse<OrganizationMember[]> = await response.json();
    return data.data || [];
  }

  /**
   * Invite a new member to the organization
   */
  static async inviteMember(slug: string, req: InviteMemberRequest): Promise<OrganizationInvitation> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/members/invite`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to invite member");
    }

    const data: SuccessResponse<OrganizationInvitation> = await response.json();
    return data.data;
  }

  /**
   * Update a member's role
   */
  static async updateMemberRole(slug: string, userId: number, req: UpdateMemberRoleRequest): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/members/${userId}/role`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to update member role");
    }
  }

  /**
   * Remove a member from the organization
   */
  static async removeMember(slug: string, userId: number): Promise<void> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/members/${userId}`, {
      method: "DELETE",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to remove member");
    }
  }

  /**
   * List pending invitations for an organization
   */
  static async listInvitations(slug: string): Promise<OrganizationInvitation[]> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/invitations`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to fetch invitations");
    }

    const data: SuccessResponse<OrganizationInvitation[]> = await response.json();
    return data.data || [];
  }

  /**
   * Cancel a pending invitation
   */
  static async cancelInvitation(slug: string, invitationId: number): Promise<void> {
    const response = await authenticatedFetch(
      `${API_BASE_URL}/api/v1/organizations/${slug}/invitations/${invitationId}`,
      {
        method: "DELETE",
      }
    );

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to cancel invitation");
    }
  }

  /**
   * Accept an invitation to join an organization
   */
  static async acceptInvitation(token: string): Promise<Organization> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/invitations/accept?token=${token}`, {
      method: "POST",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to accept invitation");
    }

    const data: SuccessResponse<Organization> = await response.json();
    return data.data;
  }

  // ==================== Billing Methods ====================

  /**
   * Get organization billing information
   */
  static async getBilling(slug: string): Promise<OrgBillingInfo> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/billing`);

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to fetch billing information");
    }

    const data: SuccessResponse<OrgBillingInfo> = await response.json();
    return data.data;
  }

  /**
   * Create a checkout session for organization subscription
   */
  static async createCheckoutSession(slug: string, priceId: string): Promise<CheckoutSessionResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/billing/checkout`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ price_id: priceId }),
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to create checkout session");
    }

    const data: SuccessResponse<CheckoutSessionResponse> = await response.json();
    return data.data;
  }

  /**
   * Create a billing portal session for organization
   */
  static async createPortalSession(slug: string): Promise<PortalSessionResponse> {
    const response = await authenticatedFetch(`${API_BASE_URL}/api/v1/organizations/${slug}/billing/portal`, {
      method: "POST",
    });

    if (!response.ok) {
      throw await parseErrorResponse(response, "Failed to create billing portal session");
    }

    const data: SuccessResponse<PortalSessionResponse> = await response.json();
    return data.data;
  }
}
