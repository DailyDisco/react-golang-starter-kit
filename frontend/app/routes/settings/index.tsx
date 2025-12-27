import { useState } from "react";

import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, Outlet, useLocation } from "@tanstack/react-router";
import { Bell, ChevronRight, Globe, History, Key, Link2, Palette, Search, Shield, User } from "lucide-react";

import { Card, CardContent } from "../../components/ui/card";
import { Input } from "../../components/ui/input";
import { requireAuth } from "../../lib/guards";
import { cn } from "../../lib/utils";
import { AuthService } from "../../services/auth/authService";

export const Route = createFileRoute("/settings/")({
  beforeLoad: () => requireAuth(),
  component: SettingsLayout,
});

const settingsNavItems = [
  {
    title: "Profile",
    href: "/settings/profile",
    icon: User,
    description: "Manage your personal information",
    keywords: ["name", "email", "avatar", "photo"],
  },
  {
    title: "Security",
    href: "/settings/security",
    icon: Shield,
    description: "Password, 2FA, and active sessions",
    keywords: ["password", "two-factor", "2fa", "sessions", "authentication"],
  },
  {
    title: "Login History",
    href: "/settings/login-history",
    icon: History,
    description: "View your recent login activity",
    keywords: ["login", "history", "activity", "devices"],
  },
  {
    title: "Preferences",
    href: "/settings/preferences",
    icon: Palette,
    description: "Theme, language, and display settings",
    keywords: ["theme", "dark", "light", "language", "timezone", "date", "time"],
  },
  {
    title: "Notifications",
    href: "/settings/notifications",
    icon: Bell,
    description: "Email notification preferences",
    keywords: ["email", "notifications", "alerts", "marketing", "updates"],
  },
  {
    title: "Privacy",
    href: "/settings/privacy",
    icon: Key,
    description: "Data export and account deletion",
    keywords: ["privacy", "data", "export", "delete", "account"],
  },
  {
    title: "Connected Accounts",
    href: "/settings/connected-accounts",
    icon: Link2,
    description: "Manage linked OAuth providers",
    keywords: ["google", "github", "oauth", "connected", "linked"],
  },
];

function SettingsLayout() {
  const location = useLocation();
  const [searchQuery, setSearchQuery] = useState("");

  const { data: user } = useQuery({
    queryKey: ["currentUser"],
    queryFn: () => AuthService.getCurrentUser(),
    staleTime: 60 * 1000,
  });

  // Filter nav items based on search
  const filteredNavItems = settingsNavItems.filter((item) => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      item.title.toLowerCase().includes(query) ||
      item.description.toLowerCase().includes(query) ||
      item.keywords.some((k) => k.includes(query))
    );
  });

  // Check if we're on the main settings page (not a subpage)
  const isMainPage = location.pathname === "/settings" || location.pathname === "/settings/";

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="mt-1 text-gray-500">Manage your account settings and preferences</p>
      </div>

      <div className="flex flex-col gap-8 lg:flex-row">
        {/* Sidebar */}
        <aside className="w-full shrink-0 lg:w-64">
          {/* Search */}
          <div className="relative mb-4">
            <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <Input
              type="text"
              placeholder="Search settings..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>

          {/* Navigation */}
          <nav className="space-y-1">
            {filteredNavItems.map((item) => {
              const isActive = location.pathname.startsWith(item.href);
              const Icon = item.icon;

              return (
                <Link
                  key={item.href}
                  to={item.href}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                    isActive ? "bg-blue-50 text-blue-700" : "text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                  )}
                >
                  <Icon className={cn("h-5 w-5", isActive ? "text-blue-600" : "text-gray-400")} />
                  <span className="flex-1">{item.title}</span>
                  <ChevronRight
                    className={cn("h-4 w-4 transition-transform", isActive ? "text-blue-600" : "text-gray-300")}
                  />
                </Link>
              );
            })}
          </nav>

          {/* User Info Card */}
          {user && (
            <Card className="mt-6">
              <CardContent className="p-4">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 font-medium text-white">
                    {user.name?.charAt(0).toUpperCase() || "U"}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium text-gray-900">{user.name}</p>
                    <p className="truncate text-xs text-gray-500">{user.email}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </aside>

        {/* Main Content */}
        <main className="flex-1">
          {isMainPage ? (
            // Show overview cards when on main settings page
            <div className="grid gap-4 sm:grid-cols-2">
              {filteredNavItems.map((item) => {
                const Icon = item.icon;
                return (
                  <Link
                    key={item.href}
                    to={item.href}
                  >
                    <Card className="h-full transition-all hover:border-blue-200 hover:shadow-md">
                      <CardContent className="p-6">
                        <div className="flex items-start gap-4">
                          <div className="rounded-lg bg-blue-50 p-2.5">
                            <Icon className="h-5 w-5 text-blue-600" />
                          </div>
                          <div className="flex-1">
                            <h3 className="font-semibold text-gray-900">{item.title}</h3>
                            <p className="mt-1 text-sm text-gray-500">{item.description}</p>
                          </div>
                          <ChevronRight className="h-5 w-5 text-gray-300" />
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                );
              })}
            </div>
          ) : (
            <Outlet />
          )}
        </main>
      </div>
    </div>
  );
}
