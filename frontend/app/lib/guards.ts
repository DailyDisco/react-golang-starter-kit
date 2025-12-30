import { redirect } from "@tanstack/react-router";

import type { User } from "../services/types";

/**
 * Route Guards for TanStack Router
 *
 * These functions can be used in route `beforeLoad` hooks to enforce
 * authentication and authorization requirements.
 */

export type UserRole = "user" | "premium" | "admin" | "super_admin";

/**
 * Parses and validates the stored user from localStorage
 * Note: Authentication is verified via httpOnly cookies when making API calls.
 * The user data in localStorage is only for UI purposes.
 */
function getStoredUser(): User | null {
  const storedUser = localStorage.getItem("auth_user");

  if (!storedUser) {
    return null;
  }

  try {
    return JSON.parse(storedUser) as User;
  } catch {
    // Invalid user data - clear auth storage
    localStorage.removeItem("auth_user");
    return null;
  }
}

/**
 * Requires the user to be authenticated.
 * Redirects to /login if not authenticated.
 *
 * @param redirectTo - Optional path to redirect to after login
 * @returns The authenticated user
 *
 * @example
 * beforeLoad: async () => {
 *   return requireAuth();
 * }
 */
export function requireAuth(redirectTo?: string): { user: User } {
  const user = getStoredUser();

  if (!user) {
    throw redirect({
      to: "/login",
      search: redirectTo ? { redirect: redirectTo } : undefined,
    });
  }

  return { user };
}

/**
 * Requires the user to have one of the specified roles.
 * Redirects to /login if not authenticated, or / if not authorized.
 *
 * @param allowedRoles - Array of roles that are allowed access
 * @param options - Optional configuration
 * @returns The authenticated user with verified role
 *
 * @example
 * beforeLoad: async () => {
 *   return requireRole(["admin", "super_admin"]);
 * }
 */
export function requireRole(
  allowedRoles: UserRole[],
  options?: {
    redirectTo?: string;
    unauthorizedRedirect?: string;
  }
): { user: User } {
  const user = getStoredUser();

  // Not authenticated
  if (!user) {
    throw redirect({
      to: "/login",
      search: options?.redirectTo ? { redirect: options.redirectTo } : undefined,
    });
  }

  // Check role
  if (!user.role || !allowedRoles.includes(user.role as UserRole)) {
    // Authenticated but not authorized
    throw redirect({
      to: options?.unauthorizedRedirect || "/",
    });
  }

  return { user };
}

/**
 * Requires admin-level access (admin or super_admin).
 * Convenience wrapper around requireRole.
 *
 * @example
 * beforeLoad: async () => {
 *   return requireAdmin();
 * }
 */
export function requireAdmin(): { user: User } {
  return requireRole(["admin", "super_admin"], {
    redirectTo: "/admin",
    unauthorizedRedirect: "/",
  });
}

/**
 * Requires super admin access only.
 * Convenience wrapper around requireRole.
 *
 * @example
 * beforeLoad: async () => {
 *   return requireSuperAdmin();
 * }
 */
export function requireSuperAdmin(): { user: User } {
  return requireRole(["super_admin"], {
    redirectTo: "/admin",
    unauthorizedRedirect: "/admin",
  });
}
