import { useState } from "react";

import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "@tanstack/react-router";
import {
  Bell,
  ChevronLeft,
  ChevronRight,
  FileText,
  Flag,
  Heart,
  LayoutDashboard,
  Mail,
  Menu,
  Settings,
  Shield,
  Users,
} from "lucide-react";
import { useTranslation } from "react-i18next";

interface NavItem {
  href: string;
  labelKey: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NavGroup {
  titleKey: string;
  items: NavItem[];
}

const adminNavGroups: NavGroup[] = [
  {
    titleKey: "sidebar.groups.overview",
    items: [{ href: "/admin", labelKey: "sidebar.dashboard", icon: LayoutDashboard }],
  },
  {
    titleKey: "sidebar.groups.management",
    items: [
      { href: "/admin/users", labelKey: "sidebar.users", icon: Users },
      { href: "/admin/feature-flags", labelKey: "sidebar.featureFlags", icon: Flag },
      { href: "/admin/announcements", labelKey: "sidebar.announcements", icon: Bell },
      { href: "/admin/email-templates", labelKey: "sidebar.emailTemplates", icon: Mail },
    ],
  },
  {
    titleKey: "sidebar.groups.system",
    items: [
      { href: "/admin/audit-logs", labelKey: "sidebar.auditLogs", icon: FileText },
      { href: "/admin/health", labelKey: "sidebar.health", icon: Heart },
      { href: "/admin/settings", labelKey: "sidebar.settings", icon: Settings },
    ],
  },
];

interface AdminSidebarProps {
  collapsed?: boolean;
  onCollapsedChange?: (collapsed: boolean) => void;
}

export function AdminSidebar({ collapsed = false, onCollapsedChange }: AdminSidebarProps) {
  const { t } = useTranslation("admin");
  const location = useLocation();

  const isActive = (href: string) => {
    if (href === "/admin") {
      return location.pathname === "/admin" || location.pathname === "/admin/";
    }
    return location.pathname.startsWith(href);
  };

  const handleCollapse = () => {
    onCollapsedChange?.(!collapsed);
  };

  return (
    <aside
      className={cn(
        "bg-card flex h-full flex-col border-r transition-all duration-300 ease-in-out",
        collapsed ? "w-16" : "w-64"
      )}
    >
      {/* Header */}
      <div className="flex h-14 items-center justify-between border-b px-4">
        {!collapsed && (
          <div className="flex items-center gap-2">
            <Shield className="text-primary h-5 w-5" />
            <span className="font-semibold">{t("sidebar.title", "Admin")}</span>
          </div>
        )}
        {collapsed && <Shield className="text-primary mx-auto h-5 w-5" />}
        {onCollapsedChange && (
          <Button
            variant="ghost"
            size="icon"
            className={cn("h-8 w-8", collapsed && "mx-auto")}
            onClick={handleCollapse}
          >
            {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </Button>
        )}
      </div>

      {/* Navigation */}
      <ScrollArea className="flex-1 py-2">
        <nav className="space-y-2 px-2">
          {adminNavGroups.map((group, groupIndex) => (
            <div key={group.titleKey}>
              {groupIndex > 0 && <Separator className="my-2" />}
              {!collapsed && (
                <p className="text-muted-foreground mb-2 px-3 text-xs font-medium tracking-wider uppercase">
                  {t(group.titleKey as never)}
                </p>
              )}
              <div className="space-y-1">
                {group.items.map((item) => {
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
                              "mx-auto flex h-10 w-10 items-center justify-center rounded-md transition-colors",
                              active
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                            )}
                          >
                            <Icon className="h-5 w-5" />
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
                        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                        active
                          ? "bg-primary text-primary-foreground"
                          : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                      )}
                    >
                      <Icon className="h-5 w-5" />
                      {t(item.labelKey as never)}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>
      </ScrollArea>

      {/* Footer - Back to App */}
      <div className="border-t p-2">
        {collapsed ? (
          <Tooltip delayDuration={0}>
            <TooltipTrigger asChild>
              <Link
                to="/dashboard"
                className="text-muted-foreground hover:bg-accent hover:text-accent-foreground mx-auto flex h-10 w-10 items-center justify-center rounded-md transition-colors"
              >
                <ChevronLeft className="h-5 w-5" />
              </Link>
            </TooltipTrigger>
            <TooltipContent side="right">{t("sidebar.backToApp", "Back to App")}</TooltipContent>
          </Tooltip>
        ) : (
          <Link
            to="/dashboard"
            className="text-muted-foreground hover:bg-accent hover:text-accent-foreground flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors"
          >
            <ChevronLeft className="h-5 w-5" />
            {t("sidebar.backToApp", "Back to App")}
          </Link>
        )}
      </div>
    </aside>
  );
}

// Mobile sidebar using Sheet
interface MobileAdminSidebarProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function MobileAdminSidebar({ open, onOpenChange }: MobileAdminSidebarProps) {
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
          <span className="sr-only">Toggle admin menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent
        side="left"
        className="w-64 p-0"
      >
        <AdminSidebar />
      </SheetContent>
    </Sheet>
  );
}
