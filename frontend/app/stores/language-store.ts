import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

import i18n, { supportedLanguages, type SupportedLanguage } from "../i18n";
import { logger } from "../lib/logger";

interface LanguageState {
  language: SupportedLanguage;
  isInitialized: boolean;

  // Actions
  setLanguage: (language: SupportedLanguage) => void;
  syncFromBackend: (language: string) => void;
  initialize: () => void;
}

export const useLanguageStore = create<LanguageState>()(
  devtools(
    persist(
      (set, get) => ({
        language: "en",
        isInitialized: false,

        setLanguage: (language) => {
          if (!supportedLanguages.includes(language)) {
            logger.warn(`Unsupported language: ${language}`);
            return;
          }

          // Update i18next
          void i18n.changeLanguage(language);

          // Update store
          set({ language });

          // Update HTML lang attribute for accessibility
          if (typeof document !== "undefined") {
            document.documentElement.lang = language;
          }
        },

        syncFromBackend: (language) => {
          // Called when user preferences are loaded from backend
          if (supportedLanguages.includes(language as SupportedLanguage)) {
            get().setLanguage(language as SupportedLanguage);
          }
        },

        initialize: () => {
          if (get().isInitialized) return;

          // Get current i18n language (detected by LanguageDetector or from localStorage)
          const detectedLang = i18n.language?.split("-")[0] as SupportedLanguage;

          if (supportedLanguages.includes(detectedLang)) {
            set({ language: detectedLang });
          }

          // Set HTML lang attribute
          if (typeof document !== "undefined") {
            document.documentElement.lang = get().language;
          }

          set({ isInitialized: true });
        },
      }),
      {
        name: "language-storage",
        partialize: (state) => ({ language: state.language }),
      }
    ),
    { name: "language-store" }
  )
);
