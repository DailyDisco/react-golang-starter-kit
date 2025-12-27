import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, Outlet, useLocation } from "@tanstack/react-router";
import {
  Activity,
  ArrowDownRight,
  ArrowUpRight,
  BarChart3,
  Bell,
  Download,
  FileText,
  Flag,
  Mail,
  Menu,
  Search,
  Settings,
  Shield,
  TrendingUp,
  Users,
} from "lucide-react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

import { CommandPalette, useCommandPalette } from "../../components/admin";
import { Avatar, AvatarFallback } from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "../../components/ui/chart";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "../../components/ui/sheet";
import { requireAdmin } from "../../lib/guards";
import { AdminService, type AdminStats } from "../../services/admin";

export const Route = createFileRoute("/admin/")({
  beforeLoad: () => requireAdmin(),
  component: AdminDashboard,
});

function AdminDashboard() {
  const location = useLocation();
  const isIndex = location.pathname === "/admin" || location.pathname === "/admin/";
  const { open: commandOpen, setOpen: setCommandOpen } = useCommandPalette();

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Command Palette */}
      <CommandPalette
        open={commandOpen}
        onOpenChange={setCommandOpen}
      />

      <div className="flex">
        {/* Desktop Sidebar */}
        <nav className="hidden min-h-screen w-64 bg-white p-4 shadow-sm lg:block dark:bg-gray-800">
          <SidebarContent onCommandOpen={() => setCommandOpen(true)} />
        </nav>

        {/* Main content */}
        <main className="flex-1">
          {/* Mobile Header */}
          <div className="sticky top-0 z-10 flex items-center justify-between border-b bg-white p-4 lg:hidden dark:bg-gray-800">
            <div className="flex items-center gap-2">
              <Sheet>
                <SheetTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                  >
                    <Menu className="h-5 w-5" />
                  </Button>
                </SheetTrigger>
                <SheetContent
                  side="left"
                  className="w-64 p-0"
                >
                  <SheetHeader className="border-b p-4">
                    <SheetTitle className="flex items-center gap-2">
                      <Shield className="h-5 w-5" />
                      Admin Panel
                    </SheetTitle>
                  </SheetHeader>
                  <div className="p-4">
                    <SidebarContent onCommandOpen={() => setCommandOpen(true)} />
                  </div>
                </SheetContent>
              </Sheet>
              <h1 className="flex items-center gap-2 text-lg font-bold">
                <Shield className="h-5 w-5" />
                Admin
              </h1>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCommandOpen(true)}
              className="gap-2"
            >
              <Search className="h-4 w-4" />
              <span className="hidden sm:inline">Search</span>
              <kbd className="bg-muted pointer-events-none hidden h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium opacity-100 select-none sm:flex">
                <span className="text-xs">⌘</span>K
              </kbd>
            </Button>
          </div>

          <div className="mx-auto max-w-7xl p-4 lg:p-8">{isIndex ? <DashboardStats /> : <Outlet />}</div>
        </main>
      </div>
    </div>
  );
}

function SidebarContent({ onCommandOpen }: { onCommandOpen: () => void }) {
  return (
    <>
      <div className="mb-6 hidden lg:block">
        <h1 className="flex items-center gap-2 text-xl font-bold text-gray-900 dark:text-white">
          <Shield className="h-6 w-6" />
          Admin Panel
        </h1>
      </div>

      {/* Search Button */}
      <button
        onClick={onCommandOpen}
        className="mb-4 flex w-full items-center gap-2 rounded-lg border bg-gray-50 px-3 py-2 text-sm text-gray-500 transition-colors hover:bg-gray-100 dark:border-gray-700 dark:bg-gray-800 dark:hover:bg-gray-700"
      >
        <Search className="h-4 w-4" />
        <span className="flex-1 text-left">Search...</span>
        <kbd className="pointer-events-none hidden h-5 items-center gap-1 rounded border bg-white px-1.5 font-mono text-[10px] font-medium select-none sm:flex dark:bg-gray-900">
          <span className="text-xs">⌘</span>K
        </kbd>
      </button>

      <ul className="space-y-1">
        <NavLink
          to="/admin"
          icon={<BarChart3 className="h-4 w-4" />}
          exact
        >
          Dashboard
        </NavLink>
        <NavLink
          to="/admin/users"
          icon={<Users className="h-4 w-4" />}
        >
          Users
        </NavLink>
        <NavLink
          to="/admin/audit-logs"
          icon={<FileText className="h-4 w-4" />}
        >
          Audit Logs
        </NavLink>
        <NavLink
          to="/admin/feature-flags"
          icon={<Flag className="h-4 w-4" />}
        >
          Feature Flags
        </NavLink>

        {/* Divider */}
        <li className="py-2">
          <div className="border-t border-gray-200 dark:border-gray-700" />
        </li>

        <NavLink
          to="/admin/health"
          icon={<Activity className="h-4 w-4" />}
        >
          System Health
        </NavLink>
        <NavLink
          to="/admin/announcements"
          icon={<Bell className="h-4 w-4" />}
        >
          Announcements
        </NavLink>
        <NavLink
          to="/admin/email-templates"
          icon={<Mail className="h-4 w-4" />}
        >
          Email Templates
        </NavLink>
        <NavLink
          to="/admin/settings"
          icon={<Settings className="h-4 w-4" />}
        >
          Settings
        </NavLink>
      </ul>
    </>
  );
}

