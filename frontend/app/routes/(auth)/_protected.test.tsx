import { beforeEach, describe, expect, it, vi } from "vitest";

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
    get store() {
      return store;
    },
  };
})();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
});

describe("Protected Route beforeLoad", () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  // Helper to simulate the beforeLoad logic
  const checkAuthentication = () => {
    const storedToken = localStorage.getItem("auth_token");
    const storedUser = localStorage.getItem("auth_user");

    let isAuthenticated = false;
    if (storedToken && storedUser) {
      try {
        JSON.parse(storedUser);
        isAuthenticated = true;
      } catch {
        localStorage.removeItem("auth_token");
        localStorage.removeItem("auth_user");
      }
    }

    return isAuthenticated;
  };

  it("returns false when no token exists", () => {
    const result = checkAuthentication();
    expect(result).toBe(false);
  });

  it("returns false when token exists but no user data", () => {
    localStorage.setItem("auth_token", "valid-token");

    const result = checkAuthentication();
    expect(result).toBe(false);
  });

  it("returns false when user data exists but no token", () => {
    localStorage.setItem("auth_user", JSON.stringify({ id: 1, name: "Test" }));

    const result = checkAuthentication();
    expect(result).toBe(false);
  });

  it("returns true when valid token and user data exist", () => {
    localStorage.setItem("auth_token", "valid-token");
    localStorage.setItem("auth_user", JSON.stringify({ id: 1, name: "Test", email: "test@example.com" }));

    const result = checkAuthentication();
    expect(result).toBe(true);
  });

  it("clears storage and returns false when user data is invalid JSON", () => {
    localStorage.setItem("auth_token", "valid-token");
    localStorage.setItem("auth_user", "invalid-json{");

    const result = checkAuthentication();

    expect(result).toBe(false);
    expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_token");
    expect(localStorageMock.removeItem).toHaveBeenCalledWith("auth_user");
  });

  it("does not clear storage when data is valid", () => {
    localStorage.setItem("auth_token", "valid-token");
    localStorage.setItem("auth_user", JSON.stringify({ id: 1 }));

    checkAuthentication();

    expect(localStorageMock.removeItem).not.toHaveBeenCalled();
  });
});
