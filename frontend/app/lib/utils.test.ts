import { describe, it, expect } from "vitest";
import { cn } from "../lib/utils";

describe("cn utility function", () => {
  it("merges class names correctly", () => {
    expect(cn("px-2 py-1", "bg-red-500")).toBe("px-2 py-1 bg-red-500");
  });

  it("handles conditional classes", () => {
    const isActive = true;
    expect(cn("base-class", isActive && "active-class")).toBe(
      "base-class active-class",
    );
  });

  it("removes duplicate classes", () => {
    expect(cn("px-2", "px-4")).toBe("px-4");
  });

  it("handles undefined and null values", () => {
    expect(cn("class1", undefined, null, "class2")).toBe("class1 class2");
  });
});
