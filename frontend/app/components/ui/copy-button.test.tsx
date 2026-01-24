import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { CopyButton } from "./copy-button";

// Mock tooltip components as simple pass-through elements
vi.mock("./tooltip", () => ({
  Tooltip: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  TooltipTrigger: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  TooltipContent: ({ children }: { children: React.ReactNode }) => <span>{children}</span>,
  TooltipProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

// Mock clipboard API
const mockWriteText = vi.fn();

describe("CopyButton", () => {
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

  beforeEach(() => {
    vi.clearAllMocks();
    mockWriteText.mockResolvedValue(undefined);
  });

  it("renders copy icon initially", () => {
    render(<CopyButton value="test-value" />);

    const button = screen.getByRole("button");
    expect(button).toBeInTheDocument();
    expect(button).toHaveAttribute("aria-label", "Copy to clipboard");
  });

  it("copies text to clipboard on click", async () => {
    render(<CopyButton value="test-value" />);

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith("test-value");
    });
  });

  it("shows check icon after copy", async () => {
    render(<CopyButton value="test-value" />);

    const button = screen.getByRole("button");
    fireEvent.click(button);

    await waitFor(() => {
      expect(button).toHaveAttribute("aria-label", "Copied!");
      expect(button).toHaveAttribute("data-copied", "true");
    });
  });

  it("reverts to copy icon after duration", async () => {
    vi.useFakeTimers();

    render(<CopyButton value="test-value" successDuration={1000} />);

    const button = screen.getByRole("button");

    // Click and wait for state update
    await act(async () => {
      fireEvent.click(button);
      // Flush the clipboard.writeText promise
      await Promise.resolve();
    });

    expect(button).toHaveAttribute("data-copied", "true");

    // Advance time past the success duration
    await act(async () => {
      vi.advanceTimersByTime(1100);
    });

    expect(button).toHaveAttribute("data-copied", "false");

    vi.useRealTimers();
  });

  it("calls onCopy callback after successful copy", async () => {
    const onCopy = vi.fn();

    render(<CopyButton value="test-value" onCopy={onCopy} />);

    const button = screen.getByRole("button");
    fireEvent.click(button);

    // Wait for the clipboard promise to resolve
    await vi.waitFor(() => {
      expect(mockWriteText).toHaveBeenCalled();
    });
    expect(onCopy).toHaveBeenCalledTimes(1);
  });

  it("does not call onCopy if clipboard fails", async () => {
    mockWriteText.mockRejectedValue(new Error("Clipboard error"));
    const onCopy = vi.fn();
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    render(<CopyButton value="test-value" onCopy={onCopy} />);

    const button = screen.getByRole("button");
    fireEvent.click(button);

    // Wait for the clipboard promise to reject and error to be logged
    await vi.waitFor(() => {
      expect(consoleSpy).toHaveBeenCalled();
    });
    expect(onCopy).not.toHaveBeenCalled();

    consoleSpy.mockRestore();
  });

  it("is keyboard accessible", () => {
    render(<CopyButton value="test-value" />);

    const button = screen.getByRole("button");
    button.focus();
    expect(button).toHaveFocus();
    // Button should be accessible via keyboard (role="button" allows keyboard activation)
  });

  it("supports custom labels", () => {
    render(
      <CopyButton
        value="test-value"
        label="Copy ID"
        successLabel="ID copied!"
      />
    );

    const button = screen.getByRole("button");
    expect(button).toHaveAttribute("aria-label", "Copy ID");
  });

  it("does not copy empty value", () => {
    render(<CopyButton value="" />);

    const button = screen.getByRole("button");
    fireEvent.click(button);

    // The handler returns early if value is empty, so writeText should never be called
    expect(mockWriteText).not.toHaveBeenCalled();
  });

  it("applies custom className", () => {
    render(<CopyButton value="test-value" className="custom-class" />);

    const button = screen.getByRole("button");
    expect(button).toHaveClass("custom-class");
  });

  it("supports different button variants", () => {
    render(<CopyButton value="test-value" variant="outline" size="sm" />);

    const button = screen.getByRole("button");
    expect(button).toBeInTheDocument();
  });
});
