import { useMemo } from "react";

import { useAuth } from "@/hooks/auth/useAuth";
import { useCommandPaletteStore } from "@/hooks/useCommandPalette";
import type { CommandContext } from "@/services/command-palette/types";
import { useLocation, useNavigate, useParams } from "@tanstack/react-router";

/**
 * Hook that provides the command execution context
 *
 * This context is passed to all command handlers and contains:
 * - Current user information
 * - Current route information
 * - Navigation function
 * - Palette control functions
 */
export function useCommandContext(): CommandContext {
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams({ strict: false });
  const { close, setMode, setSearch } = useCommandPaletteStore();

  return useMemo(
    () => ({
      user: user
        ? {
            id: user.id,
            email: user.email,
            name: user.name,
            role: user.role as "user" | "admin" | "super_admin",
          }
        : null,
      pathname: location.pathname,
      routeParams: (params as Record<string, string>) ?? {},
      navigate: (to: string) => {
        close();
        navigate({ to });
      },
      close,
      setMode,
      setSearch,
    }),
    [user, location.pathname, params, navigate, close, setMode, setSearch]
  );
}

/**
 * Hook that returns filtered commands based on current context
 */
export function useFilteredCommands(
  commands: Array<{
    id: string;
    roles?: Array<"user" | "admin" | "super_admin">;
    routePatterns?: string[];
    condition?: (ctx: CommandContext) => boolean;
  }>,
  ctx: CommandContext
) {
  return useMemo(() => {
    return commands.filter((cmd) => {
      // Role-based filtering
      if (cmd.roles && cmd.roles.length > 0) {
        if (!ctx.user?.role) return false;

        const roleHierarchy: Record<string, number> = {
          user: 0,
          admin: 1,
          super_admin: 2,
        };

        const userLevel = roleHierarchy[ctx.user.role] ?? 0;
        const hasRequiredRole = cmd.roles.some((role) => userLevel >= (roleHierarchy[role] ?? 0));

        if (!hasRequiredRole) return false;
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
  }, [commands, ctx]);
}

/**
 * Matches a route pattern against a pathname
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
    if (part.startsWith(":")) return true;
    return part === pathParts[i];
  });
}
