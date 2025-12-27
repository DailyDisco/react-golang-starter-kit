import { createFileRoute, Link } from "@tanstack/react-router";
import { CreditCard, Settings, Shield, User } from "lucide-react";

import { Button } from "../../../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../../components/ui/card";
import { useAuth } from "../../../hooks/auth/useAuth";

export const Route = createFileRoute("/(auth)/_protected/dashboard")({
  component: ProtectedDashboard,
});

function ProtectedDashboard() {
  const { user } = useAuth();

  const quickLinks = [
    {
      title: "Profile",
      description: "Update your personal information and avatar",
      href: "/settings/profile",
      icon: User,
    },
    {
      title: "Security",
      description: "Manage passwords, 2FA, and active sessions",
      href: "/settings/security",
      icon: Shield,
    },
    {
      title: "Billing",
      description: "View and manage your subscription",
      href: "/billing",
      icon: CreditCard,
    },
    {
      title: "Settings",
      description: "Preferences, notifications, and privacy",
      href: "/settings",
      icon: Settings,
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Welcome back{user?.name ? `, ${user.name.split(" ")[0]}` : ""}!</h1>
        <p className="text-muted-foreground mt-2">Here's an overview of your account.</p>
      </div>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        {quickLinks.map((link) => {
          const Icon = link.icon;
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
                  <CardTitle className="text-lg">{link.title}</CardTitle>
                  <CardDescription>{link.description}</CardDescription>
                </div>
              </CardHeader>
              <CardContent>
                <Button
                  asChild
                  variant="outline"
                  className="w-full"
                >
                  <Link to={link.href}>Go to {link.title}</Link>
                </Button>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
