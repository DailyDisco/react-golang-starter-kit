import type { UseMutationResult } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { AuthResponse, LoginRequest } from "../../services";
import { renderWithProviders, screen } from "../../test/test-utils";
import { LoginForm } from "./LoginForm";

// Mock the hooks
const mockLoginMutate = vi.fn();
const mockNavigate = vi.fn();

vi.mock("../../hooks/mutations/use-auth-mutations", () => ({
  useLogin: vi.fn(() => ({
    mutate: mockLoginMutate,
    isPending: false,
    isError: false,
    error: null,
  })),
}));

vi.mock("../../stores/auth-store", () => ({
  useAuthStore: vi.fn(() => ({
    isLoading: false,
  })),
}));

vi.mock("@tanstack/react-router", async () => {
  const actual = await vi.importActual("@tanstack/react-router");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useLocation: () => ({
      pathname: "/login",
      state: {},
    }),
    Link: ({ children, to, ...props }: { children: React.ReactNode; to: string }) => (
      <a
        href={to}
        {...props}
      >
        {children}
      </a>
    ),
  };
});

describe("LoginForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders email and password fields", () => {
    renderWithProviders(<LoginForm />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it("renders sign in button", () => {
    renderWithProviders(<LoginForm />);

    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("renders sign up link", () => {
    renderWithProviders(<LoginForm />);

    expect(screen.getByRole("link", { name: /sign up/i })).toBeInTheDocument();
  });

  it("has form with email and password inputs", () => {
    renderWithProviders(<LoginForm />);

    const form = screen.getByRole("form");
    expect(form).toBeInTheDocument();

    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/password/i);

    expect(emailInput).toHaveAttribute("type", "email");
    expect(passwordInput).toHaveAttribute("type", "password");
  });

  it("has password field with password type by default", () => {
    renderWithProviders(<LoginForm />);

    const passwordInput = screen.getByLabelText(/password/i);
    expect(passwordInput).toHaveAttribute("type", "password");
  });
});

describe("LoginForm - Pending State", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("disables form fields when mutation is pending", async () => {
    // Re-mock with isPending: true
    const { useLogin } = await import("../../hooks/mutations/use-auth-mutations");
    vi.mocked(useLogin).mockReturnValue({
      mutate: mockLoginMutate,
      mutateAsync: vi.fn(),
      isPending: true,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      reset: vi.fn(),
      variables: undefined,
      status: "pending",
      failureCount: 0,
      failureReason: null,
      isIdle: false,
      isPaused: false,
      context: undefined,
      submittedAt: 0,
    } as unknown as UseMutationResult<AuthResponse, Error, LoginRequest, unknown>);

    renderWithProviders(<LoginForm />);

    expect(screen.getByLabelText(/email/i)).toBeDisabled();
    expect(screen.getByLabelText(/password/i)).toBeDisabled();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeDisabled();
  });
});
