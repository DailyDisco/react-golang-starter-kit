import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import { useAuth } from "@/hooks/auth/useAuth";
import { useCommandContext } from "@/hooks/useCommandContext";
import { useCommandPalette } from "@/hooks/useCommandPalette";
import { useCommandSearch, useSearchProviders } from "@/hooks/useCommandSearch";
import { formatShortcut, useKeyboardShortcuts } from "@/hooks/useKeyboardShortcuts";
import { useTheme } from "@/providers/theme-provider";
import {
  adminProvider,
  commandRegistry,
  contextualProvider,
  createActionProvider,
  navigationProvider,
  settingsProvider,
  type Command,
} from "@/services/command-palette";
import type { SearchResult } from "@/services/command-palette/types";
import { AlertTriangle, ArrowLeft, Loader2, LogOut, Moon, Search, Sun, ToggleLeft, User, UserCog } from "lucide-react";
import { toast } from "sonner";

// =============================================================================
// Types
// =============================================================================

interface GroupedCommands {
  navigation: Command[];
  actions: Command[];
  settings: Command[];
  admin: Command[];
  contextual: Command[];
}

// =============================================================================
// Provider Registration
// =============================================================================

/**
 * Hook to register all command providers on mount
 * Uses refs to prevent re-registration loops when theme/auth changes
 */
function useCommandProviders() {
  const { setTheme, resolvedTheme } = useTheme();
  const { logout } = useAuth();

  // Use refs to access current values without re-registering providers
  // This prevents the infinite loop caused by cleanup → notify → setState → re-render
  const themeRef = useRef(resolvedTheme);
  const setThemeRef = useRef(setTheme);
  const logoutRef = useRef(logout);

  // Update refs when values change (doesn't trigger effect re-run)
  useEffect(() => {
    themeRef.current = resolvedTheme;
    setThemeRef.current = setTheme;
    logoutRef.current = logout;
  });

  // Register providers ONCE on mount only
  useEffect(() => {
    // Register static providers
    const unregisterNav = commandRegistry.registerProvider(navigationProvider);
    const unregisterSettings = commandRegistry.registerProvider(settingsProvider);
    const unregisterAdmin = commandRegistry.registerProvider(adminProvider);
    const unregisterContextual = commandRegistry.registerProvider(contextualProvider);

    // Register action provider with ref-based callbacks
    // Note: actionProvider already includes toggle-theme and sign-out commands
    const actionProvider = createActionProvider({
      resolvedTheme: themeRef.current ?? "light",
      setTheme: (theme) => setThemeRef.current(theme),
      logout: () => logoutRef.current(),
    });
    const unregisterActions = commandRegistry.registerProvider(actionProvider);

    return () => {
      unregisterNav();
      unregisterSettings();
      unregisterAdmin();
      unregisterContextual();
      unregisterActions();
    };
  }, []); // Empty deps - register ONCE on mount
}

// =============================================================================
// Component
// =============================================================================

