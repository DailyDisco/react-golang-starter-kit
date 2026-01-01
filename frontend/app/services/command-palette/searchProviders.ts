import { AdminService, type FeatureFlag } from "@/services/admin/adminService";
import { FileText, Flag, ToggleLeft, ToggleRight, User } from "lucide-react";

import type { CommandContext, SearchProvider, SearchResult } from "./types";

// =============================================================================
// User Search Provider (Admin only)
// =============================================================================

/**
 * User search provider - currently returns placeholder results
 * TODO: Implement when /admin/users endpoint is available
 */
export const userSearchProvider: SearchProvider = {
  id: "users",
  name: "Users",
  types: ["user"],
  roles: ["admin", "super_admin"],
  debounceMs: 300,
  minQueryLength: 2,

  search: async (query: string, ctx: CommandContext): Promise<SearchResult[]> => {
    // User search API not yet implemented
    // Return empty results for now - the UI will show "No users found"
    // TODO: Implement when GET /admin/users endpoint is available
    console.info("User search not yet implemented. Query:", query);
    return [];
  },
};

// =============================================================================
// Feature Flag Search Provider (Admin only)
// =============================================================================

export const featureFlagSearchProvider: SearchProvider = {
  id: "feature-flags",
  name: "Feature Flags",
  types: ["feature_flag"],
  roles: ["admin", "super_admin"],
  debounceMs: 200,
  minQueryLength: 1,

  search: async (query: string, _ctx: CommandContext): Promise<SearchResult[]> => {
    try {
      const response = await AdminService.getFeatureFlags();
      const lowerQuery = query.toLowerCase();

      // Filter flags client-side since we have them all
      const matchingFlags = response.flags
        .filter(
          (flag: FeatureFlag) =>
            flag.name.toLowerCase().includes(lowerQuery) ||
            flag.key.toLowerCase().includes(lowerQuery) ||
            flag.description?.toLowerCase().includes(lowerQuery)
        )
        .slice(0, 5);

      return matchingFlags.map((flag: FeatureFlag) => ({
        id: `flag-${flag.id}`,
        type: "feature_flag" as const,
        title: flag.name,
        subtitle: `${flag.key} • ${flag.enabled ? "Enabled" : "Disabled"}`,
        icon: flag.enabled ? ToggleRight : ToggleLeft,
        action: async () => {
          // Toggle the flag
          try {
            await AdminService.updateFeatureFlag(flag.key, {
              enabled: !flag.enabled,
            });
            // The component will show a toast
          } catch (error) {
            throw new Error(`Failed to toggle ${flag.name}`);
          }
        },
        metadata: {
          flag,
          enabled: flag.enabled,
          key: flag.key,
        },
      }));
    } catch (error) {
      console.error("Feature flag search failed:", error);
      return [];
    }
  },
};

// =============================================================================
// Audit Log Search Provider (Admin only)
// =============================================================================

export const auditLogSearchProvider: SearchProvider = {
  id: "audit-logs",
  name: "Audit Logs",
  types: ["audit_log"],
  roles: ["admin", "super_admin"],
  debounceMs: 300,
  minQueryLength: 3,

  search: async (query: string, _ctx: CommandContext): Promise<SearchResult[]> => {
    try {
      const response = await AdminService.getAuditLogs({
        action: query,
        page: 1,
        limit: 5,
      });

      return response.logs.map((log) => ({
        id: `audit-${log.id}`,
        type: "audit_log" as const,
        title: `${log.action} - ${log.target_type || "system"}`,
        subtitle: `${log.user_name || "System"} • ${formatRelativeTime(log.created_at)}`,
        icon: FileText,
        action: (ctx: CommandContext) => {
          ctx.navigate(`/admin/audit-logs?highlight=${log.id}`);
        },
        metadata: {
          log,
          action: log.action,
          userId: log.user_id,
        },
      }));
    } catch (error) {
      console.error("Audit log search failed:", error);
      return [];
    }
  },
};

// =============================================================================
// Static Page Search Provider (All users)
// =============================================================================

const staticPages = [
  { path: "/dashboard", title: "Dashboard", keywords: ["home", "main", "overview"] },
  { path: "/billing", title: "Billing", keywords: ["payment", "subscription", "plan"] },
  { path: "/settings/profile", title: "Profile Settings", keywords: ["account", "me"] },
  {
    path: "/settings/security",
    title: "Security Settings",
    keywords: ["password", "2fa"],
  },
  {
    path: "/settings/preferences",
    title: "Preferences",
    keywords: ["theme", "language"],
  },
  {
    path: "/settings/notifications",
    title: "Notification Settings",
    keywords: ["alerts", "email"],
  },
  { path: "/settings/privacy", title: "Privacy Settings", keywords: ["data", "gdpr"] },
  {
    path: "/settings/login-history",
    title: "Login History",
    keywords: ["sessions", "activity"],
  },
  {
    path: "/settings/connected-accounts",
    title: "Connected Accounts",
    keywords: ["oauth", "social"],
  },
];

const adminPages = [
  { path: "/admin", title: "Admin Dashboard", keywords: ["overview", "stats"] },
  { path: "/admin/users", title: "User Management", keywords: ["accounts", "members"] },
  { path: "/admin/audit-logs", title: "Audit Logs", keywords: ["history", "activity"] },
  {
    path: "/admin/feature-flags",
    title: "Feature Flags",
    keywords: ["toggles", "features"],
  },
  { path: "/admin/health", title: "System Health", keywords: ["status", "monitoring"] },
  {
    path: "/admin/announcements",
    title: "Announcements",
    keywords: ["banners", "messages"],
  },
  {
    path: "/admin/email-templates",
    title: "Email Templates",
    keywords: ["emails", "templates"],
  },
  {
    path: "/admin/settings",
    title: "Admin Settings",
    keywords: ["configuration", "system"],
  },
];

export const pageSearchProvider: SearchProvider = {
  id: "pages",
  name: "Pages",
  types: ["page"],
  debounceMs: 0, // Instant - it's all client-side
  minQueryLength: 1,

  search: async (query: string, ctx: CommandContext): Promise<SearchResult[]> => {
    const lowerQuery = query.toLowerCase();
    const isAdmin = ctx.user?.role === "admin" || ctx.user?.role === "super_admin";

    const allPages = isAdmin ? [...staticPages, ...adminPages] : staticPages;

    const matchingPages = allPages
      .filter(
        (page) => page.title.toLowerCase().includes(lowerQuery) || page.keywords.some((kw) => kw.includes(lowerQuery))
      )
      .slice(0, 5);

    return matchingPages.map((page) => ({
      id: `page-${page.path}`,
      type: "page" as const,
      title: page.title,
      subtitle: page.path,
      icon: Flag, // Generic page icon
      action: (actionCtx: CommandContext) => {
        actionCtx.navigate(page.path);
      },
      metadata: {
        path: page.path,
        keywords: page.keywords,
      },
    }));
  },
};

// =============================================================================
// Helpers
// =============================================================================

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "just now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
}
