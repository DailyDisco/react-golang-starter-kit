import type { Organization } from "@/services/organizations/organizationService";
import { create } from "zustand";
import { persist } from "zustand/middleware";

interface OrgState {
  // Current selected organization
  currentOrg: Organization | null;

  // List of all organizations the user belongs to
  organizations: Organization[];

  // Loading states
  isLoading: boolean;

  // Actions
  setCurrentOrg: (org: Organization | null) => void;
  setOrganizations: (orgs: Organization[]) => void;
  addOrganization: (org: Organization) => void;
  removeOrganization: (slug: string) => void;
  updateOrganization: (slug: string, updates: Partial<Organization>) => void;
  setLoading: (loading: boolean) => void;
  reset: () => void;
}

const initialState = {
  currentOrg: null,
  organizations: [],
  isLoading: false,
};

export const useOrgStore = create<OrgState>()(
  persist(
    (set, get) => ({
      ...initialState,

      setCurrentOrg: (org) => set({ currentOrg: org }),

      setOrganizations: (orgs) => {
        const currentOrg = get().currentOrg;
        // If current org is no longer in the list, clear it
        if (currentOrg && !orgs.find((o) => o.slug === currentOrg.slug)) {
          set({ organizations: orgs, currentOrg: orgs[0] || null });
        } else {
          set({ organizations: orgs });
        }
      },

      addOrganization: (org) =>
        set((state) => ({
          organizations: [...state.organizations, org],
          // If no current org, set this as current
          currentOrg: state.currentOrg || org,
        })),

      removeOrganization: (slug) =>
        set((state) => {
          const orgs = state.organizations.filter((o) => o.slug !== slug);
          return {
            organizations: orgs,
            // If removed org was current, switch to first available
            currentOrg: state.currentOrg?.slug === slug ? orgs[0] || null : state.currentOrg,
          };
        }),

      updateOrganization: (slug, updates) =>
        set((state) => ({
          organizations: state.organizations.map((o) => (o.slug === slug ? { ...o, ...updates } : o)),
          currentOrg: state.currentOrg?.slug === slug ? { ...state.currentOrg, ...updates } : state.currentOrg,
        })),

      setLoading: (loading) => set({ isLoading: loading }),

      reset: () => set(initialState),
    }),
    {
      name: "org-storage",
      partialize: (state) => ({
        // Only persist the current org slug, not the full object
        currentOrgSlug: state.currentOrg?.slug,
      }),
      onRehydrateStorage: () => (state) => {
        // After rehydration, we need to fetch the full org data
        // This will be done by the component that uses the store
        if (state) {
          state.setLoading(false);
        }
      },
    }
  )
);

// Selector hooks for common use cases
export const useCurrentOrg = () => useOrgStore((state) => state.currentOrg);
export const useOrganizations = () => useOrgStore((state) => state.organizations);
export const useIsOrgLoading = () => useOrgStore((state) => state.isLoading);

// Helper to check if user has a specific role in current org
export const useHasOrgRole = (minRole: "owner" | "admin" | "member") => {
  const currentOrg = useCurrentOrg();
  if (!currentOrg) return false;

  const roleHierarchy = { owner: 3, admin: 2, member: 1 };
  return roleHierarchy[currentOrg.role] >= roleHierarchy[minRole];
};

// Helper to check if user is org admin or owner
export const useIsOrgAdmin = () => useHasOrgRole("admin");
export const useIsOrgOwner = () => useHasOrgRole("owner");
