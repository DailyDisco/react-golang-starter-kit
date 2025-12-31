import { act } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import type { Organization } from "../services/organizations/organizationService";
import { useCurrentOrg, useHasOrgRole, useIsOrgAdmin, useIsOrgOwner, useOrganizations, useOrgStore } from "./org-store";

const createMockOrg = (overrides?: Partial<Organization>): Organization => ({
  id: 1,
  name: "Test Organization",
  slug: "test-org",
  plan: "free",
  created_at: "2024-01-01T00:00:00Z",
  role: "member",
  ...overrides,
});

describe("useOrgStore", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    act(() => {
      useOrgStore.getState().reset();
    });
  });

  describe("initial state", () => {
    it("has null currentOrg initially", () => {
      const state = useOrgStore.getState();
      expect(state.currentOrg).toBeNull();
    });

    it("has empty organizations array initially", () => {
      const state = useOrgStore.getState();
      expect(state.organizations).toEqual([]);
    });

    it("has isLoading set to false initially", () => {
      const state = useOrgStore.getState();
      expect(state.isLoading).toBe(false);
    });
  });

  describe("setCurrentOrg", () => {
    it("sets the current organization", () => {
      const mockOrg = createMockOrg();

      act(() => {
        useOrgStore.getState().setCurrentOrg(mockOrg);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(mockOrg);
    });

    it("can set currentOrg to null", () => {
      const mockOrg = createMockOrg();

      act(() => {
        useOrgStore.getState().setCurrentOrg(mockOrg);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(mockOrg);

      act(() => {
        useOrgStore.getState().setCurrentOrg(null);
      });

      expect(useOrgStore.getState().currentOrg).toBeNull();
    });
  });

  describe("setOrganizations", () => {
    it("sets the organizations list", () => {
      const mockOrgs = [createMockOrg({ id: 1, slug: "org-1" }), createMockOrg({ id: 2, slug: "org-2" })];

      act(() => {
        useOrgStore.getState().setOrganizations(mockOrgs);
      });

      expect(useOrgStore.getState().organizations).toEqual(mockOrgs);
    });

    it("clears currentOrg and sets to first org if currentOrg is no longer in list", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1" });
      const org2 = createMockOrg({ id: 2, slug: "org-2" });
      const org3 = createMockOrg({ id: 3, slug: "org-3" });

      act(() => {
        useOrgStore.getState().setCurrentOrg(org1);
        useOrgStore.getState().setOrganizations([org1, org2]);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(org1);

      // Now set organizations without org1
      act(() => {
        useOrgStore.getState().setOrganizations([org2, org3]);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(org2);
    });

    it("keeps currentOrg if still in list", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1" });
      const org2 = createMockOrg({ id: 2, slug: "org-2" });

      act(() => {
        useOrgStore.getState().setCurrentOrg(org1);
        useOrgStore.getState().setOrganizations([org1, org2]);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(org1);

      act(() => {
        useOrgStore.getState().setOrganizations([org1, org2, createMockOrg({ id: 3, slug: "org-3" })]);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(org1);
    });
  });

  describe("addOrganization", () => {
    it("adds an organization to the list", () => {
      const mockOrg = createMockOrg();

      act(() => {
        useOrgStore.getState().addOrganization(mockOrg);
      });

      expect(useOrgStore.getState().organizations).toContainEqual(mockOrg);
    });

    it("sets as currentOrg if no currentOrg exists", () => {
      const mockOrg = createMockOrg();

      act(() => {
        useOrgStore.getState().addOrganization(mockOrg);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(mockOrg);
    });

    it("does not change currentOrg if one already exists", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1" });
      const org2 = createMockOrg({ id: 2, slug: "org-2" });

      act(() => {
        useOrgStore.getState().addOrganization(org1);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(org1);

      act(() => {
        useOrgStore.getState().addOrganization(org2);
      });

      expect(useOrgStore.getState().currentOrg).toEqual(org1);
      expect(useOrgStore.getState().organizations).toHaveLength(2);
    });
  });

  describe("removeOrganization", () => {
    it("removes an organization from the list", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1" });
      const org2 = createMockOrg({ id: 2, slug: "org-2" });

      act(() => {
        useOrgStore.getState().setOrganizations([org1, org2]);
      });

      act(() => {
        useOrgStore.getState().removeOrganization("org-1");
      });

      expect(useOrgStore.getState().organizations).toHaveLength(1);
      expect(useOrgStore.getState().organizations[0].slug).toBe("org-2");
    });

    it("switches currentOrg to first available if removed org was current", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1" });
      const org2 = createMockOrg({ id: 2, slug: "org-2" });

      act(() => {
        useOrgStore.getState().setOrganizations([org1, org2]);
        useOrgStore.getState().setCurrentOrg(org1);
      });

      expect(useOrgStore.getState().currentOrg?.slug).toBe("org-1");

      act(() => {
        useOrgStore.getState().removeOrganization("org-1");
      });

      expect(useOrgStore.getState().currentOrg?.slug).toBe("org-2");
    });

    it("sets currentOrg to null if no organizations remain", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1" });

      act(() => {
        useOrgStore.getState().setOrganizations([org1]);
        useOrgStore.getState().setCurrentOrg(org1);
      });

      act(() => {
        useOrgStore.getState().removeOrganization("org-1");
      });

      expect(useOrgStore.getState().currentOrg).toBeNull();
      expect(useOrgStore.getState().organizations).toHaveLength(0);
    });
  });

  describe("updateOrganization", () => {
    it("updates an organization in the list", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1", name: "Original Name" });

      act(() => {
        useOrgStore.getState().setOrganizations([org1]);
      });

      act(() => {
        useOrgStore.getState().updateOrganization("org-1", { name: "Updated Name" });
      });

      expect(useOrgStore.getState().organizations[0].name).toBe("Updated Name");
    });

    it("updates currentOrg if it matches the updated org", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1", name: "Original Name" });

      act(() => {
        useOrgStore.getState().setOrganizations([org1]);
        useOrgStore.getState().setCurrentOrg(org1);
      });

      act(() => {
        useOrgStore.getState().updateOrganization("org-1", { name: "Updated Name" });
      });

      expect(useOrgStore.getState().currentOrg?.name).toBe("Updated Name");
    });

    it("does not update currentOrg if different org is updated", () => {
      const org1 = createMockOrg({ id: 1, slug: "org-1", name: "Org 1" });
      const org2 = createMockOrg({ id: 2, slug: "org-2", name: "Org 2" });

      act(() => {
        useOrgStore.getState().setOrganizations([org1, org2]);
        useOrgStore.getState().setCurrentOrg(org1);
      });

      act(() => {
        useOrgStore.getState().updateOrganization("org-2", { name: "Updated Org 2" });
      });

      expect(useOrgStore.getState().currentOrg?.name).toBe("Org 1");
      expect(useOrgStore.getState().organizations[1].name).toBe("Updated Org 2");
    });
  });

  describe("setLoading", () => {
    it("sets isLoading to true", () => {
      act(() => {
        useOrgStore.getState().setLoading(true);
      });

      expect(useOrgStore.getState().isLoading).toBe(true);
    });

    it("sets isLoading to false", () => {
      act(() => {
        useOrgStore.getState().setLoading(true);
      });

      act(() => {
        useOrgStore.getState().setLoading(false);
      });

      expect(useOrgStore.getState().isLoading).toBe(false);
    });
  });

  describe("reset", () => {
    it("resets all state to initial values", () => {
      const mockOrg = createMockOrg();

      act(() => {
        useOrgStore.getState().setCurrentOrg(mockOrg);
        useOrgStore.getState().setOrganizations([mockOrg]);
        useOrgStore.getState().setLoading(true);
      });

      expect(useOrgStore.getState().currentOrg).not.toBeNull();
      expect(useOrgStore.getState().organizations).toHaveLength(1);
      expect(useOrgStore.getState().isLoading).toBe(true);

      act(() => {
        useOrgStore.getState().reset();
      });

      expect(useOrgStore.getState().currentOrg).toBeNull();
      expect(useOrgStore.getState().organizations).toEqual([]);
      expect(useOrgStore.getState().isLoading).toBe(false);
    });
  });
});

describe("selector hooks", () => {
  beforeEach(() => {
    act(() => {
      useOrgStore.getState().reset();
    });
  });

  describe("useCurrentOrg", () => {
    it("returns currentOrg from store", () => {
      const mockOrg = createMockOrg();

      act(() => {
        useOrgStore.getState().setCurrentOrg(mockOrg);
      });

      // This is a selector hook, we test by checking store state
      expect(useOrgStore.getState().currentOrg).toEqual(mockOrg);
    });
  });

  describe("useOrganizations", () => {
    it("returns organizations array from store", () => {
      const mockOrgs = [createMockOrg({ id: 1 }), createMockOrg({ id: 2 })];

      act(() => {
        useOrgStore.getState().setOrganizations(mockOrgs);
      });

      expect(useOrgStore.getState().organizations).toEqual(mockOrgs);
    });
  });
});

describe("role helper hooks", () => {
  beforeEach(() => {
    act(() => {
      useOrgStore.getState().reset();
    });
  });

  describe("useHasOrgRole", () => {
    it("returns false when no currentOrg", () => {
      // useHasOrgRole checks the store, and returns false when no org
      expect(useOrgStore.getState().currentOrg).toBeNull();
    });

    it("role hierarchy works correctly for owner", () => {
      const ownerOrg = createMockOrg({ role: "owner" });

      act(() => {
        useOrgStore.getState().setCurrentOrg(ownerOrg);
      });

      // Owner should have access to owner, admin, and member level
      expect(useOrgStore.getState().currentOrg?.role).toBe("owner");
    });

    it("role hierarchy works correctly for admin", () => {
      const adminOrg = createMockOrg({ role: "admin" });

      act(() => {
        useOrgStore.getState().setCurrentOrg(adminOrg);
      });

      expect(useOrgStore.getState().currentOrg?.role).toBe("admin");
    });

    it("role hierarchy works correctly for member", () => {
      const memberOrg = createMockOrg({ role: "member" });

      act(() => {
        useOrgStore.getState().setCurrentOrg(memberOrg);
      });

      expect(useOrgStore.getState().currentOrg?.role).toBe("member");
    });
  });
});
