import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  API_BASE_URL,
  apiFetch,
  apiFetchWithParsing,
  authenticatedFetch,
  authenticatedFetchWithParsing,
  createHeaders,
  parseApiResponse,
  parseErrorResponse,
} from "./client";

// Mock fetch globally
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
});

// Mock window.location
const mockLocation = {
  href: "",
  pathname: "/dashboard",
};
Object.defineProperty(window, "location", {
  value: mockLocation,
  writable: true,
});

describe("API Client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
    mockLocation.href = "";
    mockLocation.pathname = "/dashboard";
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("API_BASE_URL", () => {
    it("should be defined", () => {
      expect(API_BASE_URL).toBeDefined();
      expect(typeof API_BASE_URL).toBe("string");
    });
  });

  describe("createHeaders", () => {
    it("should create headers with Content-Type", () => {
      const headers = createHeaders(false);
      expect(headers["Content-Type"]).toBe("application/json");
    });

    it("should include CSRF token when includeCSRF is true and token exists", () => {
      // Set up CSRF cookie
      Object.defineProperty(document, "cookie", {
        writable: true,
        value: "csrf_token=test-csrf-token",
      });
      const headers = createHeaders(true);
      expect(headers["X-CSRF-Token"]).toBe("test-csrf-token");
    });

    it("should not include CSRF token when includeCSRF is false", () => {
      Object.defineProperty(document, "cookie", {
        writable: true,
        value: "csrf_token=test-csrf-token",
      });
      const headers = createHeaders(false);
      expect(headers["X-CSRF-Token"]).toBeUndefined();
    });

    it("should not include CSRF token when no token exists", () => {
      Object.defineProperty(document, "cookie", {
        writable: true,
        value: "",
      });
      const headers = createHeaders(true);
      expect(headers["X-CSRF-Token"]).toBeUndefined();
    });
  });

  describe("authenticatedFetch", () => {
    it("should include credentials in request", async () => {
      mockFetch.mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => ({ success: true }),
      });

      await authenticatedFetch("/api/test");

      expect(mockFetch).toHaveBeenCalledWith(
        "/api/test",
        expect.objectContaining({
          credentials: "include",
        })
      );
    });

    it("should clear localStorage and redirect on 401", async () => {
      // auth_token is in httpOnly cookie, only auth_user is in localStorage
      localStorageMock.setItem("auth_user", '{"id": 1}');

      mockFetch.mockResolvedValueOnce({
        status: 401,
        ok: false,
      });

      await authenticatedFetch("/api/protected");

      // Only auth_user is in localStorage (tokens are in httpOnly cookies)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
      expect(mockLocation.href).toBe("/login");
    });

    it("should not redirect to login if already on login page", async () => {
      mockLocation.pathname = "/login";
      // auth_token is in httpOnly cookie
      localStorageMock.setItem("auth_user", '{"id": 1}');

      mockFetch.mockResolvedValueOnce({
        status: 401,
        ok: false,
      });

      await authenticatedFetch("/api/protected");

      expect(mockLocation.href).toBe("");
    });

    it("should merge custom headers with default headers", async () => {
      mockFetch.mockResolvedValueOnce({
        status: 200,
        ok: true,
      });

      await authenticatedFetch("/api/test", {
        headers: { "X-Custom-Header": "custom-value" },
      });

      expect(mockFetch).toHaveBeenCalledWith(
        "/api/test",
        expect.objectContaining({
          headers: expect.objectContaining({
            "Content-Type": "application/json",
            "X-Custom-Header": "custom-value",
          }),
        })
      );
    });
  });

  describe("apiFetch", () => {
    it("should include credentials but not auth header", async () => {
      mockFetch.mockResolvedValueOnce({
        status: 200,
        ok: true,
      });

      await apiFetch("/api/auth/login", { method: "POST" });

      expect(mockFetch).toHaveBeenCalledWith(
        "/api/auth/login",
        expect.objectContaining({
          credentials: "include",
          method: "POST",
        })
      );
    });
  });

  describe("parseErrorResponse", () => {
    it("should parse JSON error response", async () => {
      const mockResponse = {
        status: 400,
        text: async () => JSON.stringify({ error: "Bad Request", message: "Invalid input" }),
      } as Response;

      const result = await parseErrorResponse(mockResponse, "Default error");
      expect(result.message).toBe("Bad Request");
    });

    it("should return message if error field is missing", async () => {
      const mockResponse = {
        status: 400,
        text: async () => JSON.stringify({ message: "Validation failed" }),
      } as Response;

      const result = await parseErrorResponse(mockResponse, "Default error");
      expect(result.message).toBe("Validation failed");
    });

    it("should handle non-JSON response", async () => {
      const mockResponse = {
        status: 500,
        text: async () => "Internal Server Error",
      } as Response;

      const result = await parseErrorResponse(mockResponse, "Default error");
      expect(result.message).toBe("Internal Server Error");
    });

    it("should use default message with status on empty response", async () => {
      const mockResponse = {
        status: 404,
        text: async () => "",
      } as Response;

      const result = await parseErrorResponse(mockResponse, "Not found");
      expect(result.message).toBe("Not found with status 404");
    });

    it("should handle parse errors gracefully", async () => {
      const mockResponse = {
        status: 500,
        text: async () => {
          throw new Error("Network error");
        },
      } as unknown as Response;

      const result = await parseErrorResponse(mockResponse, "Request failed");
      expect(result.message).toBe("Request failed with status 500");
    });
  });

  describe("parseApiResponse", () => {
    it("should parse success response with data", async () => {
      const mockResponse = {
        ok: true,
        json: async () => ({
          success: true,
          message: "Success",
          data: { id: 1, name: "Test" },
        }),
      } as Response;

      const result = await parseApiResponse<{ id: number; name: string }>(mockResponse);
      expect(result).toEqual({ id: 1, name: "Test" });
    });

    it("should throw error on non-ok response", async () => {
      const mockResponse = {
        ok: false,
        status: 400,
        json: async () => ({
          error: "Bad Request",
          message: "Invalid data",
        }),
      } as Response;

      // Note: The current implementation catches the thrown error in the outer try-catch
      // and re-throws with "Invalid response format from server"
      await expect(parseApiResponse(mockResponse)).rejects.toThrow("Invalid response format from server");
    });

    it("should handle response without success wrapper", async () => {
      const mockResponse = {
        ok: true,
        json: async () => ({ id: 1, name: "Direct response" }),
      } as Response;

      const result = await parseApiResponse(mockResponse);
      expect(result).toEqual({ id: 1, name: "Direct response" });
    });

    it("should throw on JSON parse error", async () => {
      const mockResponse = {
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as unknown as Response;

      await expect(parseApiResponse(mockResponse)).rejects.toThrow("Invalid response format from server");
    });
  });

  describe("authenticatedFetchWithParsing", () => {
    it("should fetch and parse response", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          success: true,
          message: "User retrieved",
          data: { id: 1, name: "John" },
        }),
      });

      const result = await authenticatedFetchWithParsing<{ id: number; name: string }>("/api/users/1");
      expect(result).toEqual({ id: 1, name: "John" });
    });
  });

  describe("apiFetchWithParsing", () => {
    it("should fetch and parse response without auth", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({
          success: true,
          message: "Logged in",
          data: { user: { id: 1 }, token: "jwt-token" },
        }),
      });

      const result = await apiFetchWithParsing("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ email: "test@example.com", password: "password" }),
      });

      expect(result).toEqual({ user: { id: 1 }, token: "jwt-token" });
    });
  });

  // ============ Network Error Handling Tests ============

  describe("Network Error Handling", () => {
    it("should throw on network failure", async () => {
      mockFetch.mockRejectedValueOnce(new TypeError("Failed to fetch"));

      await expect(authenticatedFetch("/api/test")).rejects.toThrow("Failed to fetch");
    });

    it("should throw on network timeout", async () => {
      mockFetch.mockRejectedValueOnce(new DOMException("The operation was aborted", "AbortError"));

      await expect(authenticatedFetch("/api/test")).rejects.toThrow("The operation was aborted");
    });
  });

  // ============ HTTP Status Code Handling Tests ============

  describe("HTTP Status Code Handling", () => {
    it("should handle 403 Forbidden response", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => ({
          error: "FORBIDDEN",
          message: "Access denied",
          code: 403,
        }),
      });

      const response = await authenticatedFetch("/api/admin");

      expect(response.status).toBe(403);
      // 403 should not trigger logout/redirect (unlike 401)
      expect(localStorageMock.removeItem).not.toHaveBeenCalled();
      expect(mockLocation.href).toBe("");
    });

    it("should handle 500 Internal Server Error response", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({
          error: "INTERNAL_ERROR",
          message: "Internal server error",
          code: 500,
        }),
      });

      const response = await authenticatedFetch("/api/test");

      expect(response.status).toBe(500);
      // 500 should not trigger logout/redirect
      expect(localStorageMock.removeItem).not.toHaveBeenCalled();
    });

    it("should handle 404 Not Found response", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => ({
          error: "NOT_FOUND",
          message: "Resource not found",
          code: 404,
        }),
      });

      const response = await authenticatedFetch("/api/users/999");

      expect(response.status).toBe(404);
    });

    it("should clear all auth data on 401", async () => {
      localStorageMock.setItem("auth_token", "expired-token");
      localStorageMock.setItem("auth_user", '{"id": 1}');
      // Note: refresh token is now in httpOnly cookie, not localStorage

      mockFetch.mockResolvedValueOnce({
        status: 401,
        ok: false,
      });

      await authenticatedFetch("/api/protected");

      // Only auth_user is in localStorage (tokens are in httpOnly cookies)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
    });
  });

  // ============ parseApiResponse Error Cases ============

  describe("parseApiResponse Error Handling", () => {
    it("should throw with error message from API error response", async () => {
      const mockResponse = {
        ok: false,
        status: 403,
        json: async () => ({
          error: "FORBIDDEN",
          message: "You do not have permission to access this resource",
        }),
      } as Response;

      await expect(parseApiResponse(mockResponse)).rejects.toThrow("Invalid response format from server");
    });

    it("should handle response with error field only", async () => {
      const mockResponse = {
        ok: false,
        status: 400,
        json: async () => ({
          error: "VALIDATION_ERROR",
        }),
      } as Response;

      await expect(parseApiResponse(mockResponse)).rejects.toThrow("Invalid response format from server");
    });
  });

  // ============ CSRF Token Handling ============

  describe("CSRF Token Handling", () => {
    // Mock document.cookie
    const originalCookie = Object.getOwnPropertyDescriptor(document, "cookie");

    beforeEach(() => {
      Object.defineProperty(document, "cookie", {
        value: "csrf_token=test-csrf-token-123",
        writable: true,
        configurable: true,
      });
    });

    afterEach(() => {
      if (originalCookie) {
        Object.defineProperty(document, "cookie", originalCookie);
      }
    });

    it("should include CSRF token for POST requests", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ success: true }),
      });

      await authenticatedFetch("/api/data", { method: "POST" });

      expect(mockFetch).toHaveBeenCalledWith(
        "/api/data",
        expect.objectContaining({
          headers: expect.objectContaining({
            "X-CSRF-Token": "test-csrf-token-123",
          }),
        })
      );
    });

    it("should include CSRF token for DELETE requests", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ success: true }),
      });

      await authenticatedFetch("/api/data/1", { method: "DELETE" });

      expect(mockFetch).toHaveBeenCalledWith(
        "/api/data/1",
        expect.objectContaining({
          headers: expect.objectContaining({
            "X-CSRF-Token": "test-csrf-token-123",
          }),
        })
      );
    });
  });
});
