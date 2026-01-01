import type { LucideIcon } from "lucide-react";

// =============================================================================
// Core Types
// =============================================================================

/**
 * User object shape for command context
 */
export interface CommandUser {
  id: string | number;
  email: string;
  name?: string;
  role: "user" | "admin" | "super_admin";
}

/**
 * Command execution context - passed to all command handlers
 */
export interface CommandContext {
  user: CommandUser | null;
  pathname: string;
  routeParams: Record<string, string>;
  navigate: (to: string) => void;
  close: () => void;
  setMode: (mode: PaletteMode) => void;
  setSearch: (query: string) => void;
}

/**
 * Command group categories for organization
 */
export type CommandGroup = "navigation" | "actions" | "settings" | "admin" | "contextual" | "search";

/**
 * Palette operating modes
 */
export type PaletteMode =
  | "default" // Normal command palette
  | "search" // Full-text search mode
  | "impersonate" // User impersonation search
  | "flag-toggle"; // Feature flag toggle mode

/**
 * Keyboard shortcut configuration
 */
export interface ShortcutConfig {
  key: string;
  meta?: boolean;
  shift?: boolean;
  alt?: boolean;
}

// =============================================================================
// Command Definitions
// =============================================================================

/**
 * Preview component props
 */
export interface CommandPreviewProps<T = unknown> {
  command: Command;
  ctx: CommandContext;
  data?: T;
}

/**
 * Base command definition
 */
export interface Command {
  /** Unique identifier */
  id: string;
  /** Display label */
  label: string;
  /** Optional description */
  description?: string;
  /** Lucide icon component */
  icon?: LucideIcon;
  /** Keyboard shortcut */
  shortcut?: ShortcutConfig;
  /** Search keywords */
  keywords?: string[];
  /** Command group for organization */
  group: CommandGroup;

  // Access Control
  /** Required roles to see this command */
  roles?: Array<"user" | "admin" | "super_admin">;
  /** Route patterns where this command appears (glob-style) */
  routePatterns?: string[];
  /** Custom visibility condition */
  condition?: (ctx: CommandContext) => boolean;

  // Execution
  /** Action handler - can be async */
  action: (ctx: CommandContext) => void | Promise<void>;

  // Confirmation
  /** Whether action requires confirmation */
  requiresConfirmation?: boolean;
  /** Custom confirmation message */
  confirmationMessage?: string;
  /** Whether the action is destructive (red styling) */
  isDestructive?: boolean;

  // Preview
  /** Preview panel configuration */
  preview?: {
    component: React.ComponentType<CommandPreviewProps>;
    preloadData?: (ctx: CommandContext) => Promise<unknown>;
  };
}

// =============================================================================
// Search Types
// =============================================================================

/**
 * Search result types
 */
export type SearchResultType = "user" | "audit_log" | "feature_flag" | "page" | "action" | "help";

/**
 * Search result from async providers
 */
export interface SearchResult {
  /** Unique identifier */
  id: string;
  /** Result type for categorization */
  type: SearchResultType;
  /** Display title */
  title: string;
  /** Optional subtitle */
  subtitle?: string;
  /** Lucide icon component */
  icon?: LucideIcon;
  /** Action when selected */
  action: (ctx: CommandContext) => void | Promise<void>;
  /** Preview configuration */
  preview?: Command["preview"];
  /** Additional metadata */
  metadata?: Record<string, unknown>;
}

/**
 * Search provider interface for async data sources
 */
export interface SearchProvider {
  /** Unique identifier */
  id: string;
  /** Display name */
  name: string;
  /** Result types this provider returns */
  types: SearchResultType[];
  /** Required roles to use this provider */
  roles?: Array<"user" | "admin" | "super_admin">;
  /** Search function */
  search: (query: string, ctx: CommandContext) => Promise<SearchResult[]>;
  /** Debounce delay in ms */
  debounceMs?: number;
  /** Minimum query length to trigger search */
  minQueryLength?: number;
}

/**
 * Command provider for batch command registration
 */
export interface CommandProvider {
  /** Unique identifier */
  id: string;
  /** Display name */
  name: string;
  /** Get commands for current context */
  getCommands: (ctx: CommandContext) => Command[];
}

// =============================================================================
// State Types
// =============================================================================

/**
 * Action execution status
 */
export type ActionStatus = "idle" | "confirming" | "loading" | "success" | "error";

/**
 * Command palette store state
 */
export interface CommandPaletteState {
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
  setConfirmingAction: (actionId: string | null) => void;
  setActionStatus: (status: ActionStatus) => void;
  setActionError: (error: string | null) => void;
  reset: () => void;
}

// =============================================================================
// Utility Types
// =============================================================================

/**
 * Fuzzy match result with scoring
 */
export interface FuzzyMatch<T> {
  item: T;
  score: number;
  matches: Array<{ start: number; end: number }>;
}

/**
 * Recent search entry for history
 */
export interface RecentSearch {
  query: string;
  timestamp: number;
  resultType?: SearchResultType;
  resultId?: string;
}

/**
 * Command usage stats for frequently used
 */
export interface CommandUsageStats {
  commandId: string;
  useCount: number;
  lastUsed: number;
}
