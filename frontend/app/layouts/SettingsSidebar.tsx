import { useState } from "react";

import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "@tanstack/react-router";
import {
  Bell,
  ChevronLeft,
  ChevronRight,
  History,
  Key,
  KeyRound,
  Link2,
  Menu,
  Palette,
  Settings,
  Shield,
  User,
} from "lucide-react";
import { useTranslation } from "react-i18next";

interface NavItem {
  href: string;
  labelKey: string;
  icon: React.ComponentType<{ className?: string }>;
}

const settingsNavItems: NavItem[] = [
  { href: "/settings/profile", labelKey: "nav.profile.title", icon: User },
  { href: "/settings/security", labelKey: "nav.security.title", icon: Shield },
  { href: "/settings/login-history", labelKey: "nav.loginHistory.title", icon: History },
  { href: "/settings/preferences", labelKey: "nav.preferences.title", icon: Palette },
  { href: "/settings/notifications", labelKey: "nav.notifications.title", icon: Bell },
  { href: "/settings/privacy", labelKey: "nav.privacy.title", icon: Key },
  { href: "/settings/connected-accounts", labelKey: "nav.connectedAccounts.title", icon: Link2 },
  { href: "/settings/api-keys", labelKey: "nav.apiKeys.title", icon: KeyRound },
];

interface SettingsSidebarProps {
  collapsed?: boolean;
  onCollapsedChange?: (collapsed: boolean) => void;
}

export function SettingsSidebar({ collapsed = false, onCollapsedChange }: SettingsSidebarProps) {
  const { t } = useTranslation("settings");
  const location = useLocation();

  const isActive = (href: string) => {
    return location.pathname === href || location.pathname === `${href}/`;
  };

  const handleCollapse = () => {
    onCollapsedChange?.(!collapsed);
  };

  return (
    <aside
      className={cn(
        "bg-card flex h-full flex-col border-r transition-all duration-300 ease-in-out",
        collapsed ? "w-14" : "w-56"
      )}
    >
      {/* Header */}
      <div className="flex h-12 items-center justify-between border-b px-3">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <Settings className="text-primary h-4 w-4" />
            <span className="text-sm font-semibold">{t("title", "Settings")}</span>
          </div>
        )}
        {collapsed && <Settings className="text-primary mx-auto h-4 w-4" />}
        {onCollapsedChange && (
          <Button
            variant="ghost"
            size="icon"
            className={cn("h-7 w-7", collapsed && "mx-auto")}
            onClick={handleCollapse}
          >
            {collapsed ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronLeft className="h-3.5 w-3.5" />}
          </Button>
        )}
      </div>

      {/* Navigation */}
      <ScrollArea className="flex-1 py-2">
        <nav className="space-y-1 px-2">
          {settingsNavItems.map((item) => {
            const Icon = item.icon;
            const active = isActive(item.href);

            if (collapsed) {
              return (
                <Tooltip
                  key={item.href}
                  delayDuration={0}
                >
                  <TooltipTrigger asChild>
                    <Link
                      to={item.href}
                      className={cn(
                        "mx-auto flex h-9 w-9 items-center justify-center rounded-md transition-colors",
                        active
                          ? "bg-primary text-primary-foreground"
                          : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                      )}
                    >
                      <Icon className="h-4 w-4" />
                    </Link>
                  </TooltipTrigger>
                  <TooltipContent side="right">{t(item.labelKey as never)}</TooltipContent>
                </Tooltip>
              );
            }

            return (
              <Link
                key={item.href}
                to={item.href}
                className={cn(
                  "flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors",
                  active
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                )}
              >
                <Icon className="h-4 w-4" />
                {t(item.labelKey as never)}
              </Link>
            );
          })}
        </nav>
      </ScrollArea>

      {/* Footer - Back to Dashboard */}
      <div className="border-t p-2">
        {collapsed ? (
          <Tooltip delayDuration={0}>
            <TooltipTrigger asChild>
              <Link
                to="/dashboard"
                className="text-muted-foreground hover:bg-accent hover:text-accent-foreground mx-auto flex h-9 w-9 items-center justify-center rounded-md transition-colors"
              >
                <ChevronLeft className="h-4 w-4" />
              </Link>
            </TooltipTrigger>
            <TooltipContent side="right">{t("backToDashboard", "Back to Dashboard")}</TooltipContent>
          </Tooltip>
        ) : (
          <Link
            to="/dashboard"
            className="text-muted-foreground hover:bg-accent hover:text-accent-foreground flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors"
          >
            <ChevronLeft className="h-4 w-4" />
            {t("backToDashboard", "Back to Dashboard")}
          </Link>
        )}
      </div>
    </aside>
  );
}

// Mobile sidebar using Sheet
interface MobileSettingsSidebarProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function MobileSettingsSidebar({ open, onOpenChange }: MobileSettingsSidebarProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = open ?? internalOpen;
  const setIsOpen = onOpenChange ?? setInternalOpen;

  return (
    <Sheet
      open={isOpen}
      onOpenChange={setIsOpen}
    >
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="md:hidden"
        >
          <Menu className="h-5 w-5" />
          <span className="sr-only">Toggle settings menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent
        side="left"
        className="w-56 p-0"
      >
        <SettingsSidebar />
      </SheetContent>
    </Sheet>
  );
}
