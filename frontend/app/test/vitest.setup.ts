import React from 'react';
import { vi } from 'vitest';
import { Outlet } from '@tanstack/react-router';

// A more explicit and robust mock for @tanstack/react-router
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
    Outlet: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { 'data-testid': 'router-outlet' }, children),
  };
});
