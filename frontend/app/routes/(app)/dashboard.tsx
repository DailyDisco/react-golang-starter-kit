import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/auth/useAuth";
import { createFileRoute, Link } from "@tanstack/react-router";
import { CreditCard, Settings, Shield, User } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/dashboard")({
  component: Dashboard,
});

function Dashboard() {
  const { t } = useTranslation("dashboard");
  const { user } = useAuth();

  const quickLinks = [
    {
      key: "profile",
      href: "/settings/profile",
      icon: User,
    },
    {
      key: "security",
      href: "/settings/security",
      icon: Shield,
    },
    {
      key: "billing",
      href: "/billing",
      icon: CreditCard,
    },
    {
      key: "settings",
      href: "/settings",
      icon: Settings,
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

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        {quickLinks.map((link) => {
          const Icon = link.icon;
          const title = t(`quickLinks.${link.key}.title`);
          return (
            <Card
              key={link.href}
              className="transition-shadow hover:shadow-md"
            >
              <CardHeader className="flex flex-row items-center gap-4">
                <div className="bg-primary/10 rounded-lg p-2">
                  <Icon className="text-primary h-6 w-6" />
                </div>
                <div className="flex-1">
                  <CardTitle className="text-lg">{title}</CardTitle>
                  <CardDescription>{t(`quickLinks.${link.key}.description`)}</CardDescription>
                </div>
              </CardHeader>
              <CardContent>
                <Button
                  asChild
                  variant="outline"
                  className="w-full"
                >
                  <Link to={link.href}>{t("goTo", { page: title })}</Link>
                </Button>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
