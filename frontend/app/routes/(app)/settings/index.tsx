import { Card, CardContent } from "@/components/ui/card";
import { SettingsLayout } from "@/layouts/SettingsLayout";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Bell, ChevronRight, History, Key, KeyRound, Link2, Palette, Shield, User } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/settings/")({
  component: SettingsPage,
});

const settingsNavItems = [
  {
    key: "profile",
    href: "/settings/profile",
    icon: User,
    gradient: "from-blue-500/20 to-blue-600/10",
    iconColor: "text-blue-600 dark:text-blue-400",
    borderColor: "group-hover:border-blue-500/50",
  },
  {
    key: "security",
    href: "/settings/security",
    icon: Shield,
    gradient: "from-red-500/20 to-red-600/10",
    iconColor: "text-red-600 dark:text-red-400",
    borderColor: "group-hover:border-red-500/50",
  },
  {
    key: "loginHistory",
    href: "/settings/login-history",
    icon: History,
    gradient: "from-amber-500/20 to-amber-600/10",
    iconColor: "text-amber-600 dark:text-amber-400",
    borderColor: "group-hover:border-amber-500/50",
  },
  {
    key: "preferences",
    href: "/settings/preferences",
    icon: Palette,
    gradient: "from-purple-500/20 to-purple-600/10",
    iconColor: "text-purple-600 dark:text-purple-400",
    borderColor: "group-hover:border-purple-500/50",
  },
  {
    key: "notifications",
    href: "/settings/notifications",
    icon: Bell,
    gradient: "from-orange-500/20 to-orange-600/10",
    iconColor: "text-orange-600 dark:text-orange-400",
    borderColor: "group-hover:border-orange-500/50",
  },
  {
    key: "privacy",
    href: "/settings/privacy",
    icon: Key,
    gradient: "from-emerald-500/20 to-emerald-600/10",
    iconColor: "text-emerald-600 dark:text-emerald-400",
    borderColor: "group-hover:border-emerald-500/50",
  },
  {
    key: "connectedAccounts",
    href: "/settings/connected-accounts",
    icon: Link2,
    gradient: "from-cyan-500/20 to-cyan-600/10",
    iconColor: "text-cyan-600 dark:text-cyan-400",
    borderColor: "group-hover:border-cyan-500/50",
  },
  {
    key: "apiKeys",
    href: "/settings/api-keys",
    icon: KeyRound,
    gradient: "from-fuchsia-500/20 to-fuchsia-600/10",
    iconColor: "text-fuchsia-600 dark:text-fuchsia-400",
    borderColor: "group-hover:border-fuchsia-500/50",
  },
];

function SettingsPage() {
  const { t } = useTranslation("settings");

  return (
    <SettingsLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">{t("title")}</h1>
          <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          {settingsNavItems.map((item, index) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                to={item.href}
                className="group"
              >
                <Card
                  className={`h-full transition-all duration-200 hover:-translate-y-1 hover:shadow-lg ${item.borderColor} animate-fade-in-up`}
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <CardContent className="p-6">
                    <div className="flex items-start gap-4">
                      <div
                        className={`bg-gradient-to-br ${item.gradient} rounded-xl p-2.5 transition-transform duration-200 group-hover:scale-110`}
                      >
                        <Icon className={`h-5 w-5 ${item.iconColor}`} />
                      </div>
                      <div className="flex-1">
                        <h3 className="font-semibold">{t(`nav.${item.key}.title` as never)}</h3>
                        <p className="text-muted-foreground mt-1 text-sm">
                          {t(`nav.${item.key}.description` as never)}
                        </p>
                      </div>
                      <ChevronRight className="text-muted-foreground group-hover:text-foreground h-5 w-5 transition-transform duration-200 group-hover:translate-x-1" />
                    </div>
                  </CardContent>
                </Card>
              </Link>
            );
          })}
        </div>
      </div>
    </SettingsLayout>
  );
}
