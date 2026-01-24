import { describe, expect, it, vi } from "vitest";

import {
  adminAuditLogsQueryOptions,
  adminFeatureFlagsQueryOptions,
  adminStatsQueryOptions,
  apiKeysQueryOptions,
  billingConfigQueryOptions,
  billingPlansQueryOptions,
  connectedAccountsQueryOptions,
  currentUserQueryOptions,
  filesQueryOptions,
  loginHistoryQueryOptions,
  orgInvitationsQueryOptions,
  orgMembersQueryOptions,
  preferencesQueryOptions,
  sessionsQueryOptions,
  storageStatusQueryOptions,
  subscriptionQueryOptions,
} from "./route-query-options";

// Mock the services
vi.mock("../services", () => ({
  AuthService: {
    getCurrentUser: vi.fn(),
  },
}));

vi.mock("../services/admin/adminService", () => ({
  AdminService: {
    getStats: vi.fn(),
    getFeatureFlags: vi.fn(),
    getAuditLogs: vi.fn(),
  },
}));

vi.mock("../services/billing/billingService", () => ({
  BillingService: {
    getConfig: vi.fn(),
    getPlans: vi.fn(),
    getSubscription: vi.fn(),
  },
}));

vi.mock("../services/files/fileService", () => ({
  FileService: {
    fetchFiles: vi.fn(),
    getStorageStatus: vi.fn(),
  },
}));

vi.mock("../services/organizations/organizationService", () => ({
  OrganizationService: {
    listMembers: vi.fn(),
    listInvitations: vi.fn(),
  },
}));

vi.mock("../services/settings/settingsService", () => ({
  SettingsService: {
    getPreferences: vi.fn(),
    getSessions: vi.fn(),
    getAPIKeys: vi.fn(),
    getLoginHistory: vi.fn(),
    getConnectedAccounts: vi.fn(),
  },
}));

vi.mock("./cache-config", () => ({
  CACHE_TIMES: {
    PREFERENCES: 60000,
    SESSIONS: 30000,
    API_KEYS: 60000,
    LOGIN_HISTORY: 60000,
    FILES: 30000,
    STORAGE_STATUS: 60000,
    SUBSCRIPTION: 60000,
  },
  GC_TIMES: {
    BILLING: 3600000,
  },
  SWR_CONFIG: {
    STABLE: {
      staleTime: Infinity,
      refetchOnWindowFocus: false,
    },
  },
}));

vi.mock("./query-keys", () => ({
  queryKeys: {
    auth: {
      user: ["auth", "user"],
    },
    settings: {
      preferences: () => ["settings", "preferences"],
      sessions: () => ["settings", "sessions"],
      apiKeys: () => ["settings", "api-keys"],
      loginHistory: () => ["settings", "login-history"],
      connectedAccounts: () => ["settings", "connected-accounts"],
    },
    files: {
      list: (limit?: number, offset?: number) => ["files", "list", { limit, offset }],
      storageStatus: () => ["files", "storage-status"],
    },
    billing: {
      config: () => ["billing", "config"],
      plans: () => ["billing", "plans"],
      subscription: () => ["billing", "subscription"],
    },
    organizations: {
      members: (slug: string) => ["organizations", slug, "members"],
      invitations: (slug: string) => ["organizations", slug, "invitations"],
    },
  },
}));

