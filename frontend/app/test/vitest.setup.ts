import React from "react";

import { vi } from "vitest";

// A more explicit and robust mock for @tanstack/react-router
vi.mock("@tanstack/react-router", () => {
  const navigateMock = vi.fn();
  const locationMock = {
    pathname: "/",
    search: {}, // Ensure search is an object
    hash: "",
    state: null,
    key: "default",
  };

  return {
    // Mock router hooks directly to always return vi.fn() instances
    useNavigate: vi.fn(),
    useLocation: vi.fn(() => locationMock),

    // Provide a simple RouterProvider that just renders its children
    // We'll give it a data-testid for easier debugging if needed
    RouterProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement("div", { "data-testid": "router-provider" }, children),

    // Mock other necessary router functions and components to prevent errors if they are called
    // or accessed internally by the router context itself.
    createMemoryHistory: vi.fn(() => ({
      initialEntries: ["/"],
      push: vi.fn(),
      replace: vi.fn(),
    })),
    createRootRoute: vi.fn((config: Record<string, unknown>) => ({
      ...config,
      addChildren: vi.fn(() => config),
    })),
    createRoute: vi.fn((config: Record<string, unknown>) => config),
    createRouter: vi.fn((config: Record<string, unknown>) => ({
      ...config,
      navigate: navigateMock,
      location: locationMock,
    })),
    Link: ({ to, children, ...props }: { to: string; children: React.ReactNode; [key: string]: unknown }) =>
      React.createElement("a", { href: to, ...props }, children),
    Outlet: ({ children }: { children: React.ReactNode }) =>
      React.createElement("div", { "data-testid": "router-outlet" }, children),

    // Ensure other commonly used exports are also mocked to prevent undefined errors
    // Add more as needed if specific errors arise related to other exports.
    useParams: vi.fn(() => ({})),
    useMatch: vi.fn(() => ({})),
    useRouterState: vi.fn(() => ({ location: locationMock })),
    useRouter: vi.fn(() => ({
      navigate: navigateMock,
      location: locationMock,
    })),

    // Export some internal mock functions if tests need direct access (e.g., to reset state)
    _mockNavigate: navigateMock,
    _mockLocation: locationMock,
  };
});
