import React from "react";

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

// Import after mocks
import { useAuth } from "../../hooks/auth/useAuth";
import { createMockUser } from "../../test/test-utils";
import { ProtectedRoute } from "./ProtectedRoute";

// Track Navigate calls
interface NavigateProps {
  to: string;
  replace?: boolean;
  state?: unknown;
}
let navigateProps: NavigateProps | null = null;

// Mock both modules before any imports
vi.mock("../../hooks/auth/useAuth");
vi.mock("@tanstack/react-router", () => ({
  Navigate: (props: NavigateProps) => {
    navigateProps = props;
    return React.createElement("div", { "data-testid": "navigate-mock" });
  },
  useLocation: () => ({ pathname: "/dashboard" }),
}));

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
    navigateProps = null;
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
      expect(screen.getByTestId("navigate-mock")).toBeTruthy();

      // Navigate should have been called with login redirect
      expect(navigateProps).not.toBeNull();
      expect(navigateProps?.to).toBe("/login");
      expect(navigateProps?.replace).toBe(true);
    });

    it("should redirect to custom redirectTo path when specified", () => {
      mockUseAuth.mockReturnValue(createAuthMock());

      render(
        <ProtectedRoute redirectTo="/custom-login">
          <div data-testid="protected-content">Protected Content</div>
        </ProtectedRoute>
      );

      expect(navigateProps).not.toBeNull();
      expect(navigateProps?.to).toBe("/custom-login");
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

      // Navigate should not have been called
      expect(navigateProps).toBeNull();
    });
  });
});
