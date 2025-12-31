import { create } from "zustand";

interface CommandPaletteState {
  isOpen: boolean;
  open: () => void;
  close: () => void;
  toggle: () => void;
}

/**
 * Global state for command palette visibility
 */
export const useCommandPaletteStore = create<CommandPaletteState>((set) => ({
  isOpen: false,
  open: () => set({ isOpen: true }),
  close: () => set({ isOpen: false }),
  toggle: () => set((state) => ({ isOpen: !state.isOpen })),
}));

/**
 * Hook for accessing command palette state and actions
 */
export function useCommandPalette() {
  const { isOpen, open, close, toggle } = useCommandPaletteStore();
  return { isOpen, open, close, toggle };
}
