import React from "react";

import * as matchers from "@testing-library/jest-dom/matchers";
import { cleanup } from "@testing-library/react";
import { afterEach, expect, vi } from "vitest";

// Extend expect with jest-dom matchers
expect.extend(matchers);

// Cleanup after each test case
afterEach(() => {
  cleanup();
});

// Note: Router hooks are handled by renderWithProviders in test-utils.tsx
// Individual test files can mock router hooks as needed

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock auth store
vi.mock("../stores/auth-store", () => ({
  useAuthStore: vi.fn(),
}));

// Mock auth mutations
vi.mock("../hooks/mutations/use-auth-mutations", () => ({
  useLogin: vi.fn(),
  useRegister: vi.fn(),
}));

// Mock react-hook-form
vi.mock("react-hook-form", () => ({
  useForm: vi.fn(() => ({
    register: vi.fn(),
    handleSubmit: vi.fn((fn) => fn),
    formState: {
      errors: {},
    },
    watch: vi.fn(),
    setValue: vi.fn(),
    getValues: vi.fn(),
    reset: vi.fn(),
    control: {},
  })),
}));

// Mock @hookform/resolvers
vi.mock("@hookform/resolvers/zod", () => ({
  zodResolver: vi.fn(() => vi.fn()),
}));

// Mock shadcn UI components
type MockComponentProps = React.HTMLAttributes<HTMLElement> & { children?: React.ReactNode };

vi.mock("../components/ui/button", () => ({
  Button: ({ children, ...props }: MockComponentProps) => <button {...props}>{children}</button>,
}));

vi.mock("../components/ui/input", () => ({
  Input: (props: React.InputHTMLAttributes<HTMLInputElement>) => <input {...props} />,
}));

vi.mock("../components/ui/label", () => ({
  Label: ({ children, ...props }: MockComponentProps) => <label {...props}>{children}</label>,
}));

vi.mock("../components/ui/card", () => ({
  Card: ({ children, ...props }: MockComponentProps) => <div {...props}>{children}</div>,
  CardContent: ({ children, ...props }: MockComponentProps) => <div {...props}>{children}</div>,
  CardDescription: ({ children, ...props }: MockComponentProps) => <p {...props}>{children}</p>,
  CardHeader: ({ children, ...props }: MockComponentProps) => <div {...props}>{children}</div>,
  CardTitle: ({ children, ...props }: MockComponentProps) => <h2 {...props}>{children}</h2>,
}));

vi.mock("../components/ui/alert", () => ({
  Alert: ({ children, ...props }: MockComponentProps) => <div {...props}>{children}</div>,
  AlertDescription: ({ children, ...props }: MockComponentProps) => <div {...props}>{children}</div>,
}));

// Mock lucide-react icons
vi.mock("lucide-react", () => ({
  Loader2: () => <div>Loading...</div>,
  Eye: () => <div>👁️</div>,
  EyeOff: () => <div>👁️‍🗨️</div>,
}));
