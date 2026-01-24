import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ErrorBoundary } from "./ErrorBoundary";
import { ErrorFallback } from "./ErrorFallback";

// Mock lucide-react icons
vi.mock("lucide-react", () => ({
  AlertCircle: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "alert-icon" }, "!"),
  Bug: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "bug-icon" }, "🐛"),
  ChevronDown: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "chevron-down-icon" }, "▼"),
  ChevronUp: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "chevron-up-icon" }, "▲"),
  Copy: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "copy-icon" }, "📋"),
  Home: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "home-icon" }, "H"),
  RefreshCw: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "refresh-icon" }, "R"),
  Wifi: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "wifi-icon" }, "📶"),
  WifiOff: ({ className }: { className?: string }) =>
    React.createElement("span", { className, "data-testid": "wifi-off-icon" }, "📵"),
}));

// Mock Button component with asChild support
vi.mock("../../ui/button", () => ({
  Button: ({
    children,
    asChild,
    onClick,
    ...props
  }: {
    children: React.ReactNode;
    asChild?: boolean;
    onClick?: () => void;
    className?: string;
    variant?: string;
    disabled?: boolean;
    size?: string;
  }) => {
    if (asChild && React.isValidElement(children)) {
      return React.cloneElement(children as React.ReactElement<Record<string, unknown>>, { onClick, ...props });
    }
    return React.createElement("button", { onClick, ...props }, children);
  },
}));

// Mock Card components
vi.mock("../../ui/card", () => ({
  Card: ({ children, className }: { children: React.ReactNode; className?: string }) =>
    React.createElement("div", { className }, children),
  CardContent: ({ children, className }: { children: React.ReactNode; className?: string }) =>
    React.createElement("div", { className }, children),
  CardDescription: ({ children, className }: { children: React.ReactNode; className?: string }) =>
    React.createElement("p", { className }, children),
  CardHeader: ({ children, className }: { children: React.ReactNode; className?: string }) =>
    React.createElement("div", { className }, children),
  CardTitle: ({ children, className }: { children: React.ReactNode; className?: string }) =>
    React.createElement("h2", { className }, children),
}));

// Mock react-i18next
vi.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (key: string, defaultValue: string) => defaultValue,
  }),
}));

// Mock sentry
vi.mock("../../lib/sentry", () => ({
  captureError: vi.fn(),
}));

// Mock query client
vi.mock("../../lib/query-client", () => ({
  queryClient: {
    invalidateQueries: vi.fn(),
  },
}));

// Component that throws an error
function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error("Test error message");
  }
  return <div>No error</div>;
}

// Suppress console.error in tests to reduce noise
const originalError = console.error;
beforeEach(() => {
  console.error = vi.fn();
});

afterEach(() => {
  console.error = originalError;
});

describe("ErrorBoundary", () => {
  it("renders children when there is no error", () => {
    render(
      <ErrorBoundary>
        <div>Child content</div>
      </ErrorBoundary>
    );

    expect(screen.getByText("Child content")).toBeInTheDocument();
  });

  it("catches errors and displays fallback UI", () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText("Something Went Wrong")).toBeInTheDocument();
    expect(screen.getByText("Test error message")).toBeInTheDocument();
  });

  it("displays custom fallback when provided", () => {
    render(
      <ErrorBoundary fallback={<div>Custom fallback</div>}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText("Custom fallback")).toBeInTheDocument();
  });

  it("calls onError callback when error is caught", () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(onError).toHaveBeenCalledTimes(1);
    expect(onError).toHaveBeenCalledWith(
      expect.any(Error),
      expect.objectContaining({ componentStack: expect.any(String) })
    );
  });

  it("resets error state when reset button is clicked", () => {
    const onReset = vi.fn();
    let shouldThrow = true;

    const { rerender } = render(
      <ErrorBoundary onReset={onReset}>
        <ThrowError shouldThrow={shouldThrow} />
      </ErrorBoundary>
    );

    // Error should be displayed
    expect(screen.getByText("Something Went Wrong")).toBeInTheDocument();

    // Click the reset button
    const resetButton = screen.getByRole("button", { name: /try again/i });

    // Before clicking, update the variable so re-render won't throw
    shouldThrow = false;
    fireEvent.click(resetButton);

    // onReset should have been called
    expect(onReset).toHaveBeenCalledTimes(1);
  });

  it("shows Try Again button by default", () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
  });

  it("shows Go Home link in fallback", () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByRole("link", { name: /go home/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /go home/i })).toHaveAttribute("href", "/");
  });
});

describe("ErrorFallback", () => {
  it("renders error message", () => {
    const error = new Error("Test error");

    render(<ErrorFallback error={error} />);

    expect(screen.getByText("Something Went Wrong")).toBeInTheDocument();
    expect(screen.getByText("Test error")).toBeInTheDocument();
  });

  it("renders generic message when error has no message", () => {
    const error = new Error();

    render(<ErrorFallback error={error} />);

    // When error has no message, the description fallback is shown
    expect(screen.getByText("An unexpected error occurred.")).toBeInTheDocument();
  });

  it("renders reset button when resetError is provided", () => {
    const error = new Error("Test error");
    const resetError = vi.fn();

    render(
      <ErrorFallback
        error={error}
        resetError={resetError}
      />
    );

    const resetButton = screen.getByRole("button", { name: /try again/i });
    expect(resetButton).toBeInTheDocument();

    fireEvent.click(resetButton);
    expect(resetError).toHaveBeenCalledTimes(1);
  });

  it("does not render reset button when resetError is not provided", () => {
    const error = new Error("Test error");

    render(<ErrorFallback error={error} />);

    expect(screen.queryByRole("button", { name: /try again/i })).not.toBeInTheDocument();
  });

  it("renders Go Home link", () => {
    const error = new Error("Test error");

    render(<ErrorFallback error={error} />);

    const homeLink = screen.getByRole("link", { name: /go home/i });
    expect(homeLink).toBeInTheDocument();
    expect(homeLink).toHaveAttribute("href", "/");
  });
});
