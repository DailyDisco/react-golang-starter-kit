import React, { type ReactElement } from 'react';
import { render, type RenderOptions, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi } from 'vitest';
import type { User } from '../services';

// Mock @tanstack/react-router globally
vi.mock('@tanstack/react-router', () => {
  const navigateMock = vi.fn();
  const locationMock = {
    pathname: '/',
    search: {}, // Ensure search is an object
    hash: '',
    state: null,
    key: 'default',
  };

  return {
    // Mock router hooks directly
    useNavigate: vi.fn(() => navigateMock),
    useLocation: vi.fn(() => locationMock),

    // Provide a simple RouterProvider that just renders its children
    RouterProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement(
        'div',
        { 'data-testid': 'router-provider' },
        children
      ),

    // Mock other necessary router functions to prevent errors if they are called
    createMemoryHistory: vi.fn(() => ({
      initialEntries: ['/'],
      push: vi.fn(),
      replace: vi.fn(),
    })),
    createRootRoute: vi.fn((config: any) => ({
      ...config,
      addChildren: vi.fn(() => config),
    })),
    createRoute: vi.fn((config: any) => config),
    createRouter: vi.fn((config: any) => ({
      ...config,
      navigate: navigateMock,
      location: locationMock,
    })),
    Link: ({ to, children, ...props }: any) =>
      React.createElement('a', { href: to, ...props }, children),
    Outlet: ({ children }: { children?: React.ReactNode }) =>
      React.createElement('div', { 'data-testid': 'router-outlet' }, children),
  };
});

// Import router functions after the mock is set up. These imports will now
// get the mocked versions for useNavigate and useLocation, and the mocked
// implementations for others.
import {
  createMemoryHistory,
  RouterProvider,
  createRootRoute,
  createRoute,
  createRouter,
  useNavigate,
  useLocation,
  Link,
  Outlet,
} from '@tanstack/react-router';

// Export the mocked hooks so test files can use them
export { useNavigate, useLocation };

// Test data factories
export const createMockUser = (overrides?: Partial<User>): User => ({
  id: 1,
  name: 'Test User',
  email: 'test@example.com',
  email_verified: true,
  is_active: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  ...overrides,
});

export const createMockAuthResponse = (overrides?: Partial<User>) => ({
  user: createMockUser(overrides),
  token: 'mock-jwt-token',
});

// Create a test router for testing components that use router hooks
const createTestRouter = (initialPath: string = '/') => {
  const rootRoute = createRootRoute({
    component: () => React.createElement('div', null, 'Test Root'),
  });

  const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/login',
    component: () => React.createElement('div', null, 'Login Page'),
  });

  const registerRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/register',
    component: () => React.createElement('div', null, 'Register Page'),
  });

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => React.createElement('div', null, 'Home Page'),
  });

  const dashboardRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/dashboard',
    component: () => React.createElement('div', null, 'Dashboard Page'),
  });

  const routeTree = rootRoute.addChildren([
    loginRoute,
    registerRoute,
    indexRoute,
    dashboardRoute,
  ]);

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

interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  queryClient?: QueryClient;
  router?: ReturnType<typeof createTestRouter>;
  initialPath?: string;
}

export const renderWithProviders = (
  ui: ReactElement,
  {
    queryClient = createTestQueryClient(),
    router,
    initialPath = '/',
    ...renderOptions
  }: CustomRenderOptions = {}
) => {
  try {
    const testRouter = router || createTestRouter(initialPath);
    let Wrapper: React.ComponentType<{ children: React.ReactNode }>;

    Wrapper = ({ children }: { children: React.ReactNode }) => (
      <RouterProvider router={testRouter}>
        <QueryClientProvider client={queryClient}>
          {ui}{' '}
          {/* Render the UI component directly as children of QueryClientProvider */}
          {children} {/* Render any additional children passed to Wrapper */}
        </QueryClientProvider>
      </RouterProvider>
    );

    return render(ui, { wrapper: Wrapper, ...renderOptions });
  } catch (error) {
    console.error('Error in renderWithProviders:', error);
    throw error;
  }
};

// Simple render without providers for debugging (use renderWithProviders for router-dependent components)
export const renderSimple = (ui: ReactElement, options?: RenderOptions) => {
  try {
    return render(ui, options);
  } catch (error) {
    console.error('Error in renderSimple:', error);
    throw error;
  }
};

// Mock implementations for hooks
export const createMockAuthStore = (overrides?: any) => ({
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

export const createMockMutation = (overrides?: any) => ({
  mutate: vi.fn(),
  mutateAsync: vi.fn(),
  isPending: false,
  isError: false,
  isSuccess: false,
  error: null,
  data: null,
  reset: vi.fn(),
  ...overrides,
});

export const createMockNavigate = () => vi.fn();

// Add this to help debug rendering issues
export const renderWithDebug = (ui: ReactElement) => {
  const result = renderWithProviders(ui);
  console.log('Rendered HTML:', result.container.innerHTML);
  return result;
};

export const createMockLocation = (overrides?: any) => ({
  pathname: '/',
  search: '',
  hash: '',
  state: null,
  key: 'default',
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
export * from '@testing-library/react';