export function CommandPalette() {
  const {
    isOpen,
    close,
    toggle,
    mode,
    setMode,
    search,
    setSearch,
    searchResults,
    isSearching,
    confirmingActionId,
    setConfirmingAction,
    actionStatus,
    setActionStatus,
  } = useCommandPalette();

  const ctx = useCommandContext();

  // Keep a stable ref to ctx for the subscription callback
  // This prevents re-subscribing when ctx object reference changes
  const ctxRef = useRef(ctx);
  useEffect(() => {
    ctxRef.current = ctx;
  });

  // Register providers
  useCommandProviders();

  // Register search providers
  useSearchProviders();

  // Enable async search
  useCommandSearch();

  // Get commands from registry
  const [commands, setCommands] = useState<Command[]>([]);

  // Subscribe to registry changes ONCE on mount
  // Use ref to always get current ctx in callback
  useEffect(() => {
    // Initial load
    setCommands(commandRegistry.getCommands(ctxRef.current));

    // Subscribe to registry changes
    const unsubscribe = commandRegistry.subscribe(() => {
      setCommands(commandRegistry.getCommands(ctxRef.current));
    });

    return unsubscribe;
  }, []); // Empty deps - subscribe once

  // Update commands when pathname changes (for route-based filtering)
  useEffect(() => {
    setCommands(commandRegistry.getCommands(ctx));
  }, [ctx.pathname]); // Only re-filter when route changes

  // Register Cmd+K shortcut
  useKeyboardShortcuts({
    shortcuts: [
      {
        key: "k",
        meta: true,
        handler: () => toggle(),
        description: "Open command palette",
      },
    ],
  });

  // Group commands by category
  const groupedCommands = useMemo<GroupedCommands>(() => {
    const groups: GroupedCommands = {
      navigation: [],
      actions: [],
      settings: [],
      admin: [],
      contextual: [],
    };

    for (const cmd of commands) {
      if (cmd.group in groups) {
        groups[cmd.group as keyof GroupedCommands].push(cmd);
      }
    }

    return groups;
  }, [commands]);

  // Handle command execution
  const runCommand = useCallback(
    async (command: Command) => {
      // Check if command requires confirmation
      if (command.requiresConfirmation && confirmingActionId !== command.id) {
        setConfirmingAction(command.id);
        return;
      }

      // Reset confirmation state
      setConfirmingAction(null);

      // Execute command
      try {
        setActionStatus("loading");
        await command.action(ctx);
        setActionStatus("success");

        // Show success toast for actions
        if (command.group === "actions" || command.group === "admin") {
          toast.success(`${command.label} completed`);
        }

        // Close palette after successful action
        setTimeout(() => {
          close();
          setActionStatus("idle");
        }, 150);
      } catch (error) {
        setActionStatus("error");
        toast.error(`Failed: ${error instanceof Error ? error.message : "Unknown error"}`);
      }
    },
    [ctx, close, confirmingActionId, setConfirmingAction, setActionStatus]
  );

  // Handle mode-specific rendering
  const getModeTitle = () => {
    switch (mode) {
      case "impersonate":
        return "Impersonate User";
      case "flag-toggle":
        return "Toggle Feature Flag";
      default:
        return "Command Palette";
    }
  };

  const getModeDescription = () => {
    switch (mode) {
      case "impersonate":
        return "Search for a user to impersonate";
      case "flag-toggle":
        return "Search for a feature flag to toggle";
      default:
        return "Search for commands, pages, and actions";
    }
  };

  const ModeIcon = useMemo(() => {
    switch (mode) {
      case "impersonate":
        return UserCog;
      case "flag-toggle":
        return ToggleLeft;
      default:
        return Search;
    }
  }, [mode]);

  // Render command item
  const renderCommandItem = (command: Command) => {
    const Icon = command.icon;
    const isConfirming = confirmingActionId === command.id;
    const isLoading = actionStatus === "loading" && isConfirming;

    return (
      <CommandItem
        key={command.id}
        value={`${command.label} ${command.keywords?.join(" ") || ""}`}
        onSelect={() => runCommand(command)}
        className={isConfirming ? "bg-destructive/10" : undefined}
      >
        {isLoading ? (
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
        ) : isConfirming ? (
          <AlertTriangle className="text-destructive mr-2 h-4 w-4" />
        ) : Icon ? (
          <Icon className="mr-2 h-4 w-4" />
        ) : null}

        <span className="flex-1">
          {isConfirming ? command.confirmationMessage || `Confirm: ${command.label}?` : command.label}
        </span>

        {command.shortcut && !isConfirming && <CommandShortcut>{formatShortcut(command.shortcut)}</CommandShortcut>}

        {isConfirming && <span className="text-muted-foreground text-xs">Enter to confirm</span>}
      </CommandItem>
    );
  };

  // Render back button for sub-modes
  const renderBackButton = () => {
    if (mode === "default") return null;

    return (
      <CommandItem
        value="back"
        onSelect={() => setMode("default")}
        className="text-muted-foreground"
      >
        <ArrowLeft className="mr-2 h-4 w-4" />
        <span>Back to commands</span>
        <CommandShortcut>Esc</CommandShortcut>
      </CommandItem>
    );
  };

  // Handle search result execution
  const runSearchResult = useCallback(
    async (result: SearchResult) => {
      try {
        setActionStatus("loading");
        await result.action(ctx);
        setActionStatus("success");

        // Show success toast
        if (result.type === "feature_flag") {
          const isEnabled = !(result.metadata?.enabled as boolean);
          toast.success(`${result.title} ${isEnabled ? "enabled" : "disabled"}`);
        } else {
          toast.success(`${result.title} selected`);
        }

        // Close palette after successful action
        setTimeout(() => {
          close();
          setMode("default");
          setActionStatus("idle");
        }, 150);
      } catch (error) {
        setActionStatus("error");
        toast.error(`Failed: ${error instanceof Error ? error.message : "Unknown error"}`);
      }
    },
    [ctx, close, setMode, setActionStatus]
  );

  // Render search result item
  const renderSearchResult = (result: SearchResult) => {
    const Icon = result.icon;

    return (
      <CommandItem
        key={result.id}
        value={`${result.title} ${result.subtitle || ""}`}
        onSelect={() => runSearchResult(result)}
      >
        {Icon ? <Icon className="mr-2 h-4 w-4" /> : <Search className="mr-2 h-4 w-4" />}
        <div className="flex flex-1 flex-col">
          <span>{result.title}</span>
          {result.subtitle && <span className="text-muted-foreground text-xs">{result.subtitle}</span>}
        </div>
        {result.type === "feature_flag" && (
          <span className={`text-xs ${result.metadata?.enabled ? "text-green-500" : "text-muted-foreground"}`}>
            {result.metadata?.enabled ? "ON" : "OFF"}
          </span>
        )}
      </CommandItem>
    );
  };

  return (
    <CommandDialog
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) {
          close();
          setMode("default");
          setConfirmingAction(null);
        }
      }}
      title={getModeTitle()}
      description={getModeDescription()}
    >
      <CommandInput
        placeholder={
          mode === "impersonate"
            ? "Search users by name or email..."
            : mode === "flag-toggle"
              ? "Search feature flags..."
              : "Type a command or search..."
        }
        value={search}
        onValueChange={setSearch}
      />
      <CommandList>
        <CommandEmpty>
          <div className="flex flex-col items-center gap-2 py-4">
            <ModeIcon className="text-muted-foreground h-10 w-10" />
            <p>No results found.</p>
            <p className="text-muted-foreground text-sm">
              {mode === "impersonate"
                ? "Try searching by name or email"
                : mode === "flag-toggle"
                  ? "No matching feature flags"
                  : "Try searching for something else."}
            </p>
          </div>
        </CommandEmpty>

        {/* Back button for sub-modes */}
        {mode !== "default" && (
          <>
            <CommandGroup>{renderBackButton()}</CommandGroup>
            <CommandSeparator />
          </>
        )}

        {/* Default mode: Show all command groups */}
        {mode === "default" && (
          <>
            {/* Contextual commands first (page-specific) */}
            {groupedCommands.contextual.length > 0 && (
              <>
                <CommandGroup heading="Page Actions">{groupedCommands.contextual.map(renderCommandItem)}</CommandGroup>
                <CommandSeparator />
              </>
            )}

            {/* Navigation */}
            <CommandGroup heading="Navigation">{groupedCommands.navigation.map(renderCommandItem)}</CommandGroup>

            <CommandSeparator />

            {/* Actions */}
            <CommandGroup heading="Actions">{groupedCommands.actions.map(renderCommandItem)}</CommandGroup>

            <CommandSeparator />

            {/* Settings */}
            <CommandGroup heading="Settings">{groupedCommands.settings.map(renderCommandItem)}</CommandGroup>

            {/* Admin (only if there are admin commands) */}
            {groupedCommands.admin.length > 0 && (
              <>
                <CommandSeparator />
                <CommandGroup heading="Administration">{groupedCommands.admin.map(renderCommandItem)}</CommandGroup>
              </>
            )}
          </>
        )}

        {/* Impersonate mode: Show user search results */}
        {mode === "impersonate" && (
          <CommandGroup heading="Users">
            {isSearching && (
              <CommandItem
                value="loading"
                disabled
              >
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                <span>Searching users...</span>
              </CommandItem>
            )}
            {!isSearching && search.length < 2 && (
              <CommandItem
                value="hint"
                disabled
                className="text-muted-foreground"
              >
                <User className="mr-2 h-4 w-4" />
                <span>Type at least 2 characters to search users</span>
              </CommandItem>
            )}
            {!isSearching && search.length >= 2 && searchResults.length === 0 && (
              <CommandItem
                value="no-results"
                disabled
                className="text-muted-foreground"
              >
                <Search className="mr-2 h-4 w-4" />
                <span>No users found matching "{search}"</span>
              </CommandItem>
            )}
            {!isSearching && searchResults.filter((r) => r.type === "user").map(renderSearchResult)}
          </CommandGroup>
        )}

        {/* Flag toggle mode: Show feature flag search results */}
        {mode === "flag-toggle" && (
          <CommandGroup heading="Feature Flags">
            {isSearching && (
              <CommandItem
                value="loading"
                disabled
              >
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                <span>Searching flags...</span>
              </CommandItem>
            )}
            {!isSearching && search.length === 0 && (
              <CommandItem
                value="hint"
                disabled
                className="text-muted-foreground"
              >
                <ToggleLeft className="mr-2 h-4 w-4" />
                <span>Type to search feature flags</span>
              </CommandItem>
            )}
            {!isSearching &&
              search.length > 0 &&
              searchResults.filter((r) => r.type === "feature_flag").length === 0 && (
                <CommandItem
                  value="no-results"
                  disabled
                  className="text-muted-foreground"
                >
                  <Search className="mr-2 h-4 w-4" />
                  <span>No flags found matching "{search}"</span>
                </CommandItem>
              )}
            {!isSearching && searchResults.filter((r) => r.type === "feature_flag").map(renderSearchResult)}
          </CommandGroup>
        )}
      </CommandList>

      {/* Footer with keyboard hints */}
      <div className="text-muted-foreground border-t px-4 py-2 text-xs">
        <span>↑↓ to navigate</span>
        <span className="mx-2">•</span>
        <span>↵ to select</span>
        <span className="mx-2">•</span>
        <span>esc to close</span>
      </div>
    </CommandDialog>
  );
}
