import {
  Activity,
  Bell,
  Clipboard,
  CreditCard,
  Download,
  FileText,
  Flag,
  HelpCircle,
  History,
  Home,
  Key,
  Keyboard,
  LayoutDashboard,
  Link2,
  LogOut,
  Mail,
  Moon,
  Palette,
  Receipt,
  Search,
  Settings,
  Shield,
  Sun,
  ToggleLeft,
  User,
  UserCog,
  UserPlus,
  Users,
} from "lucide-react";

import type { Command, CommandContext, CommandProvider } from "./types";

// =============================================================================
// Navigation Provider
// =============================================================================

export const navigationProvider: CommandProvider = {
  id: "navigation",
  name: "Navigation",
  getCommands: (ctx: CommandContext): Command[] => [
    {
      id: "goto-dashboard",
      label: "Go to Dashboard",
      icon: Home,
      keywords: ["home", "main"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/dashboard"),
    },
    {
      id: "goto-billing",
      label: "Go to Billing",
      icon: CreditCard,
      keywords: ["payment", "subscription", "plan"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/billing"),
    },
    {
      id: "goto-profile",
      label: "Go to Profile",
      icon: User,
      keywords: ["account", "me"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/profile"),
    },
    {
      id: "goto-security",
      label: "Go to Security Settings",
      icon: Shield,
      keywords: ["password", "2fa", "authentication"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/security"),
    },
    {
      id: "goto-preferences",
      label: "Go to Preferences",
      icon: Palette,
      keywords: ["theme", "language"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/preferences"),
    },
    {
      id: "goto-notifications",
      label: "Go to Notifications",
      icon: Bell,
      keywords: ["alerts", "email"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/notifications"),
    },
    {
      id: "goto-privacy",
      label: "Go to Privacy Settings",
      icon: Key,
      keywords: ["data", "gdpr"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/privacy"),
    },
    {
      id: "goto-login-history",
      label: "Go to Login History",
      icon: History,
      keywords: ["sessions", "activity"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/login-history"),
    },
    {
      id: "goto-connected-accounts",
      label: "Go to Connected Accounts",
      icon: Link2,
      keywords: ["oauth", "social", "google", "github"],
      group: "navigation",
      action: (ctx) => ctx.navigate("/settings/connected-accounts"),
    },
  ],
};

// =============================================================================
// Action Provider
// =============================================================================

/**
 * Theme type from theme-provider
 */
type Theme = "light" | "dark" | "system";

/**
 * Factory function to create action provider with theme state
 */
export function createActionProvider(options: {
  resolvedTheme: string;
  setTheme: (theme: Theme) => void;
  logout: () => void;
}): CommandProvider {
  return {
    id: "actions",
    name: "Actions",
    getCommands: (): Command[] => [
      {
        id: "toggle-theme",
        label: options.resolvedTheme === "dark" ? "Switch to Light Mode" : "Switch to Dark Mode",
        icon: options.resolvedTheme === "dark" ? Sun : Moon,
        keywords: ["theme", "dark", "light", "mode"],
        group: "actions",
        action: () => options.setTheme(options.resolvedTheme === "dark" ? "light" : "dark"),
      },
      {
        id: "copy-url",
        label: "Copy Current URL",
        icon: Clipboard,
        keywords: ["link", "share", "clipboard"],
        group: "actions",
        action: async () => {
          await navigator.clipboard.writeText(window.location.href);
        },
      },
      {
        id: "keyboard-shortcuts",
        label: "Show Keyboard Shortcuts",
        icon: Keyboard,
        shortcut: { key: "?" },
        keywords: ["help", "keys", "hotkeys"],
        group: "actions",
        action: (ctx) => {
          // Dispatch event for keyboard shortcuts modal
          window.dispatchEvent(new CustomEvent("command:keyboard-shortcuts"));
          ctx.close();
        },
      },
      {
        id: "sign-out",
        label: "Sign Out",
        icon: LogOut,
        keywords: ["logout", "exit"],
        group: "actions",
        requiresConfirmation: false,
        action: () => options.logout(),
      },
    ],
  };
}

// Static action provider for when theme/logout aren't available
export const actionProvider: CommandProvider = {
  id: "actions",
  name: "Actions",
  getCommands: (): Command[] => [
    {
      id: "copy-url",
      label: "Copy Current URL",
      icon: Clipboard,
      keywords: ["link", "share", "clipboard"],
      group: "actions",
      action: () => {
        navigator.clipboard.writeText(window.location.href);
      },
    },
    {
      id: "keyboard-shortcuts",
      label: "Show Keyboard Shortcuts",
      icon: Keyboard,
      shortcut: { key: "?" },
      keywords: ["help", "keys", "hotkeys"],
      group: "actions",
      action: (ctx) => {
        window.dispatchEvent(new CustomEvent("command:keyboard-shortcuts"));
        ctx.close();
      },
    },
  ],
};

// =============================================================================
// Settings Provider
// =============================================================================

export const settingsProvider: CommandProvider = {
  id: "settings",
  name: "Settings",
  getCommands: (): Command[] => [
    {
      id: "open-settings",
      label: "Open Settings",
      icon: Settings,
      shortcut: { key: ",", meta: true },
      keywords: ["preferences", "options", "config"],
      group: "settings",
      action: (ctx) => ctx.navigate("/settings"),
    },
  ],
};

// =============================================================================
// Admin Provider
// =============================================================================

export const adminProvider: CommandProvider = {
  id: "admin",
  name: "Administration",
  getCommands: (): Command[] => [
    // Navigation
    {
      id: "admin-overview",
      label: "Admin Overview",
      icon: LayoutDashboard,
      keywords: ["admin", "dashboard"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin"),
    },
    {
      id: "admin-users",
      label: "Manage Users",
      icon: Users,
      keywords: ["admin", "users", "accounts"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/users"),
    },
    {
      id: "admin-audit",
      label: "View Audit Logs",
      icon: FileText,
      keywords: ["admin", "logs", "history", "audit"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/audit-logs"),
    },
    {
      id: "admin-flags",
      label: "Feature Flags",
      icon: Flag,
      keywords: ["admin", "features", "toggles"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/feature-flags"),
    },
    {
      id: "admin-health",
      label: "System Health",
      icon: Activity,
      keywords: ["admin", "status", "monitoring"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/health"),
    },
    {
      id: "admin-announcements",
      label: "Announcements",
      icon: Bell,
      keywords: ["admin", "banners", "messages"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/announcements"),
    },
    {
      id: "admin-email",
      label: "Email Templates",
      icon: Mail,
      keywords: ["admin", "templates", "email"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/email-templates"),
    },
    {
      id: "admin-settings",
      label: "Admin Settings",
      icon: Settings,
      keywords: ["admin", "configuration"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => ctx.navigate("/admin/settings"),
    },

    // Admin Actions
    {
      id: "admin-impersonate",
      label: "Impersonate User...",
      description: "View the app as another user",
      icon: UserCog,
      keywords: ["admin", "impersonate", "switch", "as", "view as"],
      group: "admin",
      roles: ["super_admin"],
      action: (ctx) => {
        ctx.setMode("impersonate");
        ctx.setSearch("");
      },
    },
    {
      id: "admin-toggle-flag",
      label: "Toggle Feature Flag...",
      description: "Quickly enable or disable a feature flag",
      icon: ToggleLeft,
      keywords: ["admin", "feature", "toggle", "flag"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => {
        ctx.setMode("flag-toggle");
        ctx.setSearch("");
      },
    },
    {
      id: "admin-search-logs",
      label: "Search Audit Logs...",
      description: "Quick search through audit logs",
      icon: Search,
      keywords: ["admin", "audit", "logs", "search"],
      group: "admin",
      roles: ["admin", "super_admin"],
      action: (ctx) => {
        ctx.navigate("/admin/audit-logs");
        // Focus search on the page
        setTimeout(() => {
          window.dispatchEvent(new CustomEvent("command:focus-audit-search"));
        }, 100);
      },
    },
  ],
};

// =============================================================================
// Contextual Provider (Route-specific commands)
// =============================================================================

export const contextualProvider: CommandProvider = {
  id: "contextual",
  name: "Page Actions",
  getCommands: (ctx: CommandContext): Command[] => {
    const commands: Command[] = [];

    // Admin Users page commands
    if (ctx.pathname === "/admin/users" || ctx.pathname.startsWith("/admin/users")) {
      commands.push(
        {
          id: "ctx-create-user",
          label: "Create New User",
          icon: UserPlus,
          keywords: ["add", "new", "user"],
          group: "contextual",
          roles: ["admin", "super_admin"],
          routePatterns: ["/admin/users", "/admin/users/*"],
          action: () => {
            window.dispatchEvent(new CustomEvent("command:create-user"));
          },
        },
        {
          id: "ctx-export-users",
          label: "Export Users to CSV",
          icon: Download,
          keywords: ["export", "download", "csv"],
          group: "contextual",
          roles: ["admin", "super_admin"],
          routePatterns: ["/admin/users"],
          action: () => {
            window.dispatchEvent(new CustomEvent("command:export-users"));
          },
        }
      );
    }

    // Feature Flags page commands
    if (ctx.pathname === "/admin/feature-flags") {
      commands.push({
        id: "ctx-create-flag",
        label: "Create Feature Flag",
        icon: Flag,
        keywords: ["add", "new", "flag"],
        group: "contextual",
        roles: ["admin", "super_admin"],
        routePatterns: ["/admin/feature-flags"],
        action: () => {
          window.dispatchEvent(new CustomEvent("command:create-flag"));
        },
      });
    }

    // Billing page commands
    if (ctx.pathname === "/billing") {
      commands.push({
        id: "ctx-view-invoices",
        label: "View Invoices",
        icon: Receipt,
        keywords: ["invoices", "receipts", "payments"],
        group: "contextual",
        routePatterns: ["/billing"],
        action: () => {
          window.dispatchEvent(new CustomEvent("command:view-invoices"));
        },
      });
    }

    // Help command available everywhere
    commands.push({
      id: "ctx-help",
      label: "Get Help",
      icon: HelpCircle,
      keywords: ["help", "support", "docs", "documentation"],
      group: "contextual",
      action: () => {
        window.dispatchEvent(new CustomEvent("command:show-help"));
      },
    });

    return commands;
  },
};
