import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  AdminService,
  AdminSettingsService,
  AnnouncementService,
  FeatureFlagService,
  type AdminStats,
  type AuditLogsResponse,
  type CreateFeatureFlagRequest,
  type FeatureFlag,
  type FeatureFlagsResponse,
} from "./adminService";

// Mock the apiClient
const mockApiClient = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
};

vi.mock("../api/client", () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

describe("AdminService", () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    const { apiClient } = await import("../api/client");
    Object.assign(mockApiClient, apiClient);
  });

  describe("getStats", () => {
    it("fetches admin stats", async () => {
      const { apiClient } = await import("../api/client");

      const mockStats: AdminStats = {
        total_users: 100,
        active_users: 80,
        verified_users: 75,
        new_users_today: 5,
        new_users_this_week: 20,
        new_users_this_month: 50,
        total_subscriptions: 30,
        active_subscriptions: 25,
        canceled_subscriptions: 5,
        total_files: 200,
        total_file_size: 1024000,
        users_by_role: { user: 90, admin: 10 },
      };

      vi.mocked(apiClient.get).mockResolvedValue(mockStats);

      const result = await AdminService.getStats();

      expect(apiClient.get).toHaveBeenCalledWith("/admin/stats");
      expect(result).toEqual(mockStats);
    });
  });

  describe("getAuditLogs", () => {
    it("fetches audit logs without filters", async () => {
      const { apiClient } = await import("../api/client");

      const mockResponse: AuditLogsResponse = {
        logs: [],
        count: 0,
        total: 0,
        page: 1,
        limit: 20,
        total_pages: 0,
      };

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse);

      const result = await AdminService.getAuditLogs();

      expect(apiClient.get).toHaveBeenCalledWith("/admin/audit-logs");
      expect(result).toEqual(mockResponse);
    });

    it("fetches audit logs with filters", async () => {
      const { apiClient } = await import("../api/client");

      const mockResponse: AuditLogsResponse = {
        logs: [],
        count: 0,
        total: 0,
        page: 1,
        limit: 10,
        total_pages: 0,
      };

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse);

      await AdminService.getAuditLogs({
        user_id: 5,
        action: "login",
        page: 1,
        limit: 10,
      });

      expect(apiClient.get).toHaveBeenCalledWith(expect.stringContaining("/admin/audit-logs?"));
      expect(apiClient.get).toHaveBeenCalledWith(expect.stringContaining("user_id=5"));
      expect(apiClient.get).toHaveBeenCalledWith(expect.stringContaining("action=login"));
    });
  });

  describe("updateUserRole", () => {
    it("updates user role", async () => {
      const { apiClient } = await import("../api/client");

      const mockUser = { id: 5, name: "Test", email: "test@example.com", role: "admin" };
      vi.mocked(apiClient.put).mockResolvedValue(mockUser);

      const result = await AdminService.updateUserRole(5, "admin");

      expect(apiClient.put).toHaveBeenCalledWith("/admin/users/5/role", { role: "admin" });
      expect(result).toEqual(mockUser);
    });
  });

  describe("deactivateUser", () => {
    it("deactivates a user", async () => {
      const { apiClient } = await import("../api/client");

      vi.mocked(apiClient.post).mockResolvedValue(undefined);

      await AdminService.deactivateUser(5);

      expect(apiClient.post).toHaveBeenCalledWith("/admin/users/5/deactivate", {});
    });
  });

  describe("reactivateUser", () => {
    it("reactivates a user", async () => {
      const { apiClient } = await import("../api/client");

      vi.mocked(apiClient.post).mockResolvedValue(undefined);

      await AdminService.reactivateUser(5);

      expect(apiClient.post).toHaveBeenCalledWith("/admin/users/5/reactivate", {});
    });
  });

  describe("impersonateUser", () => {
    it("starts user impersonation", async () => {
      const { apiClient } = await import("../api/client");

      const mockResponse = {
        user: { id: 5, name: "Target User" },
        token: "impersonation-token",
        original_user_id: 1,
      };

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse);

      const result = await AdminService.impersonateUser({
        user_id: 5,
        reason: "Debugging user issue",
      });

      expect(apiClient.post).toHaveBeenCalledWith("/admin/impersonate", {
        user_id: 5,
        reason: "Debugging user issue",
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("stopImpersonation", () => {
    it("stops user impersonation", async () => {
      const { apiClient } = await import("../api/client");

      const mockResponse = {
        user: { id: 1, name: "Original User" },
        token: "original-token",
      };

      vi.mocked(apiClient.post).mockResolvedValue(mockResponse);

      const result = await AdminService.stopImpersonation();

      expect(apiClient.post).toHaveBeenCalledWith("/admin/stop-impersonate", {});
      expect(result).toEqual(mockResponse);
    });
  });

  describe("Feature Flags", () => {
    it("gets all feature flags", async () => {
      const { apiClient } = await import("../api/client");

      const mockResponse: FeatureFlagsResponse = {
        flags: [
          {
            id: 1,
            key: "new_feature",
            name: "New Feature",
            enabled: true,
            rollout_percentage: 100,
            created_at: "2024-01-01",
            updated_at: "2024-01-01",
          },
        ],
        count: 1,
      };

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse);

      const result = await AdminService.getFeatureFlags();

      expect(apiClient.get).toHaveBeenCalledWith("/admin/feature-flags");
      expect(result).toEqual(mockResponse);
    });

    it("creates a feature flag", async () => {
      const { apiClient } = await import("../api/client");

      const request: CreateFeatureFlagRequest = {
        key: "new_feature",
        name: "New Feature",
        enabled: false,
        rollout_percentage: 50,
      };

      const mockFlag: FeatureFlag = {
        id: 1,
        ...request,
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      } as FeatureFlag;

      vi.mocked(apiClient.post).mockResolvedValue(mockFlag);

      const result = await AdminService.createFeatureFlag(request);

      expect(apiClient.post).toHaveBeenCalledWith("/admin/feature-flags", request);
      expect(result).toEqual(mockFlag);
    });

    it("updates a feature flag", async () => {
      const { apiClient } = await import("../api/client");

      const mockFlag: FeatureFlag = {
        id: 1,
        key: "new_feature",
        name: "Updated Feature",
        enabled: true,
        rollout_percentage: 100,
        created_at: "2024-01-01",
        updated_at: "2024-01-02",
      };

      vi.mocked(apiClient.put).mockResolvedValue(mockFlag);

      const result = await AdminService.updateFeatureFlag("new_feature", {
        name: "Updated Feature",
        enabled: true,
      });

      expect(apiClient.put).toHaveBeenCalledWith("/admin/feature-flags/new_feature", {
        name: "Updated Feature",
        enabled: true,
      });
      expect(result).toEqual(mockFlag);
    });

    it("deletes a feature flag", async () => {
      const { apiClient } = await import("../api/client");

      vi.mocked(apiClient.delete).mockResolvedValue(undefined);

      await AdminService.deleteFeatureFlag("old_feature");

      expect(apiClient.delete).toHaveBeenCalledWith("/admin/feature-flags/old_feature");
    });
  });
});

describe("FeatureFlagService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("gets user feature flags", async () => {
    const { apiClient } = await import("../api/client");

    const mockFlags = {
      new_feature: true,
      beta_feature: false,
    };

    vi.mocked(apiClient.get).mockResolvedValue(mockFlags);

    const result = await FeatureFlagService.getFlags();

    expect(apiClient.get).toHaveBeenCalledWith("/feature-flags");
    expect(result).toEqual(mockFlags);
  });
});

