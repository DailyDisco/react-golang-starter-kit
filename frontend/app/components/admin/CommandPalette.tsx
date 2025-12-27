import { useEffect, useState } from "react";

import { useNavigate } from "@tanstack/react-router";
import { Activity, BarChart3, Bell, FileText, Flag, Mail, Search, Settings, Shield, Users } from "lucide-react";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "../ui/command";

const adminRoutes = [
  {
    label: "Dashboard",
    href: "/admin",
    icon: BarChart3,
    keywords: ["home", "overview", "stats"],
  },
  {
    label: "Users",
    href: "/admin/users",
    icon: Users,
    keywords: ["members", "accounts", "people"],
  },
  {
    label: "Audit Logs",
    href: "/admin/audit-logs",
    icon: FileText,
    keywords: ["history", "activity", "events"],
  },
  {
    label: "Feature Flags",
    href: "/admin/feature-flags",
    icon: Flag,
    keywords: ["toggles", "features", "flags"],
  },
  {
    label: "System Health",
    href: "/admin/health",
    icon: Activity,
    keywords: ["status", "monitoring", "metrics"],
  },
  {
    label: "Announcements",
    href: "/admin/announcements",
    icon: Bell,
    keywords: ["banners", "notifications", "alerts"],
  },
  {
    label: "Email Templates",
    href: "/admin/email-templates",
    icon: Mail,
    keywords: ["emails", "templates", "messages"],
  },
  {
    label: "Settings",
    href: "/admin/settings",
    icon: Settings,
    keywords: ["config", "configuration", "preferences"],
  },
];

const quickActions = [
  {
    label: "Create new user",
    action: "create-user",
    icon: Users,
  },
  {
    label: "Create feature flag",
    action: "create-flag",
    icon: Flag,
  },
  {
    label: "Create announcement",
    action: "create-announcement",
    icon: Bell,
  },
];

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const navigate = useNavigate();

  const handleSelect = (href: string) => {
    onOpenChange(false);
    navigate({ to: href });
  };

  const handleAction = (action: string) => {
    onOpenChange(false);
    switch (action) {
      case "create-user":
        navigate({ to: "/admin/users" });
        break;
      case "create-flag":
        navigate({ to: "/admin/feature-flags" });
        break;
      case "create-announcement":
        navigate({ to: "/admin/announcements" });
        break;
    }
  };

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
    >
      <CommandInput placeholder="Search admin panel..." />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        <CommandGroup heading="Navigation">
          {adminRoutes.map((route) => (
            <CommandItem
              key={route.href}
              value={route.label + " " + route.keywords.join(" ")}
              onSelect={() => handleSelect(route.href)}
            >
              <route.icon className="mr-2 h-4 w-4" />
              <span>{route.label}</span>
            </CommandItem>
          ))}
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Quick Actions">
          {quickActions.map((action) => (
            <CommandItem
              key={action.action}
              value={action.label}
              onSelect={() => handleAction(action.action)}
            >
              <action.icon className="mr-2 h-4 w-4" />
              <span>{action.label}</span>
            </CommandItem>
          ))}
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
