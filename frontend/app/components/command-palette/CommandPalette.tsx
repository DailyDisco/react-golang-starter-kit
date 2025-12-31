import { useCallback, useMemo, useState } from "react";

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
import { useCommandPalette } from "@/hooks/useCommandPalette";
import { formatShortcut, useKeyboardShortcuts } from "@/hooks/useKeyboardShortcuts";
import { useTheme } from "@/providers/theme-provider";
import { useNavigate } from "@tanstack/react-router";
import {
  Activity,
  Bell,
  CreditCard,
  FileText,
  Flag,
  History,
  Home,
  Key,
  LayoutDashboard,
  Link2,
  LogOut,
  Mail,
  Moon,
  Palette,
  Search,
  Settings,
  Shield,
  Sun,
  User,
  Users,
} from "lucide-react";
import { useTranslation } from "react-i18next";

interface CommandItemData {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  shortcut?: string;
  action: () => void;
  keywords?: string[];
  group: "navigation" | "actions" | "admin" | "settings";
}

export function CommandPalette() {
  const { t } = useTranslation();
  const { isOpen, close, toggle } = useCommandPalette();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { setTheme, resolvedTheme } = useTheme();
  const [search, setSearch] = useState("");

  const isAdmin = user?.role === "admin" || user?.role === "super_admin";

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

  const runCommand = useCallback(
    (command: () => void) => {
      close();
      command();
    },
    [close]
  );

  const commands = useMemo<CommandItemData[]>(() => {
    const items: CommandItemData[] = [
      // Navigation
      {
        id: "dashboard",
        label: "Go to Dashboard",
        icon: Home,
        action: () => navigate({ to: "/dashboard" }),
        keywords: ["home", "main"],
        group: "navigation",
      },
      {
        id: "billing",
        label: "Go to Billing",
        icon: CreditCard,
        action: () => navigate({ to: "/billing" }),
        keywords: ["payment", "subscription", "plan"],
        group: "navigation",
      },
      {
        id: "profile",
        label: "Go to Profile",
        icon: User,
        action: () => navigate({ to: "/settings/profile" }),
        keywords: ["account", "me"],
        group: "navigation",
      },
      {
        id: "security",
        label: "Go to Security Settings",
        icon: Shield,
        action: () => navigate({ to: "/settings/security" }),
        keywords: ["password", "2fa", "authentication"],
        group: "navigation",
      },
      {
        id: "preferences",
        label: "Go to Preferences",
        icon: Palette,
        action: () => navigate({ to: "/settings/preferences" }),
        keywords: ["theme", "language"],
        group: "navigation",
      },
      {
        id: "notifications",
        label: "Go to Notifications",
        icon: Bell,
        action: () => navigate({ to: "/settings/notifications" }),
        keywords: ["alerts", "email"],
        group: "navigation",
      },
      {
        id: "privacy",
        label: "Go to Privacy Settings",
        icon: Key,
        action: () => navigate({ to: "/settings/privacy" }),
        keywords: ["data"],
        group: "navigation",
      },
      {
        id: "login-history",
        label: "Go to Login History",
        icon: History,
        action: () => navigate({ to: "/settings/login-history" }),
        keywords: ["sessions", "activity"],
        group: "navigation",
      },
      {
        id: "connected-accounts",
        label: "Go to Connected Accounts",
        icon: Link2,
        action: () => navigate({ to: "/settings/connected-accounts" }),
        keywords: ["oauth", "social", "google", "github"],
        group: "navigation",
      },

      // Actions
      {
        id: "toggle-theme",
        label: resolvedTheme === "dark" ? "Switch to Light Mode" : "Switch to Dark Mode",
        icon: resolvedTheme === "dark" ? Sun : Moon,
        action: () => setTheme(resolvedTheme === "dark" ? "light" : "dark"),
        keywords: ["theme", "dark", "light", "mode"],
        group: "actions",
      },
      {
        id: "sign-out",
        label: "Sign Out",
        icon: LogOut,
        action: () => logout(),
        keywords: ["logout", "exit"],
        group: "actions",
      },

      // Settings shortcuts
      {
        id: "settings",
        label: "Open Settings",
        icon: Settings,
        shortcut: formatShortcut({ key: ",", meta: true }),
        action: () => navigate({ to: "/settings" }),
        keywords: ["preferences", "options"],
        group: "settings",
      },
    ];

    // Admin commands
    if (isAdmin) {
      items.push(
        {
          id: "admin-overview",
          label: "Admin Overview",
          icon: LayoutDashboard,
          action: () => navigate({ to: "/admin" }),
          keywords: ["admin", "dashboard"],
          group: "admin",
        },
        {
          id: "admin-users",
          label: "Manage Users",
          icon: Users,
          action: () => navigate({ to: "/admin/users" }),
          keywords: ["admin", "users", "accounts"],
          group: "admin",
        },
        {
          id: "admin-audit",
          label: "View Audit Logs",
          icon: FileText,
          action: () => navigate({ to: "/admin/audit-logs" }),
          keywords: ["admin", "logs", "history"],
          group: "admin",
        },
        {
          id: "admin-flags",
          label: "Feature Flags",
          icon: Flag,
          action: () => navigate({ to: "/admin/feature-flags" }),
          keywords: ["admin", "features", "toggles"],
          group: "admin",
        },
        {
          id: "admin-health",
          label: "System Health",
          icon: Activity,
          action: () => navigate({ to: "/admin/health" }),
          keywords: ["admin", "status", "monitoring"],
          group: "admin",
        },
        {
          id: "admin-announcements",
          label: "Announcements",
          icon: Bell,
          action: () => navigate({ to: "/admin/announcements" }),
          keywords: ["admin", "banners", "messages"],
          group: "admin",
        },
        {
          id: "admin-email",
          label: "Email Templates",
          icon: Mail,
          action: () => navigate({ to: "/admin/email-templates" }),
          keywords: ["admin", "templates", "email"],
          group: "admin",
        },
        {
          id: "admin-settings",
          label: "Admin Settings",
          icon: Settings,
          action: () => navigate({ to: "/admin/settings" }),
          keywords: ["admin", "configuration"],
          group: "admin",
        }
      );
    }

    return items;
  }, [isAdmin, navigate, logout, setTheme, resolvedTheme]);

  const navigationCommands = commands.filter((c) => c.group === "navigation");
  const actionCommands = commands.filter((c) => c.group === "actions");
  const settingsCommands = commands.filter((c) => c.group === "settings");
  const adminCommands = commands.filter((c) => c.group === "admin");

  return (
    <CommandDialog
      open={isOpen}
      onOpenChange={(open) => (open ? undefined : close())}
      title="Command Palette"
      description="Search for commands, pages, and actions"
    >
      <CommandInput
        placeholder="Type a command or search..."
        value={search}
        onValueChange={setSearch}
      />
      <CommandList>
        <CommandEmpty>
          <div className="flex flex-col items-center gap-2 py-4">
            <Search className="text-muted-foreground h-10 w-10" />
            <p>No results found.</p>
            <p className="text-muted-foreground text-sm">Try searching for something else.</p>
          </div>
        </CommandEmpty>

        <CommandGroup heading="Navigation">
          {navigationCommands.map((command) => {
            const Icon = command.icon;
            return (
              <CommandItem
                key={command.id}
                value={`${command.label} ${command.keywords?.join(" ") || ""}`}
                onSelect={() => runCommand(command.action)}
              >
                <Icon className="mr-2 h-4 w-4" />
                <span>{command.label}</span>
                {command.shortcut && <CommandShortcut>{command.shortcut}</CommandShortcut>}
              </CommandItem>
            );
          })}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Actions">
          {actionCommands.map((command) => {
            const Icon = command.icon;
            return (
              <CommandItem
                key={command.id}
                value={`${command.label} ${command.keywords?.join(" ") || ""}`}
                onSelect={() => runCommand(command.action)}
              >
                <Icon className="mr-2 h-4 w-4" />
                <span>{command.label}</span>
                {command.shortcut && <CommandShortcut>{command.shortcut}</CommandShortcut>}
              </CommandItem>
            );
          })}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Settings">
          {settingsCommands.map((command) => {
            const Icon = command.icon;
            return (
              <CommandItem
                key={command.id}
                value={`${command.label} ${command.keywords?.join(" ") || ""}`}
                onSelect={() => runCommand(command.action)}
              >
                <Icon className="mr-2 h-4 w-4" />
                <span>{command.label}</span>
                {command.shortcut && <CommandShortcut>{command.shortcut}</CommandShortcut>}
              </CommandItem>
            );
          })}
        </CommandGroup>

        {isAdmin && (
          <>
            <CommandSeparator />
            <CommandGroup heading="Administration">
              {adminCommands.map((command) => {
                const Icon = command.icon;
                return (
                  <CommandItem
                    key={command.id}
                    value={`${command.label} ${command.keywords?.join(" ") || ""}`}
                    onSelect={() => runCommand(command.action)}
                  >
                    <Icon className="mr-2 h-4 w-4" />
                    <span>{command.label}</span>
                    {command.shortcut && <CommandShortcut>{command.shortcut}</CommandShortcut>}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </>
        )}
      </CommandList>
    </CommandDialog>
  );
}
