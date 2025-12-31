import { act } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useLanguageStore } from "./language-store";

// Mock i18n
vi.mock("../i18n", () => ({
  default: {
    changeLanguage: vi.fn().mockResolvedValue(undefined),
    language: "en",
  },
  supportedLanguages: ["en", "es"],
}));

// Mock logger
vi.mock("../lib/logger", () => ({
  logger: {
    warn: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

describe("useLanguageStore", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset store to initial state before each test
    act(() => {
      useLanguageStore.setState({
        language: "en",
        isInitialized: false,
      });
    });
  });

  describe("initial state", () => {
    it("has 'en' as default language", () => {
      const state = useLanguageStore.getState();
      expect(state.language).toBe("en");
    });

    it("has isInitialized set to false initially", () => {
      const state = useLanguageStore.getState();
      expect(state.isInitialized).toBe(false);
    });
  });

  describe("setLanguage", () => {
    it("sets the language to a supported language", async () => {
      const i18n = await import("../i18n");

      act(() => {
        useLanguageStore.getState().setLanguage("es");
      });

      expect(useLanguageStore.getState().language).toBe("es");
      expect(i18n.default.changeLanguage).toHaveBeenCalledWith("es");
    });

    it("updates the HTML lang attribute", () => {
      act(() => {
        useLanguageStore.getState().setLanguage("es");
      });

      expect(document.documentElement.lang).toBe("es");
    });

    it("logs warning for unsupported language", async () => {
      const { logger } = await import("../lib/logger");

      act(() => {
        // @ts-expect-error - Testing invalid language
        useLanguageStore.getState().setLanguage("fr");
      });

      expect(logger.warn).toHaveBeenCalledWith("Unsupported language: fr");
      // Language should remain unchanged
      expect(useLanguageStore.getState().language).toBe("en");
    });

    it("does not change language for unsupported language", async () => {
      const i18n = await import("../i18n");

      act(() => {
        useLanguageStore.getState().setLanguage("es");
      });

      expect(useLanguageStore.getState().language).toBe("es");

      act(() => {
        // @ts-expect-error - Testing invalid language
        useLanguageStore.getState().setLanguage("invalid");
      });

      // Should remain 'es'
      expect(useLanguageStore.getState().language).toBe("es");
    });
  });

  describe("syncFromBackend", () => {
    it("syncs language from backend if supported", () => {
      act(() => {
        useLanguageStore.getState().syncFromBackend("es");
      });

      expect(useLanguageStore.getState().language).toBe("es");
    });

    it("ignores unsupported language from backend", () => {
      act(() => {
        useLanguageStore.getState().syncFromBackend("fr");
      });

      // Should remain 'en'
      expect(useLanguageStore.getState().language).toBe("en");
    });
  });

  describe("initialize", () => {
    it("sets isInitialized to true", () => {
      expect(useLanguageStore.getState().isInitialized).toBe(false);

      act(() => {
        useLanguageStore.getState().initialize();
      });

      expect(useLanguageStore.getState().isInitialized).toBe(true);
    });

    it("does not re-initialize if already initialized", () => {
      act(() => {
        useLanguageStore.getState().initialize();
      });

      const firstInitState = useLanguageStore.getState();
      expect(firstInitState.isInitialized).toBe(true);

      // Set language to something different
      act(() => {
        useLanguageStore.getState().setLanguage("es");
      });

      // Try to initialize again
      act(() => {
        useLanguageStore.getState().initialize();
      });

      // Language should remain 'es', not reset
      expect(useLanguageStore.getState().language).toBe("es");
    });

    it("sets HTML lang attribute during initialization", async () => {
      const i18n = await import("../i18n");

      // Mock i18n.language to return "es"
      Object.defineProperty(i18n.default, "language", {
        value: "es",
        writable: true,
        configurable: true,
      });

      act(() => {
        useLanguageStore.setState({ language: "en", isInitialized: false });
      });

      act(() => {
        useLanguageStore.getState().initialize();
      });

      // The initialize function detects language from i18n, sets store, then sets HTML lang
      expect(document.documentElement.lang).toBe("es");
      expect(useLanguageStore.getState().language).toBe("es");
    });
  });
});
