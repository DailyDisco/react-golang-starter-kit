import { useCallback, useEffect } from "react";

type KeyboardHandler = (event: KeyboardEvent) => void;

interface ShortcutConfig {
  /** The key to listen for (e.g., "k", "s", "Escape") */
  key: string;
  /** Require Cmd (Mac) / Ctrl (Windows/Linux) */
  meta?: boolean;
  /** Require Shift */
  shift?: boolean;
  /** Require Alt/Option */
  alt?: boolean;
  /** Handler function */
  handler: KeyboardHandler;
  /** Whether to prevent default browser behavior */
  preventDefault?: boolean;
  /** Description for help display */
  description?: string;
}

interface UseKeyboardShortcutsOptions {
  /** Whether shortcuts are enabled (default: true) */
  enabled?: boolean;
  /** Shortcuts to register */
  shortcuts: ShortcutConfig[];
}

/**
 * Hook for registering global keyboard shortcuts
 *
 * @example
 * useKeyboardShortcuts({
 *   shortcuts: [
 *     { key: "k", meta: true, handler: () => openCommandPalette(), description: "Open command palette" },
 *     { key: "s", meta: true, handler: () => save(), preventDefault: true, description: "Save" },
 *     { key: "Escape", handler: () => closeModal(), description: "Close modal" },
 *   ],
 * });
 */
export function useKeyboardShortcuts({ enabled = true, shortcuts }: UseKeyboardShortcutsOptions) {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Guard against undefined event.key (can happen with some special keys)
      if (!event.key) return;

      // Don't trigger shortcuts when typing in inputs
      const target = event.target as HTMLElement;
      const isInput =
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable ||
        target.closest("[role='textbox']");

      for (const shortcut of shortcuts) {
        const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase();
        const metaMatches = shortcut.meta ? event.metaKey || event.ctrlKey : !event.metaKey && !event.ctrlKey;
        const shiftMatches = shortcut.shift ? event.shiftKey : !event.shiftKey;
        const altMatches = shortcut.alt ? event.altKey : !event.altKey;

        // Allow Escape to work even in inputs
        const isEscape = shortcut.key.toLowerCase() === "escape";

        if (keyMatches && metaMatches && shiftMatches && altMatches) {
          // Skip if in input and not Escape
          if (isInput && !isEscape) continue;

          if (shortcut.preventDefault !== false) {
            event.preventDefault();
          }
          shortcut.handler(event);
          return;
        }
      }
    },
    [shortcuts]
  );

  useEffect(() => {
    if (!enabled) return;

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [enabled, handleKeyDown]);
}

/**
 * Get platform-aware modifier key label
 */
export function getModifierKey(): string {
  if (typeof navigator === "undefined") return "Ctrl";
  return navigator.platform?.toLowerCase().includes("mac") ? "⌘" : "Ctrl";
}

/**
 * Format a shortcut for display (e.g., "⌘K" or "Ctrl+K")
 */
export function formatShortcut(shortcut: Pick<ShortcutConfig, "key" | "meta" | "shift" | "alt">): string {
  const parts: string[] = [];
  const isMac = typeof navigator !== "undefined" && navigator.platform?.toLowerCase().includes("mac");

  if (shortcut.meta) parts.push(isMac ? "⌘" : "Ctrl");
  if (shortcut.shift) parts.push(isMac ? "⇧" : "Shift");
  if (shortcut.alt) parts.push(isMac ? "⌥" : "Alt");

  parts.push(shortcut.key.toUpperCase());

  return isMac ? parts.join("") : parts.join("+");
}
