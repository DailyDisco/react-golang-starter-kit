import * as React from "react";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

type Theme = "light" | "dark" | "system";

interface ThemeProviderState {
  theme: Theme;
  resolvedTheme: "light" | "dark";
  setTheme: (theme: Theme) => void;
}

const ThemeProviderContext = createContext<ThemeProviderState | undefined>(undefined);

interface ThemeProviderProps {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
}

function getSystemTheme(): "light" | "dark" {
  if (typeof window === "undefined") return "light";
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function ThemeProvider({ children, defaultTheme = "system", storageKey = "theme" }: ThemeProviderProps) {
  const [theme, setThemeState] = useState<Theme>(() => {
    if (typeof window === "undefined") return defaultTheme;
    const stored = localStorage.getItem(storageKey) as Theme | null;
    return stored || defaultTheme;
  });

  const [resolvedTheme, setResolvedTheme] = useState<"light" | "dark">(() => {
    if (theme === "system") return getSystemTheme();
    return theme;
  });

  // Apply theme to document
  const applyTheme = useCallback((newTheme: "light" | "dark") => {
    const root = document.documentElement;
    root.classList.remove("light", "dark");
    root.classList.add(newTheme);
    setResolvedTheme(newTheme);
  }, []);

  // Handle theme changes
  const setTheme = useCallback(
    (newTheme: Theme) => {
      setThemeState(newTheme);
      localStorage.setItem(storageKey, newTheme);

      if (newTheme === "system") {
        applyTheme(getSystemTheme());
      } else {
        applyTheme(newTheme);
      }
    },
    [storageKey, applyTheme]
  );

  // Initialize theme on mount
  useEffect(() => {
    const appliedTheme = theme === "system" ? getSystemTheme() : theme;
    applyTheme(appliedTheme);
  }, [theme, applyTheme]);

  // Listen for system theme changes
  useEffect(() => {
    if (theme !== "system") return;

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

    const handleChange = (e: MediaQueryListEvent) => {
      applyTheme(e.matches ? "dark" : "light");
    };

    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, [theme, applyTheme]);

  // Prevent flash of wrong theme - apply immediately on hydration
  useEffect(() => {
    const stored = localStorage.getItem(storageKey) as Theme | null;
    const initialTheme = stored || defaultTheme;
    const appliedTheme = initialTheme === "system" ? getSystemTheme() : initialTheme;

    document.documentElement.classList.remove("light", "dark");
    document.documentElement.classList.add(appliedTheme);
  }, [storageKey, defaultTheme]);

  const value = useMemo(
    () => ({
      theme,
      resolvedTheme,
      setTheme,
    }),
    [theme, resolvedTheme, setTheme]
  );

  return <ThemeProviderContext.Provider value={value}>{children}</ThemeProviderContext.Provider>;
}

export function useTheme() {
  const context = useContext(ThemeProviderContext);

  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }

  return context;
}
