import type { QueryClient } from "@tanstack/react-query";
import { redirect } from "@tanstack/react-router";

import type { User } from "../services/types";
import { currentUserQueryOptions } from "./route-query-options";

/**
 * Route Guards for TanStack Router
 *
 * These functions are used in route `beforeLoad` hooks to enforce
 * authentication and authorization requirements.
 *
 * Guards are async and verify the session with the backend using
 * TanStack Query's ensureQueryData for efficient caching.
 */

export type UserRole = "user" | "premium" | "admin" | "super_admin";

/**
 * Context passed to beforeLoad hooks by the router
 */
interface BeforeLoadContext {
  context: {
    queryClient: QueryClient;
  };
  location: {
    pathname: string;
  };
}

/**
 * Parses and validates the stored user from localStorage
 * Used as a fast fallback when backend verification is not needed.
 */
function getStoredUser(): User | null {
  if (typeof window === "undefined") return null;

  const storedUser = localStorage.getItem("auth_user");

  if (!storedUser) {
    return null;
  }

  try {
    return JSON.parse(storedUser) as User;
  } catch {
    localStorage.removeItem("auth_user");
    return null;
  }
}

/**
 * Requires the user to be authenticated.
 * Verifies the session with the backend using TanStack Query.
 * Redirects to /login if not authenticated.
 *
 * @param ctx - The beforeLoad context from the router
 * @param redirectTo - Optional path to redirect to after login
 * @returns The authenticated user
 *
 * @example
 * beforeLoad: async (ctx) => {
 *   return requireAuth(ctx);
 * }
 */
export async function requireAuth(ctx: BeforeLoadContext, redirectTo?: string): Promise<{ user: User }> {
  const { queryClient } = ctx.context;

  // First check localStorage for cached user (fast path)
  const storedUser = getStoredUser();
  if (!storedUser) {
    throw redirect({
      to: "/login",
      search: redirectTo ? { redirect: redirectTo } : undefined,
    });
  }

  try {
    // Verify with backend - uses cache if available and not stale
    const user = await queryClient.ensureQueryData(currentUserQueryOptions());
    return { user };
  } catch {
    // Session is invalid - clear local state and redirect
    localStorage.removeItem("auth_user");
    throw redirect({
      to: "/login",
      search: redirectTo ? { redirect: redirectTo } : undefined,
    });
  }
}

/**
 * Requires the user to have one of the specified roles.
 * Verifies the session with the backend and checks role.
 * Redirects to /login if not authenticated, or / if not authorized.
 *
 * @param ctx - The beforeLoad context from the router
 * @param allowedRoles - Array of roles that are allowed access
 * @param options - Optional configuration
 * @returns The authenticated user with verified role
 *
 * @example
 * beforeLoad: async (ctx) => {
 *   return requireRole(ctx, ["admin", "super_admin"]);
 * }
 */
export async function requireRole(
  ctx: BeforeLoadContext,
  allowedRoles: UserRole[],
  options?: {
    redirectTo?: string;
    unauthorizedRedirect?: string;
  }
): Promise<{ user: User }> {
  const { queryClient } = ctx.context;

  // First check localStorage for cached user (fast path)
  const storedUser = getStoredUser();
  if (!storedUser) {
    throw redirect({
      to: "/login",
      search: options?.redirectTo ? { redirect: options.redirectTo } : undefined,
    });
  }

  try {
    // Verify with backend
    const user = await queryClient.ensureQueryData(currentUserQueryOptions());

    // Check role
    if (!user.role || !allowedRoles.includes(user.role as UserRole)) {
      throw redirect({
        to: options?.unauthorizedRedirect || "/",
      });
    }

    return { user };
  } catch (error) {
    // Check if this is already a redirect
    if (error instanceof Response || (error as { code?: string })?.code === "REDIRECT") {
      throw error;
    }

    // Session is invalid - clear local state and redirect
    localStorage.removeItem("auth_user");
    throw redirect({
      to: "/login",
      search: options?.redirectTo ? { redirect: options.redirectTo } : undefined,
    });
  }
}

/**
 * Requires admin-level access (admin or super_admin).
 * Convenience wrapper around requireRole.
 *
 * @example
 * beforeLoad: async (ctx) => {
 *   return requireAdmin(ctx);
 * }
 */
export async function requireAdmin(ctx: BeforeLoadContext): Promise<{ user: User }> {
  return requireRole(ctx, ["admin", "super_admin"], {
    redirectTo: "/admin",
    unauthorizedRedirect: "/",
  });
}

/**
 * Requires super admin access only.
 * Convenience wrapper around requireRole.
 *
 * @example
 * beforeLoad: async (ctx) => {
 *   return requireSuperAdmin(ctx);
 * }
 */
export async function requireSuperAdmin(ctx: BeforeLoadContext): Promise<{ user: User }> {
  return requireRole(ctx, ["super_admin"], {
    redirectTo: "/admin",
    unauthorizedRedirect: "/admin",
  });
}

/**
 * Sync version of requireAuth for cases where async is not needed
 * Uses only localStorage - does NOT verify with backend
 *
 * @deprecated Prefer async requireAuth for proper session validation
 */
export function requireAuthSync(redirectTo?: string): { user: User } {
  const user = getStoredUser();

  if (!user) {
    throw redirect({
      to: "/login",
      search: redirectTo ? { redirect: redirectTo } : undefined,
    });
  }

  return { user };
}
