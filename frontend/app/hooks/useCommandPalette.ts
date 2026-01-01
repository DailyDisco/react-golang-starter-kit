import type { ActionStatus, Command, PaletteMode, SearchResult } from "@/services/command-palette/types";
import { create } from "zustand";
import { devtools } from "zustand/middleware";

// =============================================================================
// Types
// =============================================================================

interface CommandPaletteState {
  // UI State
  isOpen: boolean;
  mode: PaletteMode;
  search: string;
  selectedIndex: number;

  // Search state
  isSearching: boolean;
  searchResults: SearchResult[];
  searchError: Error | null;

  // Preview state
  previewCommand: Command | SearchResult | null;
  previewData: unknown;
  isPreviewLoading: boolean;

  // Confirmation state
  confirmingActionId: string | null;
  actionStatus: ActionStatus;
  actionError: string | null;

  // Actions
  open: (mode?: PaletteMode) => void;
  close: () => void;
  toggle: () => void;
  setMode: (mode: PaletteMode) => void;
  setSearch: (query: string) => void;
  setSelectedIndex: (index: number) => void;
  setSearchResults: (results: SearchResult[]) => void;
  setSearching: (searching: boolean) => void;
  setSearchError: (error: Error | null) => void;
  setPreview: (command: Command | SearchResult | null) => void;
  setPreviewData: (data: unknown) => void;
  setPreviewLoading: (loading: boolean) => void;
  setConfirmingAction: (actionId: string | null) => void;
  setActionStatus: (status: ActionStatus) => void;
  setActionError: (error: string | null) => void;
  reset: () => void;
}

// =============================================================================
// Initial State
// =============================================================================

const initialState = {
  isOpen: false,
  mode: "default" as PaletteMode,
  search: "",
  selectedIndex: 0,
  isSearching: false,
  searchResults: [],
  searchError: null,
  previewCommand: null,
  previewData: null,
  isPreviewLoading: false,
  confirmingActionId: null,
  actionStatus: "idle" as ActionStatus,
  actionError: null,
};

// =============================================================================
// Store
// =============================================================================

/**
 * Enhanced command palette store with support for:
 * - Multiple modes (default, search, impersonate, flag-toggle)
 * - Async search with loading/error states
 * - Preview panel for selected commands
 * - Action confirmation flow
 */
export const useCommandPaletteStore = create<CommandPaletteState>()(
  devtools(
    (set) => ({
      ...initialState,

      // Open palette, optionally in a specific mode
      open: (mode: PaletteMode = "default") =>
        set({
          isOpen: true,
          mode,
          search: "",
          selectedIndex: 0,
          searchResults: [],
          searchError: null,
          confirmingActionId: null,
          actionStatus: "idle",
          actionError: null,
        }),

      // Close palette and reset state
      close: () =>
        set({
          isOpen: false,
          confirmingActionId: null,
          actionStatus: "idle",
          actionError: null,
        }),

      // Toggle palette open/closed
      toggle: () =>
        set((state) => ({
          isOpen: !state.isOpen,
          // Reset to default mode when toggling open
          mode: state.isOpen ? state.mode : "default",
          search: state.isOpen ? state.search : "",
          selectedIndex: 0,
        })),

      // Switch palette mode
      setMode: (mode: PaletteMode) =>
        set({
          mode,
          search: "",
          selectedIndex: 0,
          searchResults: [],
          searchError: null,
        }),

      // Update search query
      setSearch: (search: string) =>
        set({
          search,
          selectedIndex: 0,
        }),

      // Update selected index for keyboard navigation
      setSelectedIndex: (selectedIndex: number) => set({ selectedIndex }),

      // Update search results
      setSearchResults: (searchResults: SearchResult[]) => set({ searchResults }),

      // Update searching state
      setSearching: (isSearching: boolean) => set({ isSearching }),

      // Update search error
      setSearchError: (searchError: Error | null) => set({ searchError }),

      // Set preview command
      setPreview: (previewCommand: Command | SearchResult | null) =>
        set({
          previewCommand,
          previewData: null,
          isPreviewLoading: previewCommand !== null,
        }),

      // Set preview data
      setPreviewData: (previewData: unknown) =>
        set({
          previewData,
          isPreviewLoading: false,
        }),

      // Set preview loading state
      setPreviewLoading: (isPreviewLoading: boolean) => set({ isPreviewLoading }),

      // Set confirming action ID
      setConfirmingAction: (confirmingActionId: string | null) => set({ confirmingActionId }),

      // Set action execution status
      setActionStatus: (actionStatus: ActionStatus) => set({ actionStatus }),

      // Set action error
      setActionError: (actionError: string | null) => set({ actionError }),

      // Reset to initial state
      reset: () => set(initialState),
    }),
    { name: "command-palette-store" }
  )
);

// =============================================================================
// Hooks
// =============================================================================

/**
 * Hook for accessing command palette state and actions
 * Provides backward compatibility with existing usage
 */
export function useCommandPalette() {
  const store = useCommandPaletteStore();

  return {
    // Basic state (backward compatible)
    isOpen: store.isOpen,
    open: store.open,
    close: store.close,
    toggle: store.toggle,

    // Extended state
    mode: store.mode,
    search: store.search,
    selectedIndex: store.selectedIndex,
    isSearching: store.isSearching,
    searchResults: store.searchResults,
    searchError: store.searchError,
    previewCommand: store.previewCommand,
    previewData: store.previewData,
    isPreviewLoading: store.isPreviewLoading,
    confirmingActionId: store.confirmingActionId,
    actionStatus: store.actionStatus,
    actionError: store.actionError,

    // Extended actions
    setMode: store.setMode,
    setSearch: store.setSearch,
    setSelectedIndex: store.setSelectedIndex,
    setSearchResults: store.setSearchResults,
    setSearching: store.setSearching,
    setSearchError: store.setSearchError,
    setPreview: store.setPreview,
    setPreviewData: store.setPreviewData,
    setPreviewLoading: store.setPreviewLoading,
    setConfirmingAction: store.setConfirmingAction,
    setActionStatus: store.setActionStatus,
    setActionError: store.setActionError,
    reset: store.reset,
  };
}
