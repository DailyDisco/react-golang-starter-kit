import { beforeEach, describe, expect, it, vi } from "vitest";

import { OrganizationService, type Organization } from "./organizationService";

// Mock the API client functions
vi.mock("../api/client", () => ({
  API_BASE_URL: "http://localhost:8080",
  authenticatedFetch: vi.fn(),
  parseErrorResponse: vi.fn(),
}));

const createMockOrg = (overrides?: Partial<Organization>): Organization => ({
  id: 1,
  name: "Test Organization",
  slug: "test-org",
  plan: "free",
  created_at: "2024-01-01T00:00:00Z",
  role: "member",
  ...overrides,
});

describe("OrganizationService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("listOrganizations", () => {
    it("fetches list of organizations", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const mockOrgs = [createMockOrg({ id: 1 }), createMockOrg({ id: 2, slug: "org-2" })];

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: mockOrgs }),
      } as unknown as Response);

      const result = await OrganizationService.listOrganizations();

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations");
      expect(result).toEqual(mockOrgs);
    });

    it("returns empty array when data is null", async () => {
      const { authenticatedFetch } = await import("../api/client");

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: null }),
      } as unknown as Response);

      const result = await OrganizationService.listOrganizations();

      expect(result).toEqual([]);
    });

    it("throws error when request fails", async () => {
      const { authenticatedFetch, parseErrorResponse } = await import("../api/client");

      const mockError = new Error("Failed to fetch organizations");
      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: false,
      } as unknown as Response);
      vi.mocked(parseErrorResponse).mockResolvedValue(mockError);

      await expect(OrganizationService.listOrganizations()).rejects.toThrow(mockError);
    });
  });

  describe("createOrganization", () => {
    it("creates a new organization", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const newOrg = createMockOrg({ role: "owner" });

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: newOrg }),
      } as unknown as Response);

      const result = await OrganizationService.createOrganization({
        name: "Test Organization",
        slug: "test-org",
      });

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "Test Organization", slug: "test-org" }),
      });
      expect(result).toEqual(newOrg);
    });
  });

  describe("getOrganization", () => {
    it("fetches organization by slug", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const mockOrg = createMockOrg();

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: mockOrg }),
      } as unknown as Response);

      const result = await OrganizationService.getOrganization("test-org");

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations/test-org");
      expect(result).toEqual(mockOrg);
    });
  });

  describe("updateOrganization", () => {
    it("updates organization details", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const updatedOrg = createMockOrg({ name: "Updated Name" });

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: updatedOrg }),
      } as unknown as Response);

      const result = await OrganizationService.updateOrganization("test-org", {
        name: "Updated Name",
      });

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations/test-org", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "Updated Name" }),
      });
      expect(result).toEqual(updatedOrg);
    });
  });

  describe("deleteOrganization", () => {
    it("deletes an organization", async () => {
      const { authenticatedFetch } = await import("../api/client");

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
      } as unknown as Response);

      await OrganizationService.deleteOrganization("test-org");

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations/test-org", {
        method: "DELETE",
      });
    });
  });

  describe("leaveOrganization", () => {
    it("leaves an organization", async () => {
      const { authenticatedFetch } = await import("../api/client");

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
      } as unknown as Response);

      await OrganizationService.leaveOrganization("test-org");

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations/test-org/leave", {
        method: "POST",
      });
    });
  });

  describe("listMembers", () => {
    it("fetches organization members", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const mockMembers = [
        { id: 1, user_id: 1, email: "user1@example.com", name: "User 1", role: "owner", status: "active" },
        { id: 2, user_id: 2, email: "user2@example.com", name: "User 2", role: "member", status: "active" },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: mockMembers }),
      } as unknown as Response);

      const result = await OrganizationService.listMembers("test-org");

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations/test-org/members");
      expect(result).toEqual(mockMembers);
    });
  });

  describe("inviteMember", () => {
    it("invites a member to organization", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const mockInvitation = {
        id: 1,
        email: "new@example.com",
        role: "member",
        invited_by: "Admin User",
        expires_at: "2024-02-01T00:00:00Z",
        created_at: "2024-01-01T00:00:00Z",
      };

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: mockInvitation }),
      } as unknown as Response);

      const result = await OrganizationService.inviteMember("test-org", {
        email: "new@example.com",
        role: "member",
      });

      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/organizations/test-org/members/invite",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email: "new@example.com", role: "member" }),
        }
      );
      expect(result).toEqual(mockInvitation);
    });
  });

  describe("updateMemberRole", () => {
    it("updates member role", async () => {
      const { authenticatedFetch } = await import("../api/client");

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
      } as unknown as Response);

      await OrganizationService.updateMemberRole("test-org", 5, { role: "admin" });

      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/organizations/test-org/members/5/role",
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ role: "admin" }),
        }
      );
    });
  });

  describe("removeMember", () => {
    it("removes a member from organization", async () => {
      const { authenticatedFetch } = await import("../api/client");

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
      } as unknown as Response);

      await OrganizationService.removeMember("test-org", 5);

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/organizations/test-org/members/5", {
        method: "DELETE",
      });
    });
  });

  describe("listInvitations", () => {
    it("fetches pending invitations", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const mockInvitations = [{ id: 1, email: "pending@example.com", role: "member", invited_by: "Admin" }];

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: mockInvitations }),
      } as unknown as Response);

      const result = await OrganizationService.listInvitations("test-org");

      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/organizations/test-org/invitations"
      );
      expect(result).toEqual(mockInvitations);
    });
  });

  describe("cancelInvitation", () => {
    it("cancels a pending invitation", async () => {
      const { authenticatedFetch } = await import("../api/client");

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
      } as unknown as Response);

      await OrganizationService.cancelInvitation("test-org", 5);

      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/organizations/test-org/invitations/5",
        { method: "DELETE" }
      );
    });
  });

  describe("acceptInvitation", () => {
    it("accepts an invitation to join organization", async () => {
      const { authenticatedFetch } = await import("../api/client");

      const mockOrg = createMockOrg({ role: "member" });

      vi.mocked(authenticatedFetch).mockResolvedValue({
        ok: true,
        json: vi.fn().mockResolvedValue({ success: true, data: mockOrg }),
      } as unknown as Response);

      const result = await OrganizationService.acceptInvitation("invite-token-123");

      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/invitations/accept?token=invite-token-123",
        { method: "POST" }
      );
      expect(result).toEqual(mockOrg);
    });
  });
});
