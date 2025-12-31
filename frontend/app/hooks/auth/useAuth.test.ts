import { beforeEach, describe, expect, it, vi } from "vitest";

import { useAuth } from "./useAuth";

// Mock dependencies
const mockLoginMutate = vi.fn();
const mockRegisterMutate = vi.fn();
const mockUpdateUserMutate = vi.fn();
const mockClearFeatureFlags = vi.fn();
const mockStoreLogout = vi.fn();
const mockSetUser = vi.fn();

vi.mock("../../services/auth/authService", () => ({
  AuthService: {
    getCurrentUser: vi.fn(),
  },
}));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(() => ({
    user: null,
    isLoading: false,
    logout: mockStoreLogout,
    setUser: mockSetUser,
  })),
}));

vi.mock("../mutations/use-auth-mutations", () => ({
  useLogin: vi.fn(() => ({
    mutate: mockLoginMutate,
    isPending: false,
  })),
  useRegister: vi.fn(() => ({
    mutate: mockRegisterMutate,
    isPending: false,
  })),
}));

vi.mock("../mutations/use-user-mutations", () => ({
  useUpdateUser: vi.fn(() => ({
    mutate: mockUpdateUserMutate,
    isPending: false,
  })),
}));

vi.mock("../queries/use-feature-flags", () => ({
  useClearFeatureFlags: vi.fn(() => mockClearFeatureFlags),
}));

describe("useAuth", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("is defined as a function", () => {
    expect(useAuth).toBeDefined();
    expect(typeof useAuth).toBe("function");
  });

  it("returns the expected interface", () => {
    const result = useAuth();

    expect(result).toHaveProperty("user");
    expect(result).toHaveProperty("isLoading");
    expect(result).toHaveProperty("isAuthenticated");
    expect(result).toHaveProperty("login");
    expect(result).toHaveProperty("register");
    expect(result).toHaveProperty("logout");
    expect(result).toHaveProperty("updateUser");
    expect(result).toHaveProperty("refreshUser");
  });

  it("returns isAuthenticated false when user is null", async () => {
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({
      user: null,
      isLoading: false,
      logout: mockStoreLogout,
      setUser: mockSetUser,
    });

    const result = useAuth();
    expect(result.isAuthenticated).toBe(false);
  });

  it("returns isAuthenticated true when user exists", async () => {
    const { useAuthStore } = await import("../../stores/auth-store");

    vi.mocked(useAuthStore).mockReturnValue({
      user: { id: 1, name: "Test User", email: "test@example.com" },
      isLoading: false,
      logout: mockStoreLogout,
      setUser: mockSetUser,
    });

    const result = useAuth();
    expect(result.isAuthenticated).toBe(true);
  });

  describe("login", () => {
    it("calls loginMutation.mutate with credentials", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");

      vi.mocked(useAuthStore).mockReturnValue({
        user: null,
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();
      const credentials = { email: "test@example.com", password: "password123" };

      // Login returns a Promise, so we need to handle it
      const loginPromise = result.login(credentials);

      // The mutate function should have been called
      expect(mockLoginMutate).toHaveBeenCalledWith(credentials, expect.any(Object));

      // Simulate success callback
      const mutateCall = mockLoginMutate.mock.calls[0];
      const callbacks = mutateCall[1];
      callbacks.onSuccess();

      await expect(loginPromise).resolves.toBeUndefined();
    });

    it("rejects when login fails", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");

      vi.mocked(useAuthStore).mockReturnValue({
        user: null,
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();
      const credentials = { email: "test@example.com", password: "wrong" };

      const loginPromise = result.login(credentials);

      // Simulate error callback
      const mutateCall = mockLoginMutate.mock.calls[0];
      const callbacks = mutateCall[1];
      const mockError = new Error("Invalid credentials");
      callbacks.onError(mockError);

      await expect(loginPromise).rejects.toThrow("Invalid credentials");
    });
  });

  describe("register", () => {
    it("calls registerMutation.mutate with user data", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");

      vi.mocked(useAuthStore).mockReturnValue({
        user: null,
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();
      const userData = {
        email: "new@example.com",
        password: "password123",
        name: "New User",
      };

      const registerPromise = result.register(userData);

      expect(mockRegisterMutate).toHaveBeenCalledWith(userData, expect.any(Object));

      // Simulate success callback
      const mutateCall = mockRegisterMutate.mock.calls[0];
      const callbacks = mutateCall[1];
      callbacks.onSuccess();

      await expect(registerPromise).resolves.toBeUndefined();
    });
  });

  describe("logout", () => {
    it("clears feature flags and calls store logout", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");

      vi.mocked(useAuthStore).mockReturnValue({
        user: { id: 1, name: "Test", email: "test@example.com" },
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();
      result.logout();

      expect(mockClearFeatureFlags).toHaveBeenCalled();
      expect(mockStoreLogout).toHaveBeenCalled();
    });
  });

  describe("updateUser", () => {
    it("throws error when not authenticated", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");

      vi.mocked(useAuthStore).mockReturnValue({
        user: null,
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();

      await expect(result.updateUser({ name: "New Name" })).rejects.toThrow("Not authenticated");
    });

    it("calls updateUserMutation when authenticated", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");

      const mockUser = { id: 1, name: "Test User", email: "test@example.com" };
      vi.mocked(useAuthStore).mockReturnValue({
        user: mockUser,
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();
      const updateData = { name: "Updated Name" };

      const updatePromise = result.updateUser(updateData);

      expect(mockUpdateUserMutate).toHaveBeenCalledWith({ ...mockUser, ...updateData }, expect.any(Object));

      // Simulate success callback
      const mutateCall = mockUpdateUserMutate.mock.calls[0];
      const callbacks = mutateCall[1];
      callbacks.onSuccess();

      await expect(updatePromise).resolves.toBeUndefined();
    });
  });

  describe("refreshUser", () => {
    it("fetches current user and updates store", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");
      const { AuthService } = await import("../../services/auth/authService");

      const freshUser = { id: 1, name: "Fresh User", email: "fresh@example.com" };
      vi.mocked(AuthService.getCurrentUser).mockResolvedValue(freshUser);

      vi.mocked(useAuthStore).mockReturnValue({
        user: { id: 1, name: "Old User", email: "old@example.com" },
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();
      await result.refreshUser();

      expect(AuthService.getCurrentUser).toHaveBeenCalled();
      expect(mockSetUser).toHaveBeenCalledWith(freshUser);
    });

    it("logs out user when refresh fails", async () => {
      const { useAuthStore } = await import("../../stores/auth-store");
      const { AuthService } = await import("../../services/auth/authService");

      vi.mocked(AuthService.getCurrentUser).mockRejectedValue(new Error("Session expired"));

      vi.mocked(useAuthStore).mockReturnValue({
        user: { id: 1, name: "Test", email: "test@example.com" },
        isLoading: false,
        logout: mockStoreLogout,
        setUser: mockSetUser,
      });

      const result = useAuth();

      await expect(result.refreshUser()).rejects.toThrow("Session expired");
      expect(mockStoreLogout).toHaveBeenCalled();
    });
  });
});
