import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";

// Import all translation files
import enAdmin from "./locales/en/admin.json";
import enAuth from "./locales/en/auth.json";
import enBilling from "./locales/en/billing.json";
import enCommon from "./locales/en/common.json";
import enDashboard from "./locales/en/dashboard.json";
import enErrors from "./locales/en/errors.json";
import enLanding from "./locales/en/landing.json";
import enPricing from "./locales/en/pricing.json";
import enSettings from "./locales/en/settings.json";
import enValidation from "./locales/en/validation.json";
import esAdmin from "./locales/es/admin.json";
import esAuth from "./locales/es/auth.json";
import esBilling from "./locales/es/billing.json";
import esCommon from "./locales/es/common.json";
import esDashboard from "./locales/es/dashboard.json";
import esErrors from "./locales/es/errors.json";
import esLanding from "./locales/es/landing.json";
import esPricing from "./locales/es/pricing.json";
import esSettings from "./locales/es/settings.json";
import esValidation from "./locales/es/validation.json";

export const supportedLanguages = ["en", "es"] as const;
export type SupportedLanguage = (typeof supportedLanguages)[number];

export const languageNames: Record<SupportedLanguage, string> = {
  en: "English",
  es: "Espanol",
};

export const defaultNS = "common";

export const resources = {
  en: {
    common: enCommon,
    auth: enAuth,
    errors: enErrors,
    validation: enValidation,
    settings: enSettings,
    admin: enAdmin,
    landing: enLanding,
    pricing: enPricing,
    dashboard: enDashboard,
    billing: enBilling,
  },
  es: {
    common: esCommon,
    auth: esAuth,
    errors: esErrors,
    validation: esValidation,
    settings: esSettings,
    admin: esAdmin,
    landing: esLanding,
    pricing: esPricing,
    dashboard: esDashboard,
    billing: esBilling,
  },
} as const;

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: "en",
    defaultNS,
    ns: ["common", "auth", "errors", "validation", "settings", "admin", "landing", "pricing", "dashboard", "billing"],

    detection: {
      // Order of detection - localStorage first, then browser preference
      order: ["localStorage", "navigator"],
      lookupLocalStorage: "language",
      caches: ["localStorage"],
    },

    interpolation: {
      escapeValue: false, // React already escapes
    },

    react: {
      useSuspense: true,
    },
  });

export default i18n;
