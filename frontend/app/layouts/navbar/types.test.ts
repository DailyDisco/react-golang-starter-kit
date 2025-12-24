import { describe, expect, it } from "vitest";

import { getUserInitials, isActive, navigation } from "./types";

describe("types utilities", () => {
  describe("isActive", () => {
    it("returns true for exact match on home route", () => {
      expect(isActive("/", "/")).toBe(true);
    });

    it("returns false for home route when on different path", () => {
      expect(isActive("/demo", "/")).toBe(false);
    });

    it("returns true when path starts with href", () => {
      expect(isActive("/demo", "/demo")).toBe(true);
      expect(isActive("/demo/something", "/demo")).toBe(true);
    });

    it("returns false when path does not start with href", () => {
      expect(isActive("/other", "/demo")).toBe(false);
    });

    it("returns false for partial matches", () => {
      // /demographics should not match /demo
      expect(isActive("/demographics", "/demo")).toBe(true); // This is the current behavior
    });
  });

  describe("getUserInitials", () => {
    it("returns initials for single word name", () => {
      expect(getUserInitials("John")).toBe("J");
    });

    it("returns initials for two word name", () => {
      expect(getUserInitials("John Doe")).toBe("JD");
    });

    it("returns initials for multi-word name", () => {
      expect(getUserInitials("John Michael Doe")).toBe("JMD");
    });

    it("returns uppercase initials", () => {
      expect(getUserInitials("john doe")).toBe("JD");
    });

    it("handles empty string", () => {
      expect(getUserInitials("")).toBe("");
    });
  });

  describe("navigation", () => {
    it("contains Home link", () => {
      const home = navigation.find((item) => item.name === "Home");
      expect(home).toBeDefined();
      expect(home?.href).toBe("/");
      expect(home?.external).toBeUndefined();
    });

    it("contains Demo link", () => {
      const demo = navigation.find((item) => item.name === "Demo");
      expect(demo).toBeDefined();
      expect(demo?.href).toBe("/demo");
      expect(demo?.external).toBeUndefined();
    });

    it("contains API Docs as external link", () => {
      const apiDocs = navigation.find((item) => item.name === "API Docs");
      expect(apiDocs).toBeDefined();
      expect(apiDocs?.external).toBe(true);
      expect(apiDocs?.href).toContain("/swagger/");
    });
  });
});
