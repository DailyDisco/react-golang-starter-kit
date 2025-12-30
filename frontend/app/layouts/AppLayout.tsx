import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/auth/useAuth";
import { Link, Outlet, useLocation } from "@tanstack/react-router";
import {
  Activity,
  Bell,
  CreditCard,
  ExternalLink,
  FileText,
  Flag,
  Globe,
  History,
  Home,
  Key,
  LayoutDashboard,
  Link2,
  LogOut,
  Mail,
  Palette,
  Settings,
  Shield,
  Sparkles,
  User,
  Users,
} from "lucide-react";

import { Breadcrumbs } from "../components/navigation/Breadcrumbs";

interface NavItem {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NavGroup {
  label: string;
  items: NavItem[];
  adminOnly?: boolean;
}

const navigationGroups: NavGroup[] = [
  {
    label: "Main",
    items: [
      { name: "Dashboard", href: "/dashboard", icon: Home },
      { name: "Billing", href: "/billing", icon: CreditCard },
    ],
  },
  {
    label: "Account",
    items: [
      { name: "Profile", href: "/settings/profile", icon: User },
      { name: "Security", href: "/settings/security", icon: Shield },
      { name: "Preferences", href: "/settings/preferences", icon: Palette },
      { name: "Notifications", href: "/settings/notifications", icon: Bell },
      { name: "Privacy", href: "/settings/privacy", icon: Key },
      { name: "Login History", href: "/settings/login-history", icon: History },
      { name: "Connected Accounts", href: "/settings/connected-accounts", icon: Link2 },
    ],
  },
  {
    label: "Administration",
    adminOnly: true,
    items: [
      { name: "Overview", href: "/admin", icon: LayoutDashboard },
      { name: "Users", href: "/admin/users", icon: Users },
      { name: "Audit Logs", href: "/admin/audit-logs", icon: FileText },
      { name: "Feature Flags", href: "/admin/feature-flags", icon: Flag },
      { name: "System Health", href: "/admin/health", icon: Activity },
      { name: "Announcements", href: "/admin/announcements", icon: Bell },
      { name: "Email Templates", href: "/admin/email-templates", icon: Mail },
      { name: "Admin Settings", href: "/admin/settings", icon: Settings },
    ],
  },
];

export function AppLayout() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const isAdmin = user?.role === "admin" || user?.role === "super_admin";

  const isActive = (href: string) => {
    // Exact match for specific routes
    if (href === "/dashboard" && location.pathname === "/dashboard") return true;
    if (href === "/admin" && location.pathname === "/admin") return true;
    // Prefix match for nested routes
    if (href !== "/dashboard" && href !== "/admin" && location.pathname.startsWith(href)) return true;
    return false;
  };

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader>
          <Link
            to="/"
            search={{}}
            className="flex items-center gap-3 px-2 py-4 transition-opacity hover:opacity-80"
          >
            <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-lg">
              <span className="text-primary-foreground text-lg font-bold">RG</span>
            </div>
            <div className="flex flex-col">
              <span className="text-sm font-semibold">React + Go</span>
              <span className="text-muted-foreground text-xs">Starter Kit</span>
            </div>
          </Link>
        </SidebarHeader>

        <SidebarSeparator />

        <SidebarContent>
          {navigationGroups.map((group) => {
            // Skip admin group for non-admin users
            if (group.adminOnly && !isAdmin) return null;

            return (
              <SidebarGroup key={group.label}>
                <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
                <SidebarGroupContent>
                  <SidebarMenu>
                    {group.items.map((item) => {
                      const Icon = item.icon;
                      return (
                        <SidebarMenuItem key={item.href}>
                          <SidebarMenuButton
                            asChild
                            isActive={isActive(item.href)}
                            tooltip={item.name}
                          >
                            <Link
                              to={item.href}
                              search={{}}
                            >
                              <Icon className="h-4 w-4" />
                              <span>{item.name}</span>
                            </Link>
                          </SidebarMenuButton>
                        </SidebarMenuItem>
                      );
                    })}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            );
          })}

          {/* Browse Site - Links to public pages */}
          <SidebarGroup>
            <SidebarGroupLabel>Browse Site</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    asChild
                    tooltip="Home"
                  >
                    <Link
                      to="/"
                      search={{}}
                    >
                      <Globe className="h-4 w-4" />
                      <span>Home</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    asChild
                    tooltip="Demo"
                  >
                    <Link
                      to="/demo"
                      search={{}}
                    >
                      <Sparkles className="h-4 w-4" />
                      <span>Demo</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    asChild
                    tooltip="Pricing"
                  >
                    <Link
                      to="/pricing"
                      search={{}}
                    >
                      <ExternalLink className="h-4 w-4" />
                      <span>Pricing</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter>
          <div className="bg-sidebar-accent/50 flex items-center gap-3 rounded-lg p-3">
            <Avatar className="h-9 w-9">
              <AvatarImage
                src=""
                alt={user?.name || "User"}
              />
              <AvatarFallback className="bg-primary text-primary-foreground text-xs">
                {user?.name
                  ?.split(" ")
                  .map((n) => n[0])
                  .join("")
                  .toUpperCase() || "U"}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{user?.name || "User"}</p>
              <p className="text-muted-foreground truncate text-xs">{user?.email || ""}</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={logout}
            className="w-full justify-start gap-2"
          >
            <LogOut className="h-4 w-4" />
            Sign Out
          </Button>
        </SidebarFooter>
      </Sidebar>

      <SidebarInset>
        <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-40 flex h-14 items-center gap-4 border-b px-4 backdrop-blur">
          <SidebarTrigger className="-ml-1" />
          <Breadcrumbs />
        </header>
        <main className="flex-1 p-6">
          <div className="mx-auto max-w-7xl">
            <Outlet />
          </div>
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
