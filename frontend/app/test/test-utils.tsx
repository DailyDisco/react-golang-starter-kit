import React, { type ReactElement } from "react";

import { QueryClient, QueryClientProvider, type UseMutationResult } from "@tanstack/react-query";
// Import router functions after the mock is set up. These imports will now
// get the mocked versions for useNavigate and useLocation, and the mocked
// implementations for others.
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
  useLocation,
  useNavigate,
} from "@tanstack/react-router";
import { fireEvent, render, type RenderOptions } from "@testing-library/react";
import { vi } from "vitest";

import type { User } from "../services";

// Mock @tanstack/react-router globally
vi.mock("@tanstack/react-router", () => {
  const navigateMock = vi.fn();
  const locationMock = {
    pathname: "/",
    search: {},
    hash: "",
    state: null,
    key: "default",
  };

  return {
    // Mock router hooks directly
    useNavigate: vi.fn(() => navigateMock),
    useLocation: vi.fn(() => locationMock),

    // Provide a simple RouterProvider that just renders its children
    RouterProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement("div", { "data-testid": "router-provider" }, children),

    // Mock other necessary router functions to prevent errors if they are called
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
    Outlet: ({ children }: { children?: React.ReactNode }) =>
      React.createElement("div", { "data-testid": "router-outlet" }, children),
  };
});

// Export the mocked hooks so test files can use them
export { useLocation, useNavigate };

// Test data factories
export const createMockUser = (overrides?: Partial<User>): User => ({
  id: 1,
  name: "Test User",
  email: "test@example.com",
  email_verified: true,
  is_active: true,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  ...overrides,
});

export const createMockAuthResponse = (overrides?: Partial<User>) => ({
  user: createMockUser(overrides),
  token: "mock-jwt-token",
});

// Create a test router for testing components that use router hooks
const createTestRouter = (initialPath: string = "/") => {
  const rootRoute = createRootRoute({
    component: () => React.createElement("div", { "data-testid": "router-outlet" }),
  });

  const routeTree = rootRoute.addChildren([]);

  return createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: [initialPath] }),
  });
};

// Custom render function that includes providers
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });

interface CustomRenderOptions extends Omit<RenderOptions, "wrapper"> {
  queryClient?: QueryClient;
  router?: ReturnType<typeof createTestRouter>;
  initialPath?: string;
}

export const renderWithProviders = (
  ui: ReactElement,
  { queryClient = createTestQueryClient(), router, initialPath = "/", ...renderOptions }: CustomRenderOptions = {}
) => {
  const testRouter = router ?? createTestRouter(initialPath);

  const TestWrapper = ({ children }: { children: React.ReactNode }) => (
    <RouterProvider router={testRouter}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </RouterProvider>
  );
  TestWrapper.displayName = "TestWrapper";

  return render(ui, { wrapper: TestWrapper, ...renderOptions });
};

// Simple render without providers (use renderWithProviders for router-dependent components)
export const renderSimple = (ui: ReactElement, options?: RenderOptions) => render(ui, options);

// Mock implementations for hooks
interface MockAuthStore {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  setUser: ReturnType<typeof vi.fn>;
  setToken: ReturnType<typeof vi.fn>;
  setLoading: ReturnType<typeof vi.fn>;
  logout: ReturnType<typeof vi.fn>;
  login: ReturnType<typeof vi.fn>;
  initialize: ReturnType<typeof vi.fn>;
}

export const createMockAuthStore = (overrides?: Partial<MockAuthStore>): MockAuthStore => ({
  user: null,
  token: null,
  isLoading: false,
  isAuthenticated: false,
  setUser: vi.fn(),
  setToken: vi.fn(),
  setLoading: vi.fn(),
  logout: vi.fn(),
  login: vi.fn(),
  initialize: vi.fn(),
  ...overrides,
});

/**
 * Creates a properly typed mock for UseMutationResult.
 * Use generic types to match your mutation's TData, TError, TVariables, TContext.
 */
export const createMockMutation = <TData = unknown, TError = Error, TVariables = void, TContext = unknown>(
  overrides?: Partial<UseMutationResult<TData, TError, TVariables, TContext>>
): UseMutationResult<TData, TError, TVariables, TContext> =>
  ({
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
    isError: false,
    isSuccess: false,
    isIdle: true,
    isPaused: false,
    status: "idle",
    error: null,
    data: undefined as TData | undefined,
    variables: undefined as TVariables | undefined,
    context: undefined as TContext | undefined,
    failureCount: 0,
    failureReason: null,
    reset: vi.fn(),
    submittedAt: 0,
    ...overrides,
  }) as UseMutationResult<TData, TError, TVariables, TContext>;

export const createMockNavigate = () => vi.fn();

interface MockLocation {
  pathname: string;
  search: string;
  hash: string;
  state: unknown;
  key: string;
}

export const createMockLocation = (overrides?: Partial<MockLocation>): MockLocation => ({
  pathname: "/",
  search: "",
  hash: "",
  state: null,
  key: "default",
  ...overrides,
});

// Form test helpers
export const fillFormField = (input: HTMLElement, value: string) => {
  fireEvent.change(input, { target: { value } });
};

export const submitForm = (form: HTMLElement) => {
  fireEvent.submit(form);
};

export const clickButton = (button: HTMLElement) => {
  fireEvent.click(button);
};

// Re-export everything from testing-library for convenience
export * from "@testing-library/react";
