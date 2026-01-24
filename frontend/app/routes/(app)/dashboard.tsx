import { ActivityFeed, UsageSummaryCard } from "@/components/dashboard";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/auth/useAuth";
import { useMyActivity } from "@/hooks/queries";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
  ArrowRight,
  Bell,
  CheckCircle,
  Clock,
  CreditCard,
  FileText,
  Key,
  Settings,
  Shield,
  Sparkles,
  User,
  Zap,
} from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/dashboard")({
  component: Dashboard,
});

function getUserInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

function Dashboard() {
  const { t } = useTranslation("dashboard");
  const { user } = useAuth();
  const { data: activities = [], isLoading: isLoadingActivity } = useMyActivity(10);

  const quickActions = [
    {
      key: "profile",
      href: "/settings/profile",
      icon: User,
      color: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
      borderHover: "hover:border-blue-500/50",
    },
    {
      key: "security",
      href: "/settings/security",
      icon: Shield,
      color: "bg-red-500/10 text-red-600 dark:text-red-400",
      borderHover: "hover:border-red-500/50",
    },
    {
      key: "billing",
      href: "/billing",
      icon: CreditCard,
      color: "bg-green-500/10 text-green-600 dark:text-green-400",
      borderHover: "hover:border-green-500/50",
    },
    {
      key: "settings",
      href: "/settings",
      icon: Settings,
      color: "bg-purple-500/10 text-purple-600 dark:text-purple-400",
      borderHover: "hover:border-purple-500/50",
    },
  ];

  const statsCards = [
    {
      label: t("stats.accountStatus", "Account Status"),
      value: user?.is_active ? t("stats.active", "Active") : t("stats.inactive", "Inactive"),
      icon: CheckCircle,
      color: user?.is_active ? "text-green-600" : "text-red-600",
      bgColor: user?.is_active ? "bg-green-500/10" : "bg-red-500/10",
    },
    {
      label: t("stats.memberSince", "Member Since"),
      value: user?.created_at
        ? new Date(user.created_at).toLocaleDateString(undefined, { month: "short", year: "numeric" })
        : "-",
      icon: Clock,
      color: "text-blue-600",
      bgColor: "bg-blue-500/10",
    },
    {
      label: t("stats.emailVerified", "Email Verified"),
      value: user?.email_verified ? t("stats.verified", "Verified") : t("stats.pending", "Pending"),
      icon: user?.email_verified ? CheckCircle : Clock,
      color: user?.email_verified ? "text-green-600" : "text-amber-600",
      bgColor: user?.email_verified ? "bg-green-500/10" : "bg-amber-500/10",
    },
    {
      label: t("stats.role", "Role"),
      value: user?.role ? user.role.charAt(0).toUpperCase() + user.role.slice(1) : "User",
      icon: Key,
      color: "text-purple-600",
      bgColor: "bg-purple-500/10",
    },
  ];

  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <div className="from-primary/5 via-primary/10 to-secondary/5 relative overflow-hidden rounded-2xl bg-gradient-to-r p-6 md:p-8">
        <div className="bg-grid-pattern absolute inset-0 opacity-5" />
        <div className="relative flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
          {/* User Info */}
          <div className="flex items-center gap-4">
            <Avatar className="ring-primary/20 h-16 w-16 ring-4 md:h-20 md:w-20">
              <AvatarImage
                src={user?.avatar_url || ""}
                alt={user?.name}
              />
              <AvatarFallback className="bg-primary/10 text-primary text-xl font-bold">
                {getUserInitials(user?.name || "U")}
              </AvatarFallback>
            </Avatar>
            <div>
              <h1 className="text-2xl font-bold md:text-3xl">
                {t("welcome")}
                {user?.name ? `, ${user.name.split(" ")[0]}` : ""}!
              </h1>
              <p className="text-muted-foreground mt-1">{t("subtitle")}</p>
              <div className="mt-2 flex flex-wrap gap-2">
                <Badge
                  variant="secondary"
                  className="capitalize"
                >
                  {user?.role || "User"}
                </Badge>
                {user?.email_verified && (
                  <Badge
                    variant="outline"
                    className="border-green-500/50 text-green-600"
                  >
                    <CheckCircle className="mr-1 h-3 w-3" />
                    {t("verified", "Verified")}
                  </Badge>
                )}
              </div>
            </div>
          </div>

          {/* Quick CTA */}
          <div className="flex flex-wrap gap-3">
            <Button
              asChild
              variant="outline"
              className="gap-2"
            >
              <Link to="/settings/profile">
                <User className="h-4 w-4" />
                {t("editProfile", "Edit Profile")}
              </Link>
            </Button>
            <Button
              asChild
              className="gap-2"
            >
              <Link to="/settings">
                <Settings className="h-4 w-4" />
                {t("settings", "Settings")}
              </Link>
            </Button>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        {statsCards.map((stat, index) => {
          const Icon = stat.icon;
          return (
            <Card
              key={index}
              className="animate-fade-in-up"
              style={{ animationDelay: `${index * 50}ms` }}
            >
              <CardContent className="flex items-center gap-4 p-4">
                <div className={`rounded-xl p-2.5 ${stat.bgColor}`}>
                  <Icon className={`h-5 w-5 ${stat.color}`} />
                </div>
                <div>
                  <p className="text-muted-foreground text-xs font-medium">{stat.label}</p>
                  <p className="text-lg font-semibold">{stat.value}</p>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Main Content Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Quick Actions */}
        <div className="space-y-4 lg:col-span-2">
          <div className="flex items-center justify-between">
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <Zap className="text-primary h-5 w-5" />
              {t("quickActions.title", "Quick Actions")}
            </h2>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            {quickActions.map((action, index) => {
              const Icon = action.icon;
              return (
                <Link
                  key={action.href}
                  to={action.href}
                  className="group"
                >
                  <Card
                    className={`h-full transition-all duration-200 hover:-translate-y-1 hover:shadow-lg ${action.borderHover} animate-fade-in-up`}
                    style={{ animationDelay: `${(index + 4) * 50}ms` }}
                  >
                    <CardContent className="flex items-center gap-4 p-4">
                      <div className={`rounded-xl p-3 ${action.color} transition-transform group-hover:scale-110`}>
                        <Icon className="h-5 w-5" />
                      </div>
                      <div className="flex-1">
                        <p className="font-medium">{t(`quickLinks.${action.key}.title` as never) as string}</p>
                        <p className="text-muted-foreground text-sm">
                          {t(`quickLinks.${action.key}.description` as never) as string}
                        </p>
                      </div>
                      <ArrowRight className="text-muted-foreground h-4 w-4 transition-transform group-hover:translate-x-1" />
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
          </div>
        </div>

        {/* Right Column: Activity Feed + Usage Summary */}
        <div className="space-y-6 lg:col-span-1">
          <div
            className="animate-fade-in-up"
            style={{ animationDelay: "400ms" }}
          >
            <ActivityFeed
              activities={activities}
              isLoading={isLoadingActivity}
              title={t("activityFeed.title", "Recent Activity")}
            />
          </div>
          <div
            className="animate-fade-in-up"
            style={{ animationDelay: "450ms" }}
          >
            <UsageSummaryCard />
          </div>
        </div>
      </div>

      {/* Getting Started / Tips Section */}
      <Card
        className="animate-fade-in-up border-dashed"
        style={{ animationDelay: "450ms" }}
      >
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Sparkles className="text-primary h-5 w-5" />
            {t("gettingStarted.title", "Getting Started")}
          </CardTitle>
          <CardDescription>
            {t("gettingStarted.subtitle", "Complete these steps to make the most of your account")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <GettingStartedItem
              icon={User}
              title={t("gettingStarted.completeProfile", "Complete your profile")}
              completed={Boolean(user?.name && user?.email)}
              href="/settings/profile"
            />
            <GettingStartedItem
              icon={Shield}
              title={t("gettingStarted.enable2FA", "Enable 2FA")}
              completed={false}
              href="/settings/security"
            />
            <GettingStartedItem
              icon={Bell}
              title={t("gettingStarted.setNotifications", "Set notifications")}
              completed={false}
              href="/settings/notifications"
            />
            <GettingStartedItem
              icon={FileText}
              title={t("gettingStarted.exploreFeatures", "Explore features")}
              completed={false}
              href="/settings"
            />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function GettingStartedItem({
  icon: Icon,
  title,
  completed,
  href,
}: {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  completed: boolean;
  href: string;
}) {
  return (
    <Link
      to={href}
      className="group"
    >
      <div
        className={`flex items-center gap-3 rounded-lg border p-3 transition-colors ${
          completed ? "border-green-500/50 bg-green-500/5" : "hover:border-primary/50 hover:bg-accent"
        }`}
      >
        <div
          className={`rounded-full p-1.5 ${
            completed ? "bg-green-500/20 text-green-600" : "bg-muted text-muted-foreground group-hover:text-primary"
          }`}
        >
          {completed ? <CheckCircle className="h-4 w-4" /> : <Icon className="h-4 w-4" />}
        </div>
        <span className={`text-sm ${completed ? "text-green-600 line-through" : "font-medium"}`}>{title}</span>
      </div>
    </Link>
  );
}
