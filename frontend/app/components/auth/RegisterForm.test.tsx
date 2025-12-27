import { beforeEach, describe, expect, it, vi } from "vitest";

import { renderWithProviders, screen } from "../../test/test-utils";
import { RegisterForm } from "./RegisterForm";

// Mock the hooks
const mockRegisterMutate = vi.fn();
const mockNavigate = vi.fn();

vi.mock("../../hooks/mutations/use-auth-mutations", () => ({
  useRegister: vi.fn(() => ({
    mutate: mockRegisterMutate,
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

describe("RegisterForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders all form fields", () => {
    renderWithProviders(<RegisterForm />);

    expect(screen.getByLabelText(/full name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
  });

  it("renders create account button", () => {
    renderWithProviders(<RegisterForm />);

    expect(screen.getByRole("button", { name: /create account/i })).toBeInTheDocument();
  });

  it("renders sign in link", () => {
    renderWithProviders(<RegisterForm />);

    expect(screen.getByRole("link", { name: /sign in/i })).toBeInTheDocument();
  });

  it("has form with all required input fields", () => {
    renderWithProviders(<RegisterForm />);

    const nameInput = screen.getByLabelText(/full name/i);
    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/^password$/i);
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

    expect(nameInput).toHaveAttribute("type", "text");
    expect(emailInput).toHaveAttribute("type", "email");
    expect(passwordInput).toHaveAttribute("type", "password");
    expect(confirmPasswordInput).toHaveAttribute("type", "password");
  });

  it("has password fields with password type by default", () => {
    renderWithProviders(<RegisterForm />);

    const passwordInput = screen.getByLabelText(/^password$/i);
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

    expect(passwordInput).toHaveAttribute("type", "password");
    expect(confirmPasswordInput).toHaveAttribute("type", "password");
  });
});

describe("RegisterForm - Pending State", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("disables form fields when mutation is pending", async () => {
    const { useRegister } = await import("../../hooks/mutations/use-auth-mutations");
    vi.mocked(useRegister).mockReturnValue({
      mutate: mockRegisterMutate,
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
    });

    renderWithProviders(<RegisterForm />);

    expect(screen.getByLabelText(/full name/i)).toBeDisabled();
    expect(screen.getByLabelText(/email/i)).toBeDisabled();
    expect(screen.getByLabelText(/^password$/i)).toBeDisabled();
    expect(screen.getByLabelText(/confirm password/i)).toBeDisabled();
    expect(screen.getByRole("button", { name: /create account/i })).toBeDisabled();
  });
});
