import { ActivityFeed, useMockActivities } from "@/components/dashboard";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/auth/useAuth";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ChevronRight, CreditCard, Settings, Shield, User } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/dashboard")({
  component: Dashboard,
});

function Dashboard() {
  const { t } = useTranslation("dashboard");
  const { user } = useAuth();
  const mockActivities = useMockActivities();

  const quickLinks = [
    {
      key: "profile",
      href: "/settings/profile",
      icon: User,
      gradient: "from-blue-500/20 to-blue-600/10",
      iconColor: "text-blue-600 dark:text-blue-400",
    },
    {
      key: "security",
      href: "/settings/security",
      icon: Shield,
      gradient: "from-red-500/20 to-red-600/10",
      iconColor: "text-red-600 dark:text-red-400",
    },
    {
      key: "billing",
      href: "/billing",
      icon: CreditCard,
      gradient: "from-green-500/20 to-green-600/10",
      iconColor: "text-green-600 dark:text-green-400",
    },
    {
      key: "settings",
      href: "/settings",
      icon: Settings,
      gradient: "from-purple-500/20 to-purple-600/10",
      iconColor: "text-purple-600 dark:text-purple-400",
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">
          {t("welcome")}
          {user?.name ? `, ${user.name.split(" ")[0]}` : ""}!
        </h1>
        <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Quick Links */}
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:col-span-2">
          {quickLinks.map((link, index) => {
            const Icon = link.icon;
            const title = t(`quickLinks.${link.key}.title`);
            return (
              <Card
                key={link.href}
                interactive
                className="group animate-fade-in-up"
                style={{ animationDelay: `${index * 75}ms` }}
              >
                <CardHeader className="flex flex-row items-center gap-4">
                  <div
                    className={`bg-gradient-to-br ${link.gradient} rounded-xl p-3 transition-transform duration-200 group-hover:scale-110`}
                  >
                    <Icon className={`h-6 w-6 ${link.iconColor}`} />
                  </div>
                  <div className="flex-1">
                    <CardTitle className="text-lg">{title}</CardTitle>
                    <CardDescription>{t(`quickLinks.${link.key}.description`)}</CardDescription>
                  </div>
                  <ChevronRight className="text-muted-foreground h-5 w-5 transition-transform duration-200 group-hover:translate-x-1" />
                </CardHeader>
                <CardContent>
                  <Button
                    asChild
                    variant="outline"
                    className="group-hover:border-primary/50 w-full"
                  >
                    <Link to={link.href}>{t("goTo", { page: title })}</Link>
                  </Button>
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Activity Feed */}
        <div
          className="animate-fade-in-up lg:col-span-1"
          style={{ animationDelay: "300ms" }}
        >
          <ActivityFeed
            activities={mockActivities}
            title={t("activityFeed.title", "Recent Activity")}
          />
        </div>
      </div>
    </div>
  );
}