describe("AdminSettingsService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Email Settings", () => {
    it("gets email settings", async () => {
      const { apiClient } = await import("../api/client");

      const mockSettings = {
        smtp_host: "smtp.example.com",
        smtp_port: 587,
        smtp_enabled: true,
      };

      vi.mocked(apiClient.get).mockResolvedValue(mockSettings);

      const result = await AdminSettingsService.getEmailSettings();

      expect(apiClient.get).toHaveBeenCalledWith("/admin/settings/email");
      expect(result).toEqual(mockSettings);
    });

    it("updates email settings", async () => {
      const { apiClient } = await import("../api/client");

      vi.mocked(apiClient.put).mockResolvedValue(undefined);

      await AdminSettingsService.updateEmailSettings({ smtp_host: "new.smtp.com" });

      expect(apiClient.put).toHaveBeenCalledWith("/admin/settings/email", {
        smtp_host: "new.smtp.com",
      });
    });

    it("tests email settings", async () => {
      const { apiClient } = await import("../api/client");

      const mockResponse = { success: true, message: "Test email sent" };
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse);

      const result = await AdminSettingsService.testEmailSettings();

      expect(apiClient.post).toHaveBeenCalledWith("/admin/settings/email/test", {});
      expect(result).toEqual(mockResponse);
    });
  });

  describe("System Health", () => {
    it("gets system health", async () => {
      const { apiClient } = await import("../api/client");

      const mockHealth = {
        status: "healthy",
        timestamp: "2024-01-01T00:00:00Z",
        components: [],
      };

      vi.mocked(apiClient.get).mockResolvedValue(mockHealth);

      const result = await AdminSettingsService.getSystemHealth();

      expect(apiClient.get).toHaveBeenCalledWith("/admin/health");
      expect(result).toEqual(mockHealth);
    });
  });
});

describe("AnnouncementService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("gets active announcements", async () => {
    const { apiClient } = await import("../api/client");

    const mockAnnouncements = [
      {
        id: 1,
        title: "Maintenance",
        message: "Scheduled maintenance tonight",
        type: "info",
        is_active: true,
      },
    ];

    vi.mocked(apiClient.get).mockResolvedValue(mockAnnouncements);

    const result = await AnnouncementService.getActiveAnnouncements();

    expect(apiClient.get).toHaveBeenCalledWith("/announcements");
    expect(result).toEqual(mockAnnouncements);
  });

  it("dismisses an announcement", async () => {
    const { apiClient } = await import("../api/client");

    vi.mocked(apiClient.post).mockResolvedValue(undefined);

    await AnnouncementService.dismissAnnouncement(1);

    expect(apiClient.post).toHaveBeenCalledWith("/announcements/1/dismiss", {});
  });
});
