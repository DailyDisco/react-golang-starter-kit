import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError, authenticatedFetch, authenticatedFetchWithParsing, parseErrorResponse } from "../api/client";
import { UserService } from "./userService";

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
  authenticatedFetchWithParsing: vi.fn(),
  parseErrorResponse: vi.fn(),
}));

describe("UserService", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("fetchUsers", () => {
    it("should return array of users on success", async () => {
      const mockUsers = [
        { id: 1, name: "User 1", email: "user1@example.com" },
        { id: 2, name: "User 2", email: "user2@example.com" },
      ];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: { users: mockUsers, count: 2 } }),
      } as Response);

      const result = await UserService.fetchUsers();

      expect(result).toEqual(mockUsers);
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users");
    });

    it("should return empty array when no users exist", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: { users: null, count: 0 } }),
      } as Response);

      const result = await UserService.fetchUsers();

      expect(result).toEqual([]);
    });

    it("should handle old response format (fallback)", async () => {
      const mockUsers = [{ id: 1, name: "User 1", email: "user1@example.com" }];

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ users: mockUsers }),
      } as Response);

      const result = await UserService.fetchUsers();

      expect(result).toEqual(mockUsers);
    });

    it("should throw error on API failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Server error", "SERVER_ERROR", 500));

      await expect(UserService.fetchUsers()).rejects.toThrow("Server error");
    });

    it("should throw error on invalid JSON response", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as unknown as Response);

      await expect(UserService.fetchUsers()).rejects.toThrow("Invalid response format from server");
    });
  });

  describe("createUser", () => {
    it("should create user with name and email only", async () => {
      const newUser = { id: 1, name: "New User", email: "new@example.com" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: newUser }),
      } as Response);

      const result = await UserService.createUser("New User", "new@example.com");

      expect(result).toEqual(newUser);
      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/users",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ name: "New User", email: "new@example.com" }),
        })
      );
    });

    it("should create user with password when provided", async () => {
      const newUser = { id: 1, name: "New User", email: "new@example.com" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: newUser }),
      } as Response);

      const result = await UserService.createUser("New User", "new@example.com", "password123");

      expect(result).toEqual(newUser);
      expect(authenticatedFetch).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/users",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ name: "New User", email: "new@example.com", password: "password123" }),
        })
      );
    });

    it("should handle old response format (fallback)", async () => {
      const newUser = { id: 1, name: "New User", email: "new@example.com" };

      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => newUser,
      } as Response);

      const result = await UserService.createUser("New User", "new@example.com");

      expect(result).toEqual(newUser);
    });

    it("should throw error on API failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 400,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("Email already exists", "EMAIL_EXISTS", 400));

      await expect(UserService.createUser("Test", "existing@example.com")).rejects.toThrow("Email already exists");
    });

    it("should throw error on invalid JSON response", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error("Invalid JSON");
        },
      } as unknown as Response);

      await expect(UserService.createUser("Test", "test@example.com")).rejects.toThrow(
        "Invalid response format from server"
      );
    });
  });

  describe("updateUser", () => {
    it("should update user and return updated data", async () => {
      const user = { id: 1, name: "Updated Name", email: "test@example.com" };

      vi.mocked(authenticatedFetchWithParsing).mockResolvedValueOnce(user);

      const result = await UserService.updateUser(user as any);

      expect(result).toEqual(user);
      expect(authenticatedFetchWithParsing).toHaveBeenCalledWith(
        "http://localhost:8080/api/v1/users/1",
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify(user),
        })
      );
    });

    it("should propagate errors from authenticatedFetchWithParsing", async () => {
      const user = { id: 1, name: "Test", email: "test@example.com" };

      vi.mocked(authenticatedFetchWithParsing).mockRejectedValueOnce(new Error("Update failed"));

      await expect(UserService.updateUser(user as any)).rejects.toThrow("Update failed");
    });
  });

  describe("deleteUser", () => {
    it("should delete user successfully", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: true,
      } as Response);

      await expect(UserService.deleteUser(1)).resolves.toBeUndefined();
      expect(authenticatedFetch).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/1", { method: "DELETE" });
    });

    it("should throw error on API failure", async () => {
      vi.mocked(authenticatedFetch).mockResolvedValueOnce({
        ok: false,
        status: 404,
      } as Response);
      vi.mocked(parseErrorResponse).mockResolvedValueOnce(new ApiError("User not found", "NOT_FOUND", 404));

      await expect(UserService.deleteUser(999)).rejects.toThrow("User not found");
    });
  });

  describe("getUserById", () => {
    it("should return user by ID", async () => {
      const user = { id: 1, name: "Test User", email: "test@example.com" };

      vi.mocked(authenticatedFetchWithParsing).mockResolvedValueOnce(user);

      const result = await UserService.getUserById(1);

      expect(result).toEqual(user);
      expect(authenticatedFetchWithParsing).toHaveBeenCalledWith("http://localhost:8080/api/v1/users/1");
    });

    it("should propagate errors from authenticatedFetchWithParsing", async () => {
      vi.mocked(authenticatedFetchWithParsing).mockRejectedValueOnce(new Error("User not found"));

      await expect(UserService.getUserById(999)).rejects.toThrow("User not found");
    });
  });
});
