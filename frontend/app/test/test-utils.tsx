import React, { type ReactElement } from 'react';
import { render, type RenderOptions, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi } from 'vitest';
import type { User } from '../services';

// Mock @tanstack/react-router at module level
vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router');
  return {
    ...actual,
    createMemoryHistory: vi.fn(() => ({
      initialEntries: ['/'],
      push: vi.fn(),
      replace: vi.fn(),
    })),
    RouterProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement(
        'div',
        { 'data-testid': 'router-provider' },
        children
      ),
    createRootRoute: vi.fn((config: any) => ({
      ...config,
      addChildren: vi.fn(() => ({
        ...config,
        addChildren: vi.fn(() => config),
      })),
    })),
    createRoute: vi.fn((config: any) => config),
    createRouter: vi.fn((config: any) => ({
      ...config,
      navigate: vi.fn(),
      location: { pathname: '/', search: '', hash: '', state: null },
    })),
    useNavigate: vi.fn(() => vi.fn()),
    useLocation: vi.fn(() => ({
      pathname: '/',
      search: '',
      hash: '',
      state: null,
    })),
    Link: ({ to, children, ...props }: any) =>
      React.createElement('a', { href: to, ...props }, children),
  };
});

// Import router functions after mock is set up
import {
  createMemoryHistory,
  RouterProvider,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router';

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
const createTestRouter = () => {
  const rootRoute = createRootRoute({
    component: () => <div>Test Root</div>,
  });

  const loginRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/login',
    component: () => <div>Login Page</div>,
  });

  const registerRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/register',
    component: () => <div>Register Page</div>,
  });

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => <div>Home Page</div>,
  });

  const dashboardRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/dashboard',
    component: () => <div>Dashboard Page</div>,
  });

  const routeTree = rootRoute.addChildren([
    loginRoute,
    registerRoute,
    indexRoute,
    dashboardRoute,
  ]);

  return createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: ['/'] }),
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
}

export const renderWithProviders = (
  ui: ReactElement,
  {
    queryClient = createTestQueryClient(),
    router = createTestRouter(),
    ...renderOptions
  }: CustomRenderOptions = {}
) => {
  try {
    let Wrapper: React.ComponentType<{ children: React.ReactNode }>;

    // Always include router unless explicitly set to null
    if (router !== null) {
      Wrapper = ({ children }: { children: React.ReactNode }) => (
        <RouterProvider router={router}>
          <QueryClientProvider client={queryClient}>
            {children}
          </QueryClientProvider>
        </RouterProvider>
      );
    } else {
      // Skip router if explicitly set to null
      Wrapper = ({ children }: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      );
    }

    return render(ui, { wrapper: Wrapper, ...renderOptions });
  } catch (error) {
    console.error('Error in renderWithProviders:', error);
    throw error;
  }
};

// Simple render without providers for debugging
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
