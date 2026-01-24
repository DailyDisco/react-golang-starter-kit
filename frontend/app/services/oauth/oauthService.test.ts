import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { authenticatedFetch } from "../api/client";
import { OAuthService, type OAuthProvider } from "./oauthService";

// Mock the API client module
vi.mock("../api/client", () => ({
  API_BASE_URL: "http://localhost:8080",
  authenticatedFetch: vi.fn(),
}));

// Mock the logger
vi.mock("../../lib/logger", () => ({
  logger: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn(),
  },
}));

// Mock window.location
const mockLocation = {
  href: "",
};
Object.defineProperty(window, "location", {
  value: mockLocation,
  writable: true,
});

describe("OAuthService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockLocation.href = "";
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("getOAuthURL", () => {
    it("should return OAuth URL for Google", async () => {
      const mockResponse = {
        url: "https://accounts.google.com/oauth/authorize?...",
        state: "abc123",
      };

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await OAuthService.getOAuthURL("google");

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/oauth/google", {
        method: "GET",
        credentials: "include",
      });
    });

    it("should return OAuth URL for GitHub", async () => {
      const mockResponse = {
        url: "https://github.com/login/oauth/authorize?...",
        state: "xyz789",
      };

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      const result = await OAuthService.getOAuthURL("github");

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/oauth/github", {
        method: "GET",
        credentials: "include",
      });
    });

    it("should throw error on failed response", async () => {
      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: false,
        text: async () => "OAuth provider not configured",
      } as Response);

      await expect(OAuthService.getOAuthURL("google")).rejects.toThrow("Failed to initialize google login");
    });
  });

  describe("initiateOAuth", () => {
    it("should redirect to OAuth URL for Google", async () => {
      const mockResponse = {
        url: "https://accounts.google.com/oauth/authorize?client_id=123",
        state: "abc123",
      };

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      await OAuthService.initiateOAuth("google");

      expect(mockLocation.href).toBe("https://accounts.google.com/oauth/authorize?client_id=123");
    });

    it("should redirect to OAuth URL for GitHub", async () => {
      const mockResponse = {
        url: "https://github.com/login/oauth/authorize?client_id=456",
        state: "xyz789",
      };

      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

      await OAuthService.initiateOAuth("github");

      expect(mockLocation.href).toBe("https://github.com/login/oauth/authorize?client_id=456");
    });

    it("should throw error if getOAuthURL fails", async () => {
      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: false,
        text: async () => "Provider not available",
      } as Response);

      await expect(OAuthService.initiateOAuth("google")).rejects.toThrow("Failed to initialize google login");
    });
  });

  describe("getLinkedProviders", () => {
    it("should return list of linked providers", async () => {
      const mockProviders = [
        { provider: "google", email: "user@gmail.com", linked_at: "2024-01-01T00:00:00Z" },
        { provider: "github", email: "user@github.com", linked_at: "2024-01-02T00:00:00Z" },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ providers: mockProviders }),
      } as Response);

      const result = await OAuthService.getLinkedProviders();

      expect(result).toEqual(mockProviders);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/oauth/providers");
    });

    it("should return empty array when no providers linked", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ providers: null }),
      } as Response);

      const result = await OAuthService.getLinkedProviders();

      expect(result).toEqual([]);
    });

    it("should return empty array on 401 (unauthorized)", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 401,
      } as Response);

      const result = await OAuthService.getLinkedProviders();

      expect(result).toEqual([]);
    });

    it("should throw error on other failures", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);

      await expect(OAuthService.getLinkedProviders()).rejects.toThrow("Failed to get linked providers");
    });
  });

  describe("unlinkProvider", () => {
    it("should unlink Google provider successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(OAuthService.unlinkProvider("google")).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/oauth/google", {
        method: "DELETE",
      });
    });

    it("should unlink GitHub provider successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(OAuthService.unlinkProvider("github")).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/oauth/github", {
        method: "DELETE",
      });
    });

    it("should throw error on unlink failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        text: async () => "Cannot unlink last authentication method",
      } as Response);

      await expect(OAuthService.unlinkProvider("google")).rejects.toThrow("Failed to unlink google");
    });
  });

  describe("isProviderConfigured", () => {
    it("should return true when provider is configured", async () => {
      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ url: "https://...", state: "123" }),
      } as Response);

      const result = await OAuthService.isProviderConfigured("google");

      expect(result).toBe(true);
    });

    it("should return false when provider is not configured", async () => {
      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: false,
        text: async () => "Provider not configured",
      } as Response);

      const result = await OAuthService.isProviderConfigured("google");

      expect(result).toBe(false);
    });

    it("should return false on network error", async () => {
      vi.mocked(global.fetch).mockRejectedValueOnce(new Error("Network error"));

      const result = await OAuthService.isProviderConfigured("github");

      expect(result).toBe(false);
    });
  });

  describe("Provider types", () => {
    it("should accept google as provider", async () => {
      const provider: OAuthProvider = "google";
      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ url: "https://...", state: "123" }),
      } as Response);

      await expect(OAuthService.getOAuthURL(provider)).resolves.toBeDefined();
    });

    it("should accept github as provider", async () => {
      const provider: OAuthProvider = "github";
      vi.mocked(global.fetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ url: "https://...", state: "123" }),
      } as Response);

      await expect(OAuthService.getOAuthURL(provider)).resolves.toBeDefined();
    });
  });
});
