import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { UserMenu } from "./UserMenu";

// Mock lucide-react icons
vi.mock("lucide-react", () => ({
  CreditCard: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "credit-card-icon" }, "💳"),
  DollarSign: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "dollar-icon" }, "$"),
  LogOut: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "logout-icon" }, "X"),
  Settings: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "settings-icon" }, "⚙"),
  Shield: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "shield-icon" }, "🛡"),
  User: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "user-icon" }, "U"),
}));

// Mock dependencies
vi.mock("@tanstack/react-router", () => ({
  Link: ({ children, to, ...props }: { children: React.ReactNode; to: string }) => (
    <a
      href={to}
      {...props}
    >
      {children}
    </a>
  ),
}));

const mockUser = {
  id: 1,
  name: "John Doe",
  email: "john@example.com",
  email_verified: true,
  is_active: true,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

describe("UserMenu", () => {
  const mockOnLogout = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders user avatar with initials", () => {
    render(
      <UserMenu
        user={mockUser}
        onLogout={mockOnLogout}
      />
    );

    // The avatar fallback should show initials
    expect(screen.getByText("JD")).toBeInTheDocument();
  });

  it("renders trigger button", () => {
    render(
      <UserMenu
        user={mockUser}
        onLogout={mockOnLogout}
      />
    );

    // The button should be present
    expect(screen.getByRole("button")).toBeInTheDocument();
  });

  it("handles single-word names correctly", () => {
    const singleNameUser = { ...mockUser, name: "John" };

    render(
      <UserMenu
        user={singleNameUser}
        onLogout={mockOnLogout}
      />
    );

    expect(screen.getByText("J")).toBeInTheDocument();
  });

  it("handles multi-word names correctly", () => {
    const multiNameUser = { ...mockUser, name: "John Michael Doe" };

    render(
      <UserMenu
        user={multiNameUser}
        onLogout={mockOnLogout}
      />
    );

    expect(screen.getByText("JMD")).toBeInTheDocument();
  });
});
