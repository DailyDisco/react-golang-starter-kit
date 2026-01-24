import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, authenticatedFetch, parseErrorResponse } from "../api/client";
import {
  DATE_FORMATS,
  LANGUAGES,
  SettingsService,
  TIMEZONES,
  type ChangePasswordRequest,
  type CreateAPIKeyRequest,
  type UpdateAPIKeyRequest,
  type UpdatePreferencesRequest,
} from "./settingsService";

// Mock the API client module
vi.mock("../api/client", () => ({
  API_BASE_URL: "http://localhost:8080",
  ApiError: class ApiError extends Error {
    code: string;
    statusCode: number;
    constructor(message: string, code: string, statusCode: number) {
      super(message);
      this.name = "ApiError";
      this.code = code;
      this.statusCode = statusCode;
    }
  },
  authenticatedFetch: vi.fn(),
  parseErrorResponse: vi.fn(),
  createHeaders: vi.fn(() => ({})),
}));

describe("SettingsService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  // ==================== API Keys Tests ====================

  describe("getAPIKeys", () => {
    it("should return array of API keys on success", async () => {
      const mockKeys = [
        { id: 1, provider: "openai", name: "Test Key", key_preview: "sk-...abc" },
        { id: 2, provider: "anthropic", name: "Claude Key", key_preview: "sk-...xyz" },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ keys: mockKeys }),
      } as Response);

      const result = await SettingsService.getAPIKeys();

      expect(result).toEqual(mockKeys);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/api-keys");
    });

    it("should return empty array when no keys exist", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ keys: null }),
      } as Response);

      const result = await SettingsService.getAPIKeys();

      expect(result).toEqual([]);
    });

    it("should throw error on API failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.getAPIKeys()).rejects.toThrow();
    });
  });

  describe("getAPIKey", () => {
    it("should return single API key", async () => {
      const mockKey = { id: 1, provider: "openai", name: "Test Key" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockKey,
      } as Response);

      const result = await SettingsService.getAPIKey(1);

      expect(result).toEqual(mockKey);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/api-keys/1");
    });

    it("should throw error when key not found", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Key not found", "NOT_FOUND", 404));

      await expect(SettingsService.getAPIKey(999)).rejects.toThrow();
    });
  });

  describe("createAPIKey", () => {
    it("should create API key successfully", async () => {
      const request: CreateAPIKeyRequest = {
        provider: "openai",
        name: "New Key",
        api_key: "sk-test123",
      };
      const mockKey = { id: 1, ...request, key_preview: "sk-...123" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockKey,
      } as Response);

      const result = await SettingsService.createAPIKey(request);

      expect(result).toEqual(mockKey);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/api-keys", {
        method: "POST",
        body: JSON.stringify(request),
      });
    });

    it("should throw error on creation failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid key", "INVALID_KEY", 400));

      await expect(
        SettingsService.createAPIKey({ provider: "openai", name: "Test", api_key: "invalid" })
      ).rejects.toThrow();
    });
  });

  describe("updateAPIKey", () => {
    it("should update API key successfully", async () => {
      const request: UpdateAPIKeyRequest = { name: "Updated Name" };
      const mockKey = { id: 1, provider: "openai", name: "Updated Name" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockKey,
      } as Response);

      const result = await SettingsService.updateAPIKey(1, request);

      expect(result).toEqual(mockKey);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/api-keys/1", {
        method: "PUT",
        body: JSON.stringify(request),
      });
    });

    it("should throw error on update failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Key not found", "NOT_FOUND", 404));

      await expect(SettingsService.updateAPIKey(999, { name: "Test" })).rejects.toThrow();
    });
  });

  describe("deleteAPIKey", () => {
    it("should delete API key successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.deleteAPIKey(1)).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/api-keys/1", {
        method: "DELETE",
      });
    });

    it("should throw error on delete failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Key not found", "NOT_FOUND", 404));

      await expect(SettingsService.deleteAPIKey(999)).rejects.toThrow();
    });
  });

  describe("testAPIKey", () => {
    it("should test API key successfully", async () => {
      const mockResult = { success: true, message: "Key is valid" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResult,
      } as Response);

      const result = await SettingsService.testAPIKey(1);

      expect(result).toEqual(mockResult);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/api-keys/1/test", {
        method: "POST",
      });
    });

    it("should throw error on test failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid key", "INVALID_KEY", 400));

      await expect(SettingsService.testAPIKey(1)).rejects.toThrow();
    });
  });

  // ==================== Preferences Tests ====================

  describe("getPreferences", () => {
    it("should return user preferences", async () => {
      const mockPreferences = {
        id: 1,
        user_id: 1,
        theme: "dark",
        timezone: "America/New_York",
        language: "en",
      };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockPreferences }),
      } as Response);

      const result = await SettingsService.getPreferences();

      expect(result).toEqual(mockPreferences);
    });

    it("should handle legacy response format", async () => {
      const mockPreferences = { theme: "light", timezone: "UTC" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockPreferences,
      } as Response);

      const result = await SettingsService.getPreferences();

      expect(result).toEqual(mockPreferences);
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.getPreferences()).rejects.toThrow();
    });
  });

  describe("updatePreferences", () => {
    it("should update preferences successfully", async () => {
      const request: UpdatePreferencesRequest = { theme: "dark" };
      const mockPreferences = { theme: "dark", timezone: "UTC" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockPreferences }),
      } as Response);

      const result = await SettingsService.updatePreferences(request);

      expect(result).toEqual(mockPreferences);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/preferences", {
        method: "PUT",
        body: JSON.stringify(request),
      });
    });

    it("should throw error on update failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid preferences", "INVALID", 400));

      await expect(SettingsService.updatePreferences({ theme: "invalid" as any })).rejects.toThrow();
    });
  });

  // ==================== Sessions Tests ====================

  describe("getSessions", () => {
    it("should return array of sessions", async () => {
      const mockSessions = [
        { id: 1, device_info: "Chrome on Mac", is_current: true },
        { id: 2, device_info: "Firefox on Windows", is_current: false },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockSessions }),
      } as Response);

      const result = await SettingsService.getSessions();

      expect(result).toEqual(mockSessions);
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.getSessions()).rejects.toThrow();
    });
  });

  describe("revokeSession", () => {
    it("should revoke session successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.revokeSession(1)).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/sessions/1", {
        method: "DELETE",
      });
    });

    it("should throw error on revoke failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Session not found", "NOT_FOUND", 404));

      await expect(SettingsService.revokeSession(999)).rejects.toThrow();
    });
  });

  describe("revokeAllSessions", () => {
    it("should revoke all sessions successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.revokeAllSessions()).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/sessions", {
        method: "DELETE",
      });
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.revokeAllSessions()).rejects.toThrow();
    });
  });

  // ==================== Login History Tests ====================

  describe("getLoginHistory", () => {
    it("should return login history with default limit", async () => {
      const mockHistory = [
        { id: 1, ip_address: "192.168.1.1", success: true },
        { id: 2, ip_address: "192.168.1.2", success: false },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ history: mockHistory }),
      } as Response);

      const result = await SettingsService.getLoginHistory();

      expect(result).toEqual(mockHistory);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/login-history?limit=20");
    });

    it("should return login history with custom limit", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ history: [] }),
      } as Response);

      await SettingsService.getLoginHistory(50);

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/login-history?limit=50");
    });

    it("should return empty array when no history", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ history: null }),
      } as Response);

      const result = await SettingsService.getLoginHistory();

      expect(result).toEqual([]);
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.getLoginHistory()).rejects.toThrow();
    });
  });

  // ==================== Password Tests ====================

  describe("changePassword", () => {
    it("should change password successfully", async () => {
      const request: ChangePasswordRequest = {
        current_password: "oldpass",
        new_password: "newpass123",
      };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.changePassword(request)).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/password", {
        method: "PUT",
        body: JSON.stringify(request),
      });
    });

    it("should throw error on invalid current password", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid password", "INVALID_PASSWORD", 400));

      await expect(
        SettingsService.changePassword({ current_password: "wrong", new_password: "new" })
      ).rejects.toThrow();
    });
  });

  // ==================== 2FA Tests ====================

  describe("setup2FA", () => {
    it("should setup 2FA and return QR code", async () => {
      const mockSetup = {
        secret: "ABCD1234",
        qr_code: "data:image/png;base64,...",
        backup_codes: ["12345678", "87654321"],
      };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockSetup }),
      } as Response);

      const result = await SettingsService.setup2FA();

      expect(result).toEqual(mockSetup);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/2fa/setup", {
        method: "POST",
      });
    });

    it("should throw error on setup failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("2FA already enabled", "ALREADY_ENABLED", 400));

      await expect(SettingsService.setup2FA()).rejects.toThrow();
    });
  });

  describe("verify2FA", () => {
    it("should verify 2FA code and return backup codes", async () => {
      const mockResponse = { backup_codes: ["12345678", "87654321"] };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await SettingsService.verify2FA("123456");

      expect(result).toEqual(mockResponse);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/2fa/verify", {
        method: "POST",
        body: JSON.stringify({ code: "123456" }),
      });
    });

    it("should throw error on invalid code", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid code", "INVALID_CODE", 400));

      await expect(SettingsService.verify2FA("000000")).rejects.toThrow();
    });
  });

  describe("disable2FA", () => {
    it("should disable 2FA successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.disable2FA("123456")).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/2fa/disable", {
        method: "POST",
        body: JSON.stringify({ code: "123456" }),
      });
    });

    it("should throw error on invalid code", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid code", "INVALID_CODE", 400));

      await expect(SettingsService.disable2FA("000000")).rejects.toThrow();
    });
  });

  describe("regenerateBackupCodes", () => {
    it("should regenerate backup codes successfully", async () => {
      const mockResponse = { backup_codes: ["11111111", "22222222"] };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await SettingsService.regenerateBackupCodes("123456");

      expect(result).toEqual(mockResponse);
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid code", "INVALID_CODE", 400));

      await expect(SettingsService.regenerateBackupCodes("000000")).rejects.toThrow();
    });
  });

  // ==================== Profile Tests ====================

  describe("updateProfile", () => {
    it("should update profile successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.updateProfile({ name: "New Name" })).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me", {
        method: "PUT",
        body: JSON.stringify({ name: "New Name" }),
      });
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid profile", "INVALID", 400));

      await expect(SettingsService.updateProfile({ email: "invalid" })).rejects.toThrow();
    });
  });

  describe("uploadAvatar", () => {
    it("should upload avatar successfully", async () => {
      const mockFile = new File(["test"], "avatar.jpg", { type: "image/jpeg" });
      const mockResponse = { avatar_url: "https://example.com/avatar.jpg" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await SettingsService.uploadAvatar(mockFile);

      expect(result).toEqual(mockResponse);
      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/users/me/avatar",
        expect.objectContaining({
          method: "POST",
          headers: {},
        })
      );
    });

    it("should throw error on upload failure", async () => {
      const mockFile = new File(["test"], "avatar.jpg", { type: "image/jpeg" });

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid file", "INVALID_FILE", 400));

      await expect(SettingsService.uploadAvatar(mockFile)).rejects.toThrow();
    });
  });

  describe("deleteAvatar", () => {
    it("should delete avatar successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.deleteAvatar()).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/avatar", {
        method: "DELETE",
      });
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.deleteAvatar()).rejects.toThrow();
    });
  });

  // ==================== Data Export Tests ====================

  describe("requestDataExport", () => {
    it("should request data export successfully", async () => {
      const mockStatus = { id: 1, status: "pending" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockStatus }),
      } as Response);

      const result = await SettingsService.requestDataExport();

      expect(result).toEqual(mockStatus);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/export", {
        method: "POST",
      });
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 429,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Rate limited", "RATE_LIMITED", 429));

      await expect(SettingsService.requestDataExport()).rejects.toThrow();
    });
  });

  describe("getDataExportStatus", () => {
    it("should return export status", async () => {
      const mockStatus = { id: 1, status: "completed", download_url: "https://example.com/export.zip" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockStatus }),
      } as Response);

      const result = await SettingsService.getDataExportStatus();

      expect(result).toEqual(mockStatus);
    });

    it("should return null when no export exists", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);

      const result = await SettingsService.getDataExportStatus();

      expect(result).toBeNull();
    });

    it("should throw error on other failures", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.getDataExportStatus()).rejects.toThrow();
    });
  });

  // ==================== Account Deletion Tests ====================

  describe("requestAccountDeletion", () => {
    it("should request account deletion successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.requestAccountDeletion("password123")).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/delete", {
        method: "POST",
        body: JSON.stringify({ password: "password123", reason: undefined }),
      });
    });

    it("should include reason when provided", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await SettingsService.requestAccountDeletion("password123", "Moving to competitor");

      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/delete", {
        method: "POST",
        body: JSON.stringify({ password: "password123", reason: "Moving to competitor" }),
      });
    });

    it("should throw error on invalid password", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 401,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Invalid password", "INVALID_PASSWORD", 401));

      await expect(SettingsService.requestAccountDeletion("wrongpass")).rejects.toThrow();
    });
  });

  describe("cancelAccountDeletion", () => {
    it("should cancel account deletion successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.cancelAccountDeletion()).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/me/delete", {
        method: "DELETE",
      });
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(
        new ApiError("No pending deletion", "NO_PENDING_DELETION", 404)
      );

      await expect(SettingsService.cancelAccountDeletion()).rejects.toThrow();
    });
  });

  // ==================== Connected Accounts Tests ====================

  describe("getConnectedAccounts", () => {
    it("should return connected accounts", async () => {
      const mockAccounts = [
        { provider: "google", provider_user_id: "123", email: "test@gmail.com" },
        { provider: "github", provider_user_id: "456", email: "test@github.com" },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockAccounts }),
      } as Response);

      const result = await SettingsService.getConnectedAccounts();

      expect(result).toEqual(mockAccounts);
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(SettingsService.getConnectedAccounts()).rejects.toThrow();
    });
  });

  describe("disconnectAccount", () => {
    it("should disconnect account successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(SettingsService.disconnectAccount("google")).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/users/me/connected-accounts/google",
        { method: "DELETE" }
      );
    });

    it("should throw error on failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(
        new ApiError("Cannot disconnect last auth method", "LAST_AUTH_METHOD", 400)
      );

      await expect(SettingsService.disconnectAccount("github")).rejects.toThrow();
    });
  });

  // ==================== Static Data Tests ====================

  describe("Static data", () => {
    it("should export TIMEZONES array", () => {
      expect(TIMEZONES).toBeDefined();
      expect(Array.isArray(TIMEZONES)).toBe(true);
      expect(TIMEZONES).toContain("UTC");
      expect(TIMEZONES).toContain("America/New_York");
    });

    it("should export LANGUAGES array", () => {
      expect(LANGUAGES).toBeDefined();
      expect(Array.isArray(LANGUAGES)).toBe(true);
      expect(LANGUAGES).toContainEqual({ code: "en", name: "English" });
    });

    it("should export DATE_FORMATS array", () => {
      expect(DATE_FORMATS).toBeDefined();
      expect(Array.isArray(DATE_FORMATS)).toBe(true);
      expect(DATE_FORMATS[0]).toHaveProperty("value");
      expect(DATE_FORMATS[0]).toHaveProperty("label");
    });
  });
});
