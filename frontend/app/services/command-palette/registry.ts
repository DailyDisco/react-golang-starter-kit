import type { Command, CommandContext, CommandProvider, SearchProvider, SearchResult } from "./types";

/**
 * Matches a route pattern against a pathname
 * Supports:
 * - Exact matches: "/admin/users"
 * - Wildcards: "/admin/*" matches "/admin/anything"
 * - Params: "/admin/users/:id" matches "/admin/users/123"
 */
function matchRoute(pattern: string, pathname: string): boolean {
  const patternParts = pattern.split("/").filter(Boolean);
  const pathParts = pathname.split("/").filter(Boolean);

  // Wildcard at end matches everything after
  if (patternParts[patternParts.length - 1] === "*") {
    const basePattern = patternParts.slice(0, -1);
    if (pathParts.length < basePattern.length) return false;
    return basePattern.every((part, i) => part === pathParts[i] || part.startsWith(":"));
  }

  // Exact length match required
  if (patternParts.length !== pathParts.length) return false;

  return patternParts.every((part, i) => {
    if (part.startsWith(":")) return true; // Param matches anything
    return part === pathParts[i];
  });
}

/**
 * Check if user has required role
 */
function hasRole(userRole: string | undefined, requiredRoles: string[] | undefined): boolean {
  if (!requiredRoles || requiredRoles.length === 0) return true;
  if (!userRole) return false;

  const roleHierarchy: Record<string, number> = {
    user: 0,
    admin: 1,
    super_admin: 2,
  };

  const userLevel = roleHierarchy[userRole] ?? 0;

  // User has access if their level is >= any of the required roles
  return requiredRoles.some((role) => userLevel >= (roleHierarchy[role] ?? 0));
}

/**
 * Command Registry - Central hub for all commands and search providers
 *
 * Implements a plugin pattern allowing features to self-register commands.
 * Supports:
 * - Individual command registration
 * - Command providers (batch registration)
 * - Search providers (async data sources)
 * - Role-based filtering
 * - Route-based filtering
 */
class CommandRegistry {
  private commands: Map<string, Command> = new Map();
  private providers: Map<string, CommandProvider> = new Map();
  private searchProviders: Map<string, SearchProvider> = new Map();
  private listeners: Set<() => void> = new Set();

  // ===========================================================================
  // Command Registration
  // ===========================================================================

  /**
   * Register a single command
   * @returns Unregister function
   */
  register(command: Command): () => void {
    this.commands.set(command.id, command);
    this.notify();
    return () => this.unregister(command.id);
  }

  /**
   * Register multiple commands at once
   * @returns Unregister function for all
   */
  registerMany(commands: Command[]): () => void {
    commands.forEach((cmd) => this.commands.set(cmd.id, cmd));
    this.notify();
    return () => {
      commands.forEach((cmd) => this.commands.delete(cmd.id));
      this.notify();
    };
  }

  /**
   * Unregister a command by ID
   */
  unregister(id: string): void {
    this.commands.delete(id);
    this.notify();
  }

  // ===========================================================================
  // Provider Registration
  // ===========================================================================

  /**
   * Register a command provider (for dynamic/contextual commands)
   * @returns Unregister function
   */
  registerProvider(provider: CommandProvider): () => void {
    this.providers.set(provider.id, provider);
    this.notify();
    return () => this.unregisterProvider(provider.id);
  }

  /**
   * Unregister a command provider
   */
  unregisterProvider(id: string): void {
    this.providers.delete(id);
    this.notify();
  }

  // ===========================================================================
  // Search Provider Registration
  // ===========================================================================

  /**
   * Register a search provider for async data sources
   * @returns Unregister function
   */
  registerSearchProvider(provider: SearchProvider): () => void {
    this.searchProviders.set(provider.id, provider);
    this.notify();
    return () => this.unregisterSearchProvider(provider.id);
  }

  /**
   * Unregister a search provider
   */
  unregisterSearchProvider(id: string): void {
    this.searchProviders.delete(id);
    this.notify();
  }

  // ===========================================================================
  // Query Methods
  // ===========================================================================

  /**
   * Get all commands filtered for the current context
   */
  getCommands(ctx: CommandContext): Command[] {
    // Collect commands from static registrations and providers
    const allCommands = [
      ...this.commands.values(),
      ...Array.from(this.providers.values()).flatMap((p) => p.getCommands(ctx)),
    ];

    // Filter by role, route, and custom conditions
    return allCommands.filter((cmd) => {
      // Role-based filtering
      if (!hasRole(ctx.user?.role, cmd.roles)) {
        return false;
      }

      // Route pattern matching
      if (cmd.routePatterns && cmd.routePatterns.length > 0) {
        const matchesRoute = cmd.routePatterns.some((pattern) => matchRoute(pattern, ctx.pathname));
        if (!matchesRoute) return false;
      }

      // Custom condition
      if (cmd.condition && !cmd.condition(ctx)) {
        return false;
      }

      return true;
    });
  }

  /**
   * Get a specific command by ID
   */
  getCommand(id: string): Command | undefined {
    return this.commands.get(id);
  }

  /**
   * Get all registered search providers filtered by user role
   */
  getSearchProviders(ctx: CommandContext): SearchProvider[] {
    return Array.from(this.searchProviders.values()).filter((p) => hasRole(ctx.user?.role, p.roles));
  }

  /**
   * Search across all providers
   */
  async search(query: string, ctx: CommandContext, options?: { signal?: AbortSignal }): Promise<SearchResult[]> {
    const providers = this.getSearchProviders(ctx);

    const results = await Promise.allSettled(
      providers.map(async (provider) => {
        // Check minimum query length
        const minLength = provider.minQueryLength ?? 2;
        if (query.length < minLength) return [];

        // Check abort signal
        if (options?.signal?.aborted) {
          throw new DOMException("Aborted", "AbortError");
        }

        return provider.search(query, ctx);
      })
    );

    // Collect successful results, ignore failures
    return results
      .filter((r): r is PromiseFulfilledResult<SearchResult[]> => r.status === "fulfilled")
      .flatMap((r) => r.value);
  }

  // ===========================================================================
  // Subscription
  // ===========================================================================

  /**
   * Subscribe to registry changes
   * @returns Unsubscribe function
   */
  subscribe(listener: () => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  /**
   * Notify all listeners of changes
   */
  private notify(): void {
    this.listeners.forEach((listener) => listener());
  }

  // ===========================================================================
  // Debugging
  // ===========================================================================

  /**
   * Get registry stats for debugging
   */
  getStats(): {
    commands: number;
    providers: number;
    searchProviders: number;
  } {
    return {
      commands: this.commands.size,
      providers: this.providers.size,
      searchProviders: this.searchProviders.size,
    };
  }
}

// Singleton instance
export const commandRegistry = new CommandRegistry();

// Export the class for testing
export { CommandRegistry };