function NavLink({
  to,
  icon,
  children,
  exact = false,
}: {
  to: string;
  icon: React.ReactNode;
  children: React.ReactNode;
  exact?: boolean;
}) {
  const location = useLocation();
  const isActive = exact
    ? location.pathname === to || location.pathname === to + "/"
    : location.pathname === to || location.pathname.startsWith(to + "/");

  return (
    <li>
      <Link
        to={to}
        className={`flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
          isActive
            ? "bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300"
            : "text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
        }`}
      >
        {icon}
        {children}
      </Link>
    </li>
  );
}

function DashboardStats() {
  const {
    data: stats,
    isLoading,
    error,
  } = useQuery<AdminStats>({
    queryKey: ["admin", "stats"],
    queryFn: () => AdminService.getStats(),
  });

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  if (error) {
    return (
      <Card className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950">
        <CardHeader>
          <CardTitle className="text-red-600 dark:text-red-400">Error Loading Stats</CardTitle>
          <CardDescription className="text-red-500 dark:text-red-400">
            {error instanceof Error ? error.message : "Failed to load dashboard statistics"}
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (!stats) return null;

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const activationRate = stats.total_users > 0 ? Math.round((stats.active_users / stats.total_users) * 100) : 0;

  // Mock chart data - in production, this would come from the API
  const chartData = [
    { date: "Mon", users: stats.new_users_today > 0 ? Math.max(1, stats.new_users_today - 5) : 2 },
    { date: "Tue", users: stats.new_users_today > 0 ? Math.max(1, stats.new_users_today - 3) : 4 },
    { date: "Wed", users: stats.new_users_today > 0 ? Math.max(1, stats.new_users_today - 2) : 3 },
    { date: "Thu", users: stats.new_users_today > 0 ? Math.max(1, stats.new_users_today - 1) : 5 },
    { date: "Fri", users: stats.new_users_today > 0 ? stats.new_users_today : 4 },
    { date: "Sat", users: stats.new_users_today > 0 ? Math.max(1, stats.new_users_today + 1) : 6 },
    { date: "Sun", users: stats.new_users_today > 0 ? Math.max(1, stats.new_users_today + 2) : 3 },
  ];

  const chartConfig = {
    users: {
      label: "New Users",
      color: "hsl(var(--chart-1))",
    },
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Dashboard Overview</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Welcome back! Here's what's happening with your platform.
          </p>
        </div>
        <Button
          variant="outline"
          className="hidden gap-2 sm:flex"
        >
          <Download className="h-4 w-4" />
          Export Report
        </Button>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title="Total Users"
          value={stats.total_users}
          change={stats.new_users_this_week}
          changeLabel="this week"
          trend="up"
          icon={<Users className="h-4 w-4" />}
        />
        <MetricCard
          title="Active Users"
          value={stats.active_users}
          change={activationRate}
          changeLabel="activation rate"
          trend="up"
          suffix="%"
          icon={<Activity className="h-4 w-4" />}
        />
        <MetricCard
          title="New Today"
          value={stats.new_users_today}
          change={stats.new_users_this_week}
          changeLabel="this week"
          trend={stats.new_users_today > 0 ? "up" : "neutral"}
          highlight
          icon={<TrendingUp className="h-4 w-4" />}
        />
        <MetricCard
          title="Active Subscriptions"
          value={stats.active_subscriptions}
          change={
            stats.total_subscriptions > 0
              ? Math.round((stats.active_subscriptions / stats.total_subscriptions) * 100)
              : 0
          }
          changeLabel="of total"
          trend="up"
          suffix="%"
          icon={<BarChart3 className="h-4 w-4" />}
        />
      </div>

      {/* Charts Row */}
      <div className="grid gap-4 lg:grid-cols-2">
        {/* User Growth Chart */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">User Growth</CardTitle>
            <CardDescription>New user registrations over the past week</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartContainer
              config={chartConfig}
              className="h-[200px] w-full"
            >
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient
                    id="fillUsers"
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop
                      offset="5%"
                      stopColor="var(--color-users)"
                      stopOpacity={0.8}
                    />
                    <stop
                      offset="95%"
                      stopColor="var(--color-users)"
                      stopOpacity={0.1}
                    />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  strokeDasharray="3 3"
                  vertical={false}
                />
                <XAxis
                  dataKey="date"
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  tickMargin={8}
                  allowDecimals={false}
                />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Area
                  type="monotone"
                  dataKey="users"
                  stroke="var(--color-users)"
                  fill="url(#fillUsers)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ChartContainer>
          </CardContent>
        </Card>

        {/* Recent Activity / Quick Stats */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Platform Summary</CardTitle>
            <CardDescription>Key metrics at a glance</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between rounded-lg bg-gray-50 p-3 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-full bg-blue-100 p-2 dark:bg-blue-900">
                    <Users className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">Verified Users</p>
                    <p className="text-xs text-gray-500">
                      {stats.total_users > 0 ? Math.round((stats.verified_users / stats.total_users) * 100) : 0}% of
                      total
                    </p>
                  </div>
                </div>
                <p className="text-lg font-bold">{stats.verified_users}</p>
              </div>

              <div className="flex items-center justify-between rounded-lg bg-gray-50 p-3 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-full bg-green-100 p-2 dark:bg-green-900">
                    <TrendingUp className="h-4 w-4 text-green-600 dark:text-green-400" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">This Month</p>
                    <p className="text-xs text-gray-500">New registrations</p>
                  </div>
                </div>
                <p className="text-lg font-bold">{stats.new_users_this_month}</p>
              </div>

              <div className="flex items-center justify-between rounded-lg bg-gray-50 p-3 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-full bg-purple-100 p-2 dark:bg-purple-900">
                    <FileText className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">Total Files</p>
                    <p className="text-xs text-gray-500">{formatBytes(stats.total_file_size)} stored</p>
                  </div>
                </div>
                <p className="text-lg font-bold">{stats.total_files}</p>
              </div>

              <div className="flex items-center justify-between rounded-lg bg-gray-50 p-3 dark:bg-gray-800">
                <div className="flex items-center gap-3">
                  <div className="rounded-full bg-orange-100 p-2 dark:bg-orange-900">
                    <Bell className="h-4 w-4 text-orange-600 dark:text-orange-400" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">Canceled</p>
                    <p className="text-xs text-gray-500">Subscriptions</p>
                  </div>
                </div>
                <p className="text-lg font-bold">{stats.canceled_subscriptions}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Users by Role */}
      {stats.users_by_role && Object.keys(stats.users_by_role).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Users by Role</CardTitle>
            <CardDescription>Distribution of user roles across the platform</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
              {Object.entries(stats.users_by_role).map(([role, count]) => (
                <div
                  key={role}
                  className="flex items-center gap-3 rounded-lg border p-3 dark:border-gray-700"
                >
                  <Avatar className="h-8 w-8">
                    <AvatarFallback className="text-xs">{role.charAt(0).toUpperCase()}</AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-sm font-medium capitalize">{role.replace("_", " ")}</p>
                    <p className="text-lg font-bold">{count}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function MetricCard({
  title,
  value,
  change,
  changeLabel,
  trend,
  highlight = false,
  suffix,
  icon,
}: {
  title: string;
  value: string | number;
  change?: number;
  changeLabel?: string;
  trend?: "up" | "down" | "neutral";
  highlight?: boolean;
  suffix?: string;
  icon?: React.ReactNode;
}) {
  return (
    <Card className={highlight ? "border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/50" : ""}>
      <CardContent className="p-4">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">{title}</p>
          {icon && (
            <div
              className={`rounded-full p-1.5 ${highlight ? "bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-400" : "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400"}`}
            >
              {icon}
            </div>
          )}
        </div>
        <p
          className={`mt-2 text-2xl font-bold ${highlight ? "text-blue-600 dark:text-blue-400" : "text-gray-900 dark:text-white"}`}
        >
          {value}
        </p>
        {change !== undefined && (
          <div className="mt-1 flex items-center gap-1 text-xs">
            {trend === "up" && <ArrowUpRight className="h-3 w-3 text-green-500" />}
            {trend === "down" && <ArrowDownRight className="h-3 w-3 text-red-500" />}
            <span
              className={
                trend === "up"
                  ? "text-green-600 dark:text-green-400"
                  : trend === "down"
                    ? "text-red-600 dark:text-red-400"
                    : "text-gray-500"
              }
            >
              {change}
              {suffix}
            </span>
            <span className="text-gray-500 dark:text-gray-400">{changeLabel}</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <div className="h-8 w-48 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
        <div className="mt-2 h-4 w-64 animate-pulse rounded bg-gray-100 dark:bg-gray-800" />
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardContent className="p-4">
              <div className="h-4 w-24 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
              <div className="mt-2 h-8 w-16 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
            </CardContent>
          </Card>
        ))}
      </div>
      <div className="grid gap-4 lg:grid-cols-2">
        {[...Array(2)].map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <div className="h-5 w-32 animate-pulse rounded bg-gray-200 dark:bg-gray-700" />
            </CardHeader>
            <CardContent>
              <div className="h-[200px] animate-pulse rounded bg-gray-100 dark:bg-gray-800" />
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
