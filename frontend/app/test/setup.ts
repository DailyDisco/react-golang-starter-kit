import * as matchers from '@testing-library/jest-dom/matchers';
import { cleanup } from '@testing-library/react';
import { afterEach, expect, vi } from 'vitest';

// Extend expect with jest-dom matchers
expect.extend(matchers);

// Cleanup after each test case
afterEach(() => {
  cleanup();
});

// Mock router hooks
vi.mock('@tanstack/react-router', async () => ({
  ...((await vi.importActual('@tanstack/react-router')) as any),
  useNavigate: vi.fn(),
  useLocation: vi.fn(),
  Link: ({ to, children, ...props }: any) => ({
    type: 'a',
    props: { href: to, ...props, children },
  }),
}));

// Mock sonner toast
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock auth store
vi.mock('../stores/auth-store', () => ({
  useAuthStore: vi.fn(),
}));

// Mock auth mutations
vi.mock('../hooks/mutations/use-auth-mutations', () => ({
  useLogin: vi.fn(),
  useRegister: vi.fn(),
}));
