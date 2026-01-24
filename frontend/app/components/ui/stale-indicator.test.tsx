import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { formatTimeAgo, StaleIndicator } from "./stale-indicator";

describe("StaleIndicator", () => {
  const mockRefresh = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders nothing when not stale and not fetching", () => {
    const { container } = render(
      <StaleIndicator
        isStale={false}
        isFetching={false}
        onRefresh={mockRefresh}
      />
    );

    expect(container.firstChild).toBeNull();
  });

  it("renders when stale", () => {
    render(
      <StaleIndicator
        isStale={true}
        isFetching={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByRole("button")).toBeInTheDocument();
    expect(screen.getByText("Stale")).toBeInTheDocument();
  });

  it("renders when fetching", () => {
    render(
      <StaleIndicator
        isStale={false}
        isFetching={true}
        onRefresh={mockRefresh}
      />
    );

    const button = screen.getByRole("button");
    expect(button).toBeInTheDocument();
    expect(button).toBeDisabled();
  });

  it("renders when showRefreshAlways is true", () => {
    render(
      <StaleIndicator
        isStale={false}
        isFetching={false}
        onRefresh={mockRefresh}
        showRefreshAlways={true}
      />
    );

    expect(screen.getByRole("button")).toBeInTheDocument();
    // Should not show stale badge when not actually stale
    expect(screen.queryByText("Stale")).not.toBeInTheDocument();
  });

  it("calls onRefresh when clicked", async () => {
    const user = userEvent.setup();

    render(
      <StaleIndicator
        isStale={true}
        isFetching={false}
        onRefresh={mockRefresh}
      />
    );

    await user.click(screen.getByRole("button"));

    expect(mockRefresh).toHaveBeenCalledTimes(1);
  });

  it("does not call onRefresh when fetching", async () => {
    const user = userEvent.setup();

    render(
      <StaleIndicator
        isStale={true}
        isFetching={true}
        onRefresh={mockRefresh}
      />
    );

    const button = screen.getByRole("button");
    expect(button).toBeDisabled();

    // Clicking disabled button shouldn't trigger callback
    await user.click(button);
    expect(mockRefresh).not.toHaveBeenCalled();
  });

  it("hides stale badge when fetching", () => {
    render(
      <StaleIndicator
        isStale={true}
        isFetching={true}
        onRefresh={mockRefresh}
      />
    );

    // Badge should not show during fetch
    expect(screen.queryByText("Stale")).not.toBeInTheDocument();
  });

  it("has correct aria-label when refreshing", () => {
    render(
      <StaleIndicator
        isStale={false}
        isFetching={true}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "Refreshing");
  });

  it("has correct aria-label when not refreshing", () => {
    render(
      <StaleIndicator
        isStale={true}
        isFetching={false}
        onRefresh={mockRefresh}
      />
    );

    expect(screen.getByRole("button")).toHaveAttribute("aria-label", "Refresh data");
  });

  it("applies custom className", () => {
    render(
      <StaleIndicator
        isStale={true}
        isFetching={false}
        onRefresh={mockRefresh}
        className="custom-class"
      />
    );

    expect(screen.getByRole("button")).toHaveClass("custom-class");
  });
});

describe("formatTimeAgo", () => {
  it("returns 'just now' for very recent times", () => {
    const now = new Date();
    expect(formatTimeAgo(now)).toBe("just now");
  });

  it("returns seconds for times under a minute", () => {
    const thirtySecsAgo = new Date(Date.now() - 30 * 1000);
    expect(formatTimeAgo(thirtySecsAgo)).toBe("30s ago");
  });

  it("returns minutes for times under an hour", () => {
    const fiveMinsAgo = new Date(Date.now() - 5 * 60 * 1000);
    expect(formatTimeAgo(fiveMinsAgo)).toBe("5m ago");
  });

  it("returns hours for times under a day", () => {
    const threeHoursAgo = new Date(Date.now() - 3 * 60 * 60 * 1000);
    expect(formatTimeAgo(threeHoursAgo)).toBe("3h ago");
  });

  it("returns days for times over a day", () => {
    const twoDaysAgo = new Date(Date.now() - 2 * 24 * 60 * 60 * 1000);
    expect(formatTimeAgo(twoDaysAgo)).toBe("2d ago");
  });
});
