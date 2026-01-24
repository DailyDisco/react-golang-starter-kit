import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { CopyableCell, IdCell, TruncatedCopyableCell } from "./CopyableCell";

// Mock tooltip components as simple pass-through elements
vi.mock("../tooltip", () => ({
  Tooltip: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  TooltipTrigger: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  TooltipContent: ({ children }: { children: React.ReactNode }) => <span>{children}</span>,
  TooltipProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

// Mock clipboard API
const mockWriteText = vi.fn();
const originalClipboard = navigator.clipboard;

beforeAll(() => {
  Object.defineProperty(navigator, "clipboard", {
    value: { writeText: mockWriteText },
    writable: true,
    configurable: true,
  });
});

afterAll(() => {
  Object.defineProperty(navigator, "clipboard", {
    value: originalClipboard,
    writable: true,
    configurable: true,
  });
});

describe("CopyableCell", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWriteText.mockResolvedValue(undefined);
  });

  it("renders the value", () => {
    render(<CopyableCell value="test-id-123" />);

    expect(screen.getByText("test-id-123")).toBeInTheDocument();
  });

  it("renders numeric values", () => {
    render(<CopyableCell value={12345} />);

    expect(screen.getByText("12345")).toBeInTheDocument();
  });

  it("renders custom displayValue", () => {
    render(
      <CopyableCell
        value="full-api-key-value"
        displayValue="full-api...value"
      />
    );

    expect(screen.getByText("full-api...value")).toBeInTheDocument();
    // Full value should not be visible
    expect(screen.queryByText("full-api-key-value")).not.toBeInTheDocument();
  });

  it("copies the actual value, not displayValue", async () => {
    render(
      <CopyableCell
        value="full-secret-value"
        displayValue="full...value"
      />
    );

    const copyButton = screen.getByRole("button");
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith("full-secret-value");
    });
  });

  it("has copy button hidden by default", () => {
    render(<CopyableCell value="test" />);

    const copyButton = screen.getByRole("button");
    // Check that the button has the opacity-0 class
    expect(copyButton).toHaveClass("opacity-0");
  });

  it("applies monospace font by default", () => {
    render(<CopyableCell value="test-id" />);

    const valueSpan = screen.getByText("test-id");
    expect(valueSpan).toHaveClass("font-mono");
  });

  it("can disable monospace font", () => {
    render(<CopyableCell value="test-id" mono={false} />);

    const valueSpan = screen.getByText("test-id");
    expect(valueSpan).not.toHaveClass("font-mono");
  });

  it("applies maxWidth and shows title", () => {
    render(<CopyableCell value="very-long-value" maxWidth="100px" />);

    const valueSpan = screen.getByText("very-long-value");
    expect(valueSpan).toHaveClass("truncate");
    expect(valueSpan).toHaveStyle({ maxWidth: "100px" });
    expect(valueSpan).toHaveAttribute("title", "very-long-value");
  });

  it("applies custom className", () => {
    render(<CopyableCell value="test" className="custom-class" />);

    const container = screen.getByText("test").parentElement;
    expect(container).toHaveClass("custom-class");
  });

  it("uses custom copyLabel", () => {
    render(<CopyableCell value="test" copyLabel="Copy ID" />);

    const copyButton = screen.getByRole("button");
    expect(copyButton).toHaveAttribute("aria-label", "Copy ID");
  });
});

describe("IdCell", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWriteText.mockResolvedValue(undefined);
  });

  it("renders with default 'Copy ID' label", () => {
    render(<IdCell id="user-123" />);

    expect(screen.getByText("user-123")).toBeInTheDocument();
    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "Copy ID");
  });

  it("works with numeric IDs", () => {
    render(<IdCell id={42} />);

    expect(screen.getByText("42")).toBeInTheDocument();
  });
});

describe("TruncatedCopyableCell", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWriteText.mockResolvedValue(undefined);
  });

  it("truncates long values with ellipsis", () => {
    const longValue = "sk-1234567890abcdef1234567890abcdef";

    render(<TruncatedCopyableCell value={longValue} visibleChars={8} />);

    // Should show first 8 + ... + last 8
    // First 8: "sk-12345", Last 8: "90abcdef"
    expect(screen.getByText("sk-12345...90abcdef")).toBeInTheDocument();
  });

  it("does not truncate short values", () => {
    const shortValue = "short";

    render(<TruncatedCopyableCell value={shortValue} visibleChars={8} />);

    expect(screen.getByText("short")).toBeInTheDocument();
  });

  it("copies full value", async () => {
    const longValue = "sk-1234567890abcdef1234567890abcdef";

    render(<TruncatedCopyableCell value={longValue} />);

    const copyButton = screen.getByRole("button");
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith(longValue);
    });
  });

  it("uses custom copyLabel", () => {
    render(
      <TruncatedCopyableCell
        value="sk-1234567890abcdef"
        copyLabel="Copy API key"
      />
    );

    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "Copy API key");
  });
});
