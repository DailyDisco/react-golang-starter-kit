import { beforeEach, describe, expect, it, vi } from "vitest";

import type { User } from "../services/types";
import { requireAdmin, requireAuth, requireRole, requireSuperAdmin } from "./guards";

// Mock @tanstack/react-router redirect
vi.mock("@tanstack/react-router", () => ({
  redirect: vi.fn((params) => {
    const error = new Error("REDIRECT");
    (error as Error & { to: string; search?: unknown }).to = params.to;
    (error as Error & { to: string; search?: unknown }).search = params.search;
    throw error;
  }),
}));

const createMockUser = (overrides?: Partial<User>): User => ({
  id: 1,
  name: "Test User",
  email: "test@example.com",
  email_verified: true,
  is_active: true,
  role: "user",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  ...overrides,
});

describe("guards", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("requireAuth", () => {
    it("returns user when authenticated", () => {
      const mockUser = createMockUser();
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = requireAuth();

      expect(result.user).toEqual(mockUser);
    });

    it("redirects to /login when not authenticated", () => {
      expect(() => requireAuth()).toThrow("REDIRECT");

      try {
        requireAuth();
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/login");
      }
    });

    it("includes redirect path in search params when provided", () => {
      try {
        requireAuth("/dashboard");
      } catch (error) {
        expect((error as Error & { to: string; search?: { redirect: string } }).to).toBe("/login");
        expect((error as Error & { to: string; search?: { redirect: string } }).search).toEqual({
          redirect: "/dashboard",
        });
      }
    });

    it("clears storage and redirects when user data is invalid JSON", () => {
      localStorage.setItem("auth_user", "invalid-json");

      expect(() => requireAuth()).toThrow("REDIRECT");
      expect(localStorage.getItem("auth_user")).toBeNull();
    });
  });

  describe("requireRole", () => {
    it("returns user when user has allowed role", () => {
      const mockUser = createMockUser({ role: "admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = requireRole(["admin", "super_admin"]);

      expect(result.user).toEqual(mockUser);
    });

    it("redirects to /login when not authenticated", () => {
      try {
        requireRole(["admin"]);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/login");
      }
    });

    it("redirects to / when authenticated but not authorized", () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireRole(["admin", "super_admin"]);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });

    it("respects custom unauthorizedRedirect", () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireRole(["admin"], { unauthorizedRedirect: "/forbidden" });
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/forbidden");
      }
    });

    it("handles user with no role", () => {
      const mockUser = createMockUser();
      delete (mockUser as Partial<User>).role;
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireRole(["admin"]);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });
  });

  describe("requireAdmin", () => {
    it("allows admin role", () => {
      const mockUser = createMockUser({ role: "admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = requireAdmin();

      expect(result.user.role).toBe("admin");
    });

    it("allows super_admin role", () => {
      const mockUser = createMockUser({ role: "super_admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = requireAdmin();

      expect(result.user.role).toBe("super_admin");
    });

    it("redirects regular user to /", () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireAdmin();
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });

    it("redirects premium user to /", () => {
      const mockUser = createMockUser({ role: "premium" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireAdmin();
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });
  });

  describe("requireSuperAdmin", () => {
    it("allows super_admin role", () => {
      const mockUser = createMockUser({ role: "super_admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = requireSuperAdmin();

      expect(result.user.role).toBe("super_admin");
    });

    it("redirects admin to /admin", () => {
      const mockUser = createMockUser({ role: "admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireSuperAdmin();
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/admin");
      }
    });

    it("redirects regular user to /admin", () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      try {
        requireSuperAdmin();
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/admin");
      }
    });
  });
});
