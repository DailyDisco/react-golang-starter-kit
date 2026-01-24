import type { AuditLogFilter } from "../services/admin/adminService";

// Query Keys Factory for Type-Safe Query Management
export const queryKeys = {
  // User management
  users: {
    all: ["users"] as const,
    lists: () => [...queryKeys.users.all, "list"] as const,
    list: (filters: Record<string, unknown>) => [...queryKeys.users.lists(), filters] as const,
    details: () => [...queryKeys.users.all, "detail"] as const,
    detail: (id: number) => [...queryKeys.users.details(), id] as const,
  },

  // Authentication
  auth: {
    all: ["auth"] as const,
    user: ["auth", "user"] as const,
    session: ["auth", "session"] as const,
  },

  // Health checks
  health: {
    status: ["health", "status"] as const,
  },

  // Feature flags
  featureFlags: {
    all: ["featureFlags"] as const,
    user: () => [...queryKeys.featureFlags.all, "user"] as const,
  },

  // Billing & subscriptions
  billing: {
    all: ["billing"] as const,
    config: () => [...queryKeys.billing.all, "config"] as const,
    plans: () => [...queryKeys.billing.all, "plans"] as const,
    subscription: () => [...queryKeys.billing.all, "subscription"] as const,
  },

  // File management
  files: {
    all: ["files"] as const,
    list: (limit?: number, offset?: number) => [...queryKeys.files.all, { limit, offset }] as const,
    url: (fileId: number) => [...queryKeys.files.all, "url", fileId] as const,
    storageStatus: () => [...queryKeys.files.all, "storage-status"] as const,
  },

  // Settings & preferences
  settings: {
    all: ["settings"] as const,
    preferences: () => [...queryKeys.settings.all, "preferences"] as const,
    dataExportStatus: () => [...queryKeys.settings.all, "data-export-status"] as const,
    sessions: () => [...queryKeys.settings.all, "sessions"] as const,
    apiKeys: () => [...queryKeys.settings.all, "api-keys"] as const,
    loginHistory: () => [...queryKeys.settings.all, "login-history"] as const,
    connectedAccounts: () => [...queryKeys.settings.all, "connected-accounts"] as const,
  },

  // Changelog / announcements
  changelog: {
    all: ["changelog"] as const,
    entries: (page: number, limit: number, category?: string) =>
      [...queryKeys.changelog.all, { page, limit, category }] as const,
  },

  // Organizations
  organizations: {
    all: ["organizations"] as const,
    list: () => [...queryKeys.organizations.all, "list"] as const,
    detail: (orgSlug: string) => [...queryKeys.organizations.all, orgSlug, "detail"] as const,
    members: (orgSlug: string) => [...queryKeys.organizations.all, orgSlug, "members"] as const,
    invitations: (orgSlug: string) => [...queryKeys.organizations.all, orgSlug, "invitations"] as const,
    billing: (orgSlug: string) => [...queryKeys.organizations.all, orgSlug, "billing"] as const,
  },

  // Usage metering
  usage: {
    all: ["usage"] as const,
    current: () => [...queryKeys.usage.all, "current"] as const,
    history: (months?: number) => [...queryKeys.usage.all, "history", { months }] as const,
    alerts: () => [...queryKeys.usage.all, "alerts"] as const,
  },

  // Audit logs & activity
  auditLogs: {
    all: ["auditLogs"] as const,
    admin: (filter: AuditLogFilter) => [...queryKeys.auditLogs.all, "admin", filter] as const,
    myActivity: (limit?: number) => [...queryKeys.auditLogs.all, "my-activity", { limit }] as const,
  },

  // Notifications
  notifications: {
    all: ["notifications"] as const,
    list: (params?: { page?: number; unread?: boolean }) => [...queryKeys.notifications.all, "list", params] as const,
    unreadCount: () => [...queryKeys.notifications.all, "unread-count"] as const,
  },
} as const;

// Type-safe query key inference
export type QueryKeys = typeof queryKeys;
