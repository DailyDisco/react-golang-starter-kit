import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { logger } from "./logger";

describe("logger", () => {
  // Store original console methods
  const originalConsoleDebug = console.debug;
  const originalConsoleInfo = console.info;
  const originalConsoleWarn = console.warn;
  const originalConsoleError = console.error;

  beforeEach(() => {
    // Mock console methods
    console.debug = vi.fn();
    console.info = vi.fn();
    console.warn = vi.fn();
    console.error = vi.fn();
  });

  afterEach(() => {
    // Restore original console methods
    console.debug = originalConsoleDebug;
    console.info = originalConsoleInfo;
    console.warn = originalConsoleWarn;
    console.error = originalConsoleError;
    vi.clearAllMocks();
  });

  describe("logger.error", () => {
    it("should always log errors", () => {
      logger.error("Test error message");
      expect(console.error).toHaveBeenCalled();
    });

    it("should include error details in log", () => {
      const error = new Error("Test error");
      logger.error("Error occurred", error);
      expect(console.error).toHaveBeenCalled();
      const loggedMessage = (console.error as ReturnType<typeof vi.fn>).mock.calls[0][0];
      expect(loggedMessage).toContain("Error occurred");
      expect(loggedMessage).toContain("Test error");
    });

    it("should include context in log", () => {
      logger.error("Error with context", null, { userId: 123, action: "test" });
      expect(console.error).toHaveBeenCalled();
      const loggedMessage = (console.error as ReturnType<typeof vi.fn>).mock.calls[0][0];
      expect(loggedMessage).toContain("userId");
      expect(loggedMessage).toContain("123");
    });
  });

  describe("logger.warn", () => {
    it("should log warnings", () => {
      logger.warn("Test warning");
      expect(console.warn).toHaveBeenCalled();
    });

    it("should include context in warning", () => {
      logger.warn("Warning with context", { detail: "some detail" });
      expect(console.warn).toHaveBeenCalled();
      const loggedMessage = (console.warn as ReturnType<typeof vi.fn>).mock.calls[0][0];
      expect(loggedMessage).toContain("Warning with context");
      expect(loggedMessage).toContain("some detail");
    });
  });

  describe("logger helper methods", () => {
    it("isDev should return a boolean", () => {
      expect(typeof logger.isDev()).toBe("boolean");
    });

    it("isProd should return a boolean", () => {
      expect(typeof logger.isProd()).toBe("boolean");
    });
  });

  describe("log message format", () => {
    it("should include timestamp in ERROR level logs", () => {
      logger.error("Test message");
      const loggedMessage = (console.error as ReturnType<typeof vi.fn>).mock.calls[0][0];
      // Check for ISO timestamp format
      expect(loggedMessage).toMatch(/\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
    });

    it("should include log level in uppercase", () => {
      logger.error("Test message");
      const loggedMessage = (console.error as ReturnType<typeof vi.fn>).mock.calls[0][0];
      expect(loggedMessage).toContain("[ERROR]");
    });
  });
});
