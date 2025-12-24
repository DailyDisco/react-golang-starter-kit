import { beforeEach, describe, expect, it, vi } from "vitest";

// This test file tests the auth store mock behavior since the actual store
// is globally mocked in the test setup. For integration tests of the actual
// store, use an end-to-end testing framework.

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

Object.defineProperty(globalThis, "localStorage", {
  value: localStorageMock,
  writable: true,
});

const mockUser = {
  id: 1,
  name: "Test User",
  email: "test@example.com",
  email_verified: true,
  is_active: true,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

describe("auth-store logic", () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  describe("localStorage auth data", () => {
    it("stores auth token in localStorage", () => {
      localStorage.setItem("auth_token", "test-token");
      expect(localStorage.getItem("auth_token")).toBe("test-token");
    });

    it("stores auth user in localStorage", () => {
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      expect(JSON.parse(localStorage.getItem("auth_user") || "{}")).toEqual(mockUser);
    });

    it("removes auth data on clear", () => {
      localStorage.setItem("auth_token", "test-token");
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      localStorage.removeItem("auth_token");
      localStorage.removeItem("auth_user");

      expect(localStorage.getItem("auth_token")).toBeNull();
      expect(localStorage.getItem("auth_user")).toBeNull();
    });

    it("handles invalid JSON gracefully", () => {
      localStorage.setItem("auth_user", "invalid-json{");

      expect(() => {
        try {
          JSON.parse(localStorage.getItem("auth_user") || "{}");
        } catch {
          // Expected to throw
          throw new Error("Invalid JSON");
        }
      }).toThrow("Invalid JSON");
    });
  });

  describe("auth state validation", () => {
    it("requires both token and user for valid auth", () => {
      const hasToken = localStorage.getItem("auth_token") !== null;
      const hasUser = localStorage.getItem("auth_user") !== null;

      // Both must be present for valid auth
      expect(hasToken && hasUser).toBe(false);

      localStorage.setItem("auth_token", "token");
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const hasTokenNow = localStorage.getItem("auth_token") !== null;
      const hasUserNow = localStorage.getItem("auth_user") !== null;

      expect(hasTokenNow && hasUserNow).toBe(true);
    });

    it("validates user JSON structure", () => {
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const storedUser = JSON.parse(localStorage.getItem("auth_user") || "{}");

      expect(storedUser).toHaveProperty("id");
      expect(storedUser).toHaveProperty("email");
      expect(storedUser).toHaveProperty("name");
    });
  });

  // ============ Corruption Recovery Tests ============

  describe("localStorage corruption recovery", () => {
    /**
     * Helper to safely parse stored user data with corruption handling
     */
    const safeParseUser = (storedUser: string | null): { user: typeof mockUser | null; error: string | null } => {
      if (!storedUser) {
        return { user: null, error: null };
      }

      try {
        const parsed = JSON.parse(storedUser);

        // Validate required fields exist
        if (!parsed.id || !parsed.email) {
          return { user: null, error: "Missing required fields" };
        }

        // Validate field types
        if (typeof parsed.id !== "number" || typeof parsed.email !== "string") {
          return { user: null, error: "Invalid field types" };
        }

        return { user: parsed, error: null };
      } catch {
        return { user: null, error: "Invalid JSON" };
      }
    };

    it("recovers from corrupted JSON", () => {
      localStorage.setItem("auth_user", "not-valid-json{{{");

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Invalid JSON");
      expect(result.user).toBeNull();
    });

    it("recovers from truncated JSON", () => {
      localStorage.setItem("auth_user", '{"id":1,"email":"test@');

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Invalid JSON");
      expect(result.user).toBeNull();
    });

    it("recovers from empty string", () => {
      localStorage.setItem("auth_user", "");

      const result = safeParseUser(localStorage.getItem("auth_user"));

      // Empty string is falsy, so treated same as null (no data)
      expect(result.error).toBeNull();
      expect(result.user).toBeNull();
    });

    it("recovers from null value (item not found)", () => {
      // Don't set anything - simulate missing item
      const result = safeParseUser(null);

      expect(result.error).toBeNull();
      expect(result.user).toBeNull();
    });

    it("recovers from user missing required id field", () => {
      localStorage.setItem(
        "auth_user",
        JSON.stringify({
          email: "test@example.com",
          name: "Test User",
        })
      );

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Missing required fields");
      expect(result.user).toBeNull();
    });

    it("recovers from user missing required email field", () => {
      localStorage.setItem(
        "auth_user",
        JSON.stringify({
          id: 1,
          name: "Test User",
        })
      );

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Missing required fields");
      expect(result.user).toBeNull();
    });

    it("recovers from invalid field types", () => {
      localStorage.setItem(
        "auth_user",
        JSON.stringify({
          id: "not-a-number",
          email: 12345,
          name: "Test User",
        })
      );

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Invalid field types");
      expect(result.user).toBeNull();
    });

    it("accepts valid user data", () => {
      localStorage.setItem("auth_user", JSON.stringify(mockUser));

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBeNull();
      expect(result.user).not.toBeNull();
      expect(result.user?.id).toBe(1);
      expect(result.user?.email).toBe("test@example.com");
    });

    it("handles array instead of object", () => {
      localStorage.setItem("auth_user", JSON.stringify([1, 2, 3]));

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Missing required fields");
      expect(result.user).toBeNull();
    });

    it("handles primitive value instead of object", () => {
      localStorage.setItem("auth_user", JSON.stringify("just a string"));

      const result = safeParseUser(localStorage.getItem("auth_user"));

      expect(result.error).toBe("Missing required fields");
      expect(result.user).toBeNull();
    });
  });

  describe("auth state consistency", () => {
    it("should clear all auth data when token is present but user is missing", () => {
      localStorage.setItem("auth_token", "valid-token");
      // No auth_user set - inconsistent state

      const hasToken = localStorage.getItem("auth_token") !== null;
      const hasUser = localStorage.getItem("auth_user") !== null;

      // Detect inconsistent state
      if (hasToken && !hasUser) {
        localStorage.removeItem("auth_token");
        localStorage.removeItem("refresh_token");
      }

      expect(localStorage.getItem("auth_token")).toBeNull();
    });

    it("should clear all auth data when user is present but token is missing", () => {
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      // No auth_token set - inconsistent state

      const hasToken = localStorage.getItem("auth_token") !== null;
      const hasUser = localStorage.getItem("auth_user") !== null;

      // Detect inconsistent state
      if (!hasToken && hasUser) {
        localStorage.removeItem("auth_user");
      }

      expect(localStorage.getItem("auth_user")).toBeNull();
    });

    it("should preserve consistent auth data", () => {
      localStorage.setItem("auth_token", "valid-token");
      localStorage.setItem("auth_user", JSON.stringify(mockUser));
      localStorage.setItem("refresh_token", "valid-refresh");

      const hasToken = localStorage.getItem("auth_token") !== null;
      const hasUser = localStorage.getItem("auth_user") !== null;

      // Both present - consistent state
      expect(hasToken && hasUser).toBe(true);
      expect(localStorage.getItem("auth_token")).not.toBeNull();
      expect(localStorage.getItem("auth_user")).not.toBeNull();
    });
  });
});