describe("Route Query Options", () => {
  describe("currentUserQueryOptions", () => {
    it("should return query options with correct queryKey", () => {
      const options = currentUserQueryOptions();

      expect(options.queryKey).toEqual(["auth", "user"]);
      expect(options.staleTime).toBe(5 * 60 * 1000);
      expect(options.retry).toBe(false);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("adminStatsQueryOptions", () => {
    it("should return query options for admin stats", () => {
      const options = adminStatsQueryOptions();

      expect(options.queryKey).toEqual(["admin", "stats"]);
      expect(options.staleTime).toBe(30 * 1000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("adminFeatureFlagsQueryOptions", () => {
    it("should return query options for feature flags", () => {
      const options = adminFeatureFlagsQueryOptions();

      expect(options.queryKey).toEqual(["admin", "feature-flags"]);
      expect(options.staleTime).toBe(60 * 1000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("adminAuditLogsQueryOptions", () => {
    it("should return query options without filter", () => {
      const options = adminAuditLogsQueryOptions();

      expect(options.queryKey).toEqual(["admin", "audit-logs", undefined]);
      expect(options.staleTime).toBe(30 * 1000);
      expect(typeof options.queryFn).toBe("function");
    });

    it("should return query options with filter", () => {
      const filter = { page: 1, limit: 20, action: "create" };
      const options = adminAuditLogsQueryOptions(filter);

      expect(options.queryKey).toEqual(["admin", "audit-logs", filter]);
      expect(typeof options.queryFn).toBe("function");
    });

    it("should include target_type in filter", () => {
      const filter = { target_type: "user" };
      const options = adminAuditLogsQueryOptions(filter);

      expect(options.queryKey).toEqual(["admin", "audit-logs", filter]);
    });
  });

  describe("orgMembersQueryOptions", () => {
    it("should return query options for org members", () => {
      const options = orgMembersQueryOptions("my-org");

      expect(options.queryKey).toEqual(["organizations", "my-org", "members"]);
      expect(options.staleTime).toBe(60 * 1000);
      expect(typeof options.queryFn).toBe("function");
    });

    it("should use different queryKey for different orgs", () => {
      const options1 = orgMembersQueryOptions("org-1");
      const options2 = orgMembersQueryOptions("org-2");

      expect(options1.queryKey).not.toEqual(options2.queryKey);
    });
  });

  describe("orgInvitationsQueryOptions", () => {
    it("should return query options for org invitations", () => {
      const options = orgInvitationsQueryOptions("my-org");

      expect(options.queryKey).toEqual(["organizations", "my-org", "invitations"]);
      expect(options.staleTime).toBe(60 * 1000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("preferencesQueryOptions", () => {
    it("should return query options for preferences", () => {
      const options = preferencesQueryOptions();

      expect(options.queryKey).toEqual(["settings", "preferences"]);
      expect(options.staleTime).toBe(60000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("sessionsQueryOptions", () => {
    it("should return query options for sessions", () => {
      const options = sessionsQueryOptions();

      expect(options.queryKey).toEqual(["settings", "sessions"]);
      expect(options.staleTime).toBe(30000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("apiKeysQueryOptions", () => {
    it("should return query options for API keys", () => {
      const options = apiKeysQueryOptions();

      expect(options.queryKey).toEqual(["settings", "api-keys"]);
      expect(options.staleTime).toBe(60000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("loginHistoryQueryOptions", () => {
    it("should return query options with default limit", () => {
      const options = loginHistoryQueryOptions();

      expect(options.queryKey).toEqual(["settings", "login-history", { limit: 50 }]);
      expect(options.staleTime).toBe(60000);
      expect(typeof options.queryFn).toBe("function");
    });

    it("should return query options with custom limit", () => {
      const options = loginHistoryQueryOptions(100);

      expect(options.queryKey).toEqual(["settings", "login-history", { limit: 100 }]);
    });
  });

  describe("connectedAccountsQueryOptions", () => {
    it("should return query options for connected accounts", () => {
      const options = connectedAccountsQueryOptions();

      expect(options.queryKey).toEqual(["settings", "connected-accounts"]);
      expect(options.staleTime).toBe(60000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("filesQueryOptions", () => {
    it("should return query options without pagination", () => {
      const options = filesQueryOptions();

      expect(options.queryKey).toEqual(["files", "list", { limit: undefined, offset: undefined }]);
      expect(options.staleTime).toBe(30000);
      expect(typeof options.queryFn).toBe("function");
    });

    it("should return query options with pagination", () => {
      const options = filesQueryOptions(20, 0);

      expect(options.queryKey).toEqual(["files", "list", { limit: 20, offset: 0 }]);
    });

    it("should return query options with offset", () => {
      const options = filesQueryOptions(20, 40);

      expect(options.queryKey).toEqual(["files", "list", { limit: 20, offset: 40 }]);
    });
  });

  describe("storageStatusQueryOptions", () => {
    it("should return query options for storage status", () => {
      const options = storageStatusQueryOptions();

      expect(options.queryKey).toEqual(["files", "storage-status"]);
      expect(options.staleTime).toBe(60000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("billingConfigQueryOptions", () => {
    it("should return query options for billing config with SWR pattern", () => {
      const options = billingConfigQueryOptions();

      expect(options.queryKey).toEqual(["billing", "config"]);
      expect(options.gcTime).toBe(3600000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("billingPlansQueryOptions", () => {
    it("should return query options for billing plans with SWR pattern", () => {
      const options = billingPlansQueryOptions();

      expect(options.queryKey).toEqual(["billing", "plans"]);
      expect(options.gcTime).toBe(3600000);
      expect(typeof options.queryFn).toBe("function");
    });
  });

  describe("subscriptionQueryOptions", () => {
    it("should return query options for subscription", () => {
      const options = subscriptionQueryOptions();

      expect(options.queryKey).toEqual(["billing", "subscription"]);
      expect(options.staleTime).toBe(60000);
      expect(typeof options.queryFn).toBe("function");
    });
  });
});
