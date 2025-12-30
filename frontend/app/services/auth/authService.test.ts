import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, apiFetch, authenticatedFetchWithParsing, parseErrorResponse } from "../api/client";
import { AuthService } from "./authService";

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
  apiFetch: vi.fn(),
  authenticatedFetchWithParsing: vi.fn(),
  createHeaders: vi.fn(() => ({ "Content-Type": "application/json" })),
  parseErrorResponse: vi.fn(),
}));

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
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
});

describe("AuthService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("login", () => {
    it("should successfully login with valid credentials", async () => {
      const mockAuthResponse = {
        user: { id: 1, name: "Test User", email: "test@example.com" },
        token: "jwt-token-123",
      };

      vi.mocked(apiFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAuthResponse,
      } as Response);

      const result = await AuthService.login({
        email: "test@example.com",
        password: "password123",
      });

      expect(result).toEqual(mockAuthResponse);
      expect(apiFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/auth/login",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ email: "test@example.com", password: "password123" }),
        })
      );
    });

    it("should throw error on login failure", async () => {
      vi.mocked(apiFetch).mockResolvedValueOnce({
        ok: false,
        status: 401,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(
        new ApiError("Invalid credentials", "INVALID_CREDENTIALS", 401)
      );

      await expect(
        AuthService.login({
          email: "test@example.com",
          password: "wrongpassword",
        })
      ).rejects.toThrow("Invalid credentials");
    });

    it("should throw error on invalid JSON response", async () => {
      vi.mocked(apiFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as unknown as Response);

      await expect(
        AuthService.login({
          email: "test@example.com",
          password: "password123",
        })
      ).rejects.toThrow("Invalid response format from server");
    });
  });

  describe("register", () => {
    it("should successfully register a new user", async () => {
      const mockAuthResponse = {
        user: { id: 1, name: "New User", email: "new@example.com" },
        token: "jwt-token-456",
      };

      vi.mocked(apiFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAuthResponse,
      } as Response);

      const result = await AuthService.register({
        name: "New User",
        email: "new@example.com",
        password: "SecurePass123!",
      });

      expect(result).toEqual(mockAuthResponse);
      expect(apiFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/auth/register",
        expect.objectContaining({
          method: "POST",
        })
      );
    });

    it("should throw error on registration failure", async () => {
      vi.mocked(apiFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Email already exists", "EMAIL_EXISTS", 400));

      await expect(
        AuthService.register({
          name: "Test User",
          email: "existing@example.com",
          password: "password123",
        })
      ).rejects.toThrow("Email already exists");
    });
  });

  describe("getCurrentUser", () => {
    it("should return current user data", async () => {
      const mockUser = { id: 1, name: "Test User", email: "test@example.com" };

      vi.mocked(authenticatedFetchWithParsing).mockResolvedValueOnce(mockUser);

      const result = await AuthService.getCurrentUser();

      expect(result).toEqual(mockUser);
      expect(authenticatedFetchWithParsing).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/me");
    });
  });

  describe("updateUser", () => {
    it("should update user profile", async () => {
      const updatedUser = { id: 1, name: "Updated Name", email: "test@example.com" };

      vi.mocked(authenticatedFetchWithParsing).mockResolvedValueOnce(updatedUser);

      const result = await AuthService.updateUser(1, { name: "Updated Name" });

      expect(result).toEqual(updatedUser);
      expect(authenticatedFetchWithParsing).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/users/1",
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify({ name: "Updated Name" }),
        })
      );
    });
  });

  describe("logout", () => {
    it("should call logout endpoint and clear localStorage", async () => {
      // Only auth_user is in localStorage (tokens are in httpOnly cookies)
      localStorageMock.setItem("auth_user", '{"id": 1}');

      vi.mocked(apiFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await AuthService.logout();

      expect(apiFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/logout", { method: "POST" });
      // Only auth_user is removed from localStorage (tokens are cleared via backend)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
    });

    it("should clear localStorage even if API call fails", async () => {
      // Only auth_user is in localStorage (tokens are in httpOnly cookies)
      localStorageMock.setItem("auth_user", '{"id": 1}');

      vi.mocked(apiFetch).mockRejectedValueOnce(new Error("Network error"));

      await AuthService.logout();

      // Only auth_user is removed from localStorage (tokens are cleared via backend)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
    });
  });

  describe("isAuthenticated", () => {
    it("should return true when user session is valid", async () => {
      vi.mocked(authenticatedFetchWithParsing).mockResolvedValueOnce({ id: 1, name: "User" });

      const result = await AuthService.isAuthenticated();

      expect(result).toBe(true);
    });

    it("should return false and clear storage when session is invalid", async () => {
      // auth_token is in httpOnly cookie, only auth_user is in localStorage
      localStorageMock.setItem("auth_user", '{"id": 1}');

      vi.mocked(authenticatedFetchWithParsing).mockRejectedValueOnce(new Error("Unauthorized"));

      const result = await AuthService.isAuthenticated();

      expect(result).toBe(false);
      // Only auth_user is in localStorage (tokens are in httpOnly cookies)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
    });
  });

  describe("storeAuthData", () => {
    it("should store user data in localStorage", () => {
      const authData = {
        user: { id: 1, name: "Test User", email: "test@example.com" },
        token: "jwt-token", // Note: token is in httpOnly cookie, not stored in localStorage
      };

      AuthService.storeAuthData(authData as any);

      // Only minimal user data is stored (tokens are in httpOnly cookies)
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        "auth_user",
        JSON.stringify({ id: 1, name: "Test User", email: "test@example.com" })
      );
    });

    it("should throw error when user data is missing", () => {
      const authData = { token: "jwt-token" };

      expect(() => AuthService.storeAuthData(authData as any)).toThrow("Invalid authentication data");
    });

    it("should throw error when user data is not serializable", () => {
      const circularObj: any = { id: 1 };
      circularObj.self = circularObj;

      const authData = {
        user: circularObj,
        token: "jwt-token",
      };

      // The error might be thrown either by our validation or by JSON.stringify in the logger
      // Either way, circular structures should cause an error
      expect(() => AuthService.storeAuthData(authData as any)).toThrow();
    });
  });

  describe("clearStorage", () => {
    it("should remove auth data from localStorage", () => {
      // Only auth_user is in localStorage (tokens are in httpOnly cookies)
      localStorageMock.setItem("auth_user", '{"id": 1}');

      AuthService.clearStorage();

      // Only auth_user is removed (tokens are cleared via backend logout endpoint)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
    });
  });
});
