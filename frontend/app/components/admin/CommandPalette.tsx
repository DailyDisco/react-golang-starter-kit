import { useEffect, useState } from "react";

import { useNavigate } from "@tanstack/react-router";
import { Activity, BarChart3, Bell, FileText, Flag, Mail, Settings, Users } from "lucide-react";
import { useTranslation } from "react-i18next";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "../ui/command";

const adminRoutes = [
  {
    labelKey: "commandPalette.routes.dashboard",
    href: "/admin",
    icon: BarChart3,
    keywords: ["home", "overview", "stats"],
  },
  {
    labelKey: "commandPalette.routes.users",
    href: "/admin/users",
    icon: Users,
    keywords: ["members", "accounts", "people"],
  },
  {
    labelKey: "commandPalette.routes.auditLogs",
    href: "/admin/audit-logs",
    icon: FileText,
    keywords: ["history", "activity", "events"],
  },
  {
    labelKey: "commandPalette.routes.featureFlags",
    href: "/admin/feature-flags",
    icon: Flag,
    keywords: ["toggles", "features", "flags"],
  },
  {
    labelKey: "commandPalette.routes.systemHealth",
    href: "/admin/health",
    icon: Activity,
    keywords: ["status", "monitoring", "metrics"],
  },
  {
    labelKey: "commandPalette.routes.announcements",
    href: "/admin/announcements",
    icon: Bell,
    keywords: ["banners", "notifications", "alerts"],
  },
  {
    labelKey: "commandPalette.routes.emailTemplates",
    href: "/admin/email-templates",
    icon: Mail,
    keywords: ["emails", "templates", "messages"],
  },
  {
    labelKey: "commandPalette.routes.settings",
    href: "/admin/settings",
    icon: Settings,
    keywords: ["config", "configuration", "preferences"],
  },
];

const quickActions = [
  {
    labelKey: "commandPalette.actions.createUser",
    action: "create-user",
    icon: Users,
  },
  {
    labelKey: "commandPalette.actions.createFlag",
    action: "create-flag",
    icon: Flag,
  },
  {
    labelKey: "commandPalette.actions.createAnnouncement",
    action: "create-announcement",
    icon: Bell,
  },
];

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const { t } = useTranslation("admin");
  const navigate = useNavigate();

  const handleSelect = (href: string) => {
    onOpenChange(false);
    void navigate({ to: href });
  };

  const handleAction = (action: string) => {
    onOpenChange(false);
    switch (action) {
      case "create-user":
        void navigate({ to: "/admin/users" });
        break;
      case "create-flag":
        void navigate({ to: "/admin/feature-flags" });
        break;
      case "create-announcement":
        void navigate({ to: "/admin/announcements" });
        break;
    }
  };

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
    >
      <CommandInput placeholder={t("commandPalette.search")} />
      <CommandList>
        <CommandEmpty>{t("commandPalette.noResults")}</CommandEmpty>
        <CommandGroup heading={t("commandPalette.navigation")}>
          {adminRoutes.map((route) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const label = t(route.labelKey as any) as string;
            return (
              <CommandItem
                key={route.href}
                value={label + " " + route.keywords.join(" ")}
                onSelect={() => handleSelect(route.href)}
              >
                <route.icon className="mr-2 h-4 w-4" />
                <span>{label}</span>
              </CommandItem>
            );
          })}
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading={t("commandPalette.quickActions")}>
          {quickActions.map((action) => {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const label = t(action.labelKey as any) as string;
            return (
              <CommandItem
                key={action.action}
                value={label}
                onSelect={() => handleAction(action.action)}
              >
                <action.icon className="mr-2 h-4 w-4" />
                <span>{label}</span>
              </CommandItem>
            );
          })}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}

// Hook to use command palette with keyboard shortcut
export function useCommandPalette() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  return { open, setOpen };
}
