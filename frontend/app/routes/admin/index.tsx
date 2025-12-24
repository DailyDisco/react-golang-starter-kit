import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, Outlet, useLocation } from "@tanstack/react-router";
import { BarChart3, FileText, Flag, Settings, Shield, Users } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/card";
import { requireAdmin } from "../../lib/guards";
import { AdminService, type AdminStats } from "../../services/admin";

export const Route = createFileRoute("/admin/")({
  // Role-based access control: only admin and super_admin can access
  beforeLoad: () => requireAdmin(),
  component: AdminDashboard,
});

function AdminDashboard() {
  const location = useLocation();
  const isIndex = location.pathname === "/admin" || location.pathname === "/admin/";

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="flex">
        {/* Sidebar */}
        <nav className="min-h-screen w-64 bg-white p-4 shadow-sm">
          <div className="mb-8">
            <h1 className="flex items-center gap-2 text-xl font-bold text-gray-900">
              <Shield className="h-6 w-6" />
              Admin Panel
            </h1>
          </div>
          <ul className="space-y-2">
            <NavLink
              to="/admin"
              icon={<BarChart3 className="h-4 w-4" />}
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
            <NavLink
              to="/admin/settings"
              icon={<Settings className="h-4 w-4" />}
            >
              Settings
            </NavLink>
          </ul>
        </nav>

        {/* Main content */}
        <main className="flex-1 p-8">{isIndex ? <DashboardStats /> : <Outlet />}</main>
      </div>
    </div>
  );
}

function NavLink({ to, icon, children }: { to: string; icon: React.ReactNode; children: React.ReactNode }) {
  const location = useLocation();
  const isActive = location.pathname === to || (to !== "/admin" && location.pathname.startsWith(to));

  return (
    <li>
      <Link
        to={to}
        className={`flex items-center gap-2 rounded-lg px-4 py-2 transition-colors ${
          isActive ? "bg-blue-100 text-blue-700" : "text-gray-600 hover:bg-gray-100"
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
    return (
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(8)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <div className="h-4 w-24 animate-pulse rounded bg-gray-200" />
            </CardHeader>
            <CardContent>
              <div className="h-8 w-16 animate-pulse rounded bg-gray-200" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Card className="border-red-200 bg-red-50">
        <CardHeader>
          <CardTitle className="text-red-600">Error Loading Stats</CardTitle>
          <CardDescription className="text-red-500">
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

  return (
    <div className="space-y-8">
      <div>
        <h2 className="mb-4 text-2xl font-bold text-gray-900">Dashboard Overview</h2>
      </div>

      {/* User Stats */}
      <div>
        <h3 className="mb-4 text-lg font-semibold text-gray-700">Users</h3>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="Total Users"
            value={stats.total_users}
          />
          <StatCard
            title="Active Users"
            value={stats.active_users}
          />
          <StatCard
            title="Verified Users"
            value={stats.verified_users}
          />
          <StatCard
            title="New Today"
            value={stats.new_users_today}
            highlight
          />
        </div>
      </div>

      {/* Growth Stats */}
      <div>
        <h3 className="mb-4 text-lg font-semibold text-gray-700">Growth</h3>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <StatCard
            title="New This Week"
            value={stats.new_users_this_week}
          />
          <StatCard
            title="New This Month"
            value={stats.new_users_this_month}
          />
          <StatCard
            title="Activation Rate"
            value={`${stats.total_users > 0 ? Math.round((stats.active_users / stats.total_users) * 100) : 0}%`}
          />
        </div>
      </div>

      {/* Subscription Stats */}
      <div>
        <h3 className="mb-4 text-lg font-semibold text-gray-700">Subscriptions</h3>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <StatCard
            title="Total Subscriptions"
            value={stats.total_subscriptions}
          />
          <StatCard
            title="Active Subscriptions"
            value={stats.active_subscriptions}
            highlight
          />
          <StatCard
            title="Canceled"
            value={stats.canceled_subscriptions}
          />
        </div>
      </div>

      {/* File Stats */}
      <div>
        <h3 className="mb-4 text-lg font-semibold text-gray-700">Files</h3>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <StatCard
            title="Total Files"
            value={stats.total_files}
          />
          <StatCard
            title="Total Storage"
            value={formatBytes(stats.total_file_size)}
          />
        </div>
      </div>

      {/* Users by Role */}
      {stats.users_by_role && Object.keys(stats.users_by_role).length > 0 && (
        <div>
          <h3 className="mb-4 text-lg font-semibold text-gray-700">Users by Role</h3>
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
            {Object.entries(stats.users_by_role).map(([role, count]) => (
              <StatCard
                key={role}
                title={role.replace("_", " ").toUpperCase()}
                value={count}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({ title, value, highlight = false }: { title: string; value: string | number; highlight?: boolean }) {
  return (
    <Card className={highlight ? "border-blue-200 bg-blue-50" : ""}>
      <CardHeader className="pb-2">
        <CardDescription className="text-sm text-gray-500">{title}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className={`text-2xl font-bold ${highlight ? "text-blue-600" : "text-gray-900"}`}>{value}</p>
      </CardContent>
    </Card>
  );
}
