import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

// Import after mocks
import { useAuth } from "../../hooks/auth/useAuth";
import { createMockUser } from "../../test/test-utils";
import { ProtectedRoute } from "./ProtectedRoute";

// Mock auth hook
vi.mock("../../hooks/auth/useAuth");

// Type the mock
const mockUseAuth = vi.mocked(useAuth);

// Create a minimal mock for the auth hook (only fields actually used by ProtectedRoute)
const createAuthMock = (overrides: Partial<ReturnType<typeof useAuth>> = {}) => ({
  isAuthenticated: false,
  isLoading: false,
  user: null,
  logout: vi.fn(),
  login: vi.fn(),
  register: vi.fn(),
  updateUser: vi.fn(),
  refreshUser: vi.fn(),
  ...overrides,
});

describe("ProtectedRoute", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("when loading", () => {
    it("should render loading spinner while authentication is loading", () => {
      mockUseAuth.mockReturnValue(createAuthMock({ isLoading: true }));

      render(
        <ProtectedRoute>
          <div data-testid="protected-content">Protected Content</div>
        </ProtectedRoute>
      );

      // Should show loading spinner
      const spinner = document.querySelector(".animate-spin");
      expect(spinner).toBeTruthy();

      // Should not show protected content
      expect(screen.queryByTestId("protected-content")).toBeNull();
    });
  });

  describe("when not authenticated", () => {
    it("should redirect to login when user is not authenticated", () => {
      mockUseAuth.mockReturnValue(createAuthMock());

      render(
        <ProtectedRoute>
          <div data-testid="protected-content">Protected Content</div>
        </ProtectedRoute>
      );

      // Should render the Navigate mock
      const navigateMock = screen.getByTestId("navigate-mock");
      expect(navigateMock).toBeTruthy();

      // Navigate should have been called with login redirect
      expect(navigateMock.getAttribute("data-to")).toBe("/login");
    });

    it("should redirect to custom redirectTo path when specified", () => {
      mockUseAuth.mockReturnValue(createAuthMock());

      render(
        <ProtectedRoute redirectTo="/custom-login">
          <div data-testid="protected-content">Protected Content</div>
        </ProtectedRoute>
      );

      const navigateMock = screen.getByTestId("navigate-mock");
      expect(navigateMock.getAttribute("data-to")).toBe("/custom-login");
    });
  });

  describe("when authenticated", () => {
    it("should render children when user is authenticated", () => {
      mockUseAuth.mockReturnValue(
        createAuthMock({
          isAuthenticated: true,
          user: createMockUser({ id: 1, name: "Test", email: "test@example.com" }),
        })
      );

      render(
        <ProtectedRoute>
          <div data-testid="protected-content">Protected Content</div>
        </ProtectedRoute>
      );

      // Should show protected content
      expect(screen.getByTestId("protected-content")).toBeTruthy();
      expect(screen.getByText("Protected Content")).toBeTruthy();

      // Navigate should not have been rendered
      expect(screen.queryByTestId("navigate-mock")).toBeNull();
    });
  });
});
