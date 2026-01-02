import { QueryClient } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { User } from "../services/types";
import { requireAdmin, requireAuth, requireAuthSync, requireRole, requireSuperAdmin } from "./guards";

// Mock @tanstack/react-router redirect
vi.mock("@tanstack/react-router", () => ({
  redirect: vi.fn((params) => {
    const error = new Error("REDIRECT");
    (error as Error & { to: string; search?: unknown; code?: string }).to = params.to;
    (error as Error & { to: string; search?: unknown; code?: string }).search = params.search;
    (error as Error & { to: string; search?: unknown; code?: string }).code = "REDIRECT";
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

const createMockContext = (user?: User | null) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  // If a user is provided, set up the query client to return it
  if (user) {
    queryClient.setQueryData(["auth", "user"], user);
  }

  return {
    context: {
      queryClient,
    },
    location: {
      pathname: "/test",
    },
  };
};

describe("guards", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("requireAuth", () => {
    it("returns user when authenticated and verified", async () => {
      const mockUser = createMockUser();
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      const result = await requireAuth(ctx);

      expect(result.user).toEqual(mockUser);
    });

    it("redirects to /login when no stored user", async () => {
      const ctx = createMockContext();

      await expect(requireAuth(ctx)).rejects.toThrow("REDIRECT");

      try {
        await requireAuth(ctx);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/login");
      }
    });

    it("includes redirect path in search params when provided", async () => {
      const ctx = createMockContext();

      try {
        await requireAuth(ctx, "/dashboard");
      } catch (error) {
        expect((error as Error & { to: string; search?: { redirect: string } }).to).toBe("/login");
        expect((error as Error & { to: string; search?: { redirect: string } }).search).toEqual({
          redirect: "/dashboard",
        });
      }
    });

    it("clears storage and redirects when user data is invalid JSON", async () => {
      localStorage.setItem("auth_user", "invalid-json");
      const ctx = createMockContext();

      await expect(requireAuth(ctx)).rejects.toThrow("REDIRECT");
      expect(localStorage.getItem("auth_user")).toBeNull();
    });
  });

  describe("requireRole", () => {
    it("returns user when user has allowed role", async () => {
      const mockUser = createMockUser({ role: "admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      const result = await requireRole(ctx, ["admin", "super_admin"]);

      expect(result.user).toEqual(mockUser);
    });

    it("redirects to /login when not authenticated", async () => {
      const ctx = createMockContext();

      try {
        await requireRole(ctx, ["admin"]);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/login");
      }
    });

    it("redirects to / when authenticated but not authorized", async () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireRole(ctx, ["admin", "super_admin"]);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });

    it("respects custom unauthorizedRedirect", async () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireRole(ctx, ["admin"], { unauthorizedRedirect: "/forbidden" });
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/forbidden");
      }
    });

    it("handles user with no role", async () => {
      const mockUser = createMockUser();
      delete (mockUser as Partial<User>).role;
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireRole(ctx, ["admin"]);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });
  });

  describe("requireAdmin", () => {
    it("allows admin role", async () => {
      const mockUser = createMockUser({ role: "admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      const result = await requireAdmin(ctx);

      expect(result.user.role).toBe("admin");
    });

    it("allows super_admin role", async () => {
      const mockUser = createMockUser({ role: "super_admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      const result = await requireAdmin(ctx);

      expect(result.user.role).toBe("super_admin");
    });

    it("redirects regular user to /", async () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireAdmin(ctx);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });

    it("redirects premium user to /", async () => {
      const mockUser = createMockUser({ role: "premium" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireAdmin(ctx);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/");
      }
    });
  });

  describe("requireSuperAdmin", () => {
    it("allows super_admin role", async () => {
      const mockUser = createMockUser({ role: "super_admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      const result = await requireSuperAdmin(ctx);

      expect(result.user.role).toBe("super_admin");
    });

    it("redirects admin to /admin", async () => {
      const mockUser = createMockUser({ role: "admin" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireSuperAdmin(ctx);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/admin");
      }
    });

    it("redirects regular user to /admin", async () => {
      const mockUser = createMockUser({ role: "user" });
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      const ctx = createMockContext(mockUser);

      try {
        await requireSuperAdmin(ctx);
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/admin");
      }
    });
  });

  describe("requireAuthSync (deprecated)", () => {
    it("returns user when authenticated", () => {
      const mockUser = createMockUser();
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = requireAuthSync();

      expect(result.user).toEqual(mockUser);
    });

    it("redirects to /login when not authenticated", () => {
      expect(() => requireAuthSync()).toThrow("REDIRECT");

      try {
        requireAuthSync();
      } catch (error) {
        expect((error as Error & { to: string }).to).toBe("/login");
      }
    });

    it("includes redirect path in search params when provided", () => {
      try {
        requireAuthSync("/dashboard");
      } catch (error) {
        expect((error as Error & { to: string; search?: { redirect: string } }).to).toBe("/login");
        expect((error as Error & { to: string; search?: { redirect: string } }).search).toEqual({
          redirect: "/dashboard",
        });
      }
    });
  });
});
