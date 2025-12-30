import { Card, CardContent } from "@/components/ui/card";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Bell, ChevronRight, History, Key, Link2, Palette, Shield, User } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/(app)/settings/")({
  component: SettingsPage,
});

const settingsNavItems = [
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
    key: "loginHistory",
    href: "/settings/login-history",
    icon: History,
  },
  {
    key: "preferences",
    href: "/settings/preferences",
    icon: Palette,
  },
  {
    key: "notifications",
    href: "/settings/notifications",
    icon: Bell,
  },
  {
    key: "privacy",
    href: "/settings/privacy",
    icon: Key,
  },
  {
    key: "connectedAccounts",
    href: "/settings/connected-accounts",
    icon: Link2,
  },
];

function SettingsPage() {
  const { t } = useTranslation("settings");

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <p className="text-muted-foreground mt-2">{t("subtitle")}</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {settingsNavItems.map((item) => {
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              to={item.href}
            >
              <Card className="hover:border-primary/50 h-full transition-all hover:shadow-md">
                <CardContent className="p-6">
                  <div className="flex items-start gap-4">
                    <div className="bg-primary/10 rounded-lg p-2.5">
                      <Icon className="text-primary h-5 w-5" />
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold">{t(`nav.${item.key}.title`)}</h3>
                      <p className="text-muted-foreground mt-1 text-sm">{t(`nav.${item.key}.description`)}</p>
                    </div>
                    <ChevronRight className="text-muted-foreground h-5 w-5" />
                  </div>
                </CardContent>
              </Card>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
