import { createFileRoute, Link } from "@tanstack/react-router";
import {
  Activity,
  Bell,
  Bot,
  Brain,
  Building2,
  CreditCard,
  FileText,
  Flag,
  Globe,
  HardDrive,
  Image,
  Key,
  LayoutDashboard,
  Lock,
  Mail,
  MessageSquare,
  MonitorCheck,
  Radio,
  ScrollText,
  Settings,
  Shield,
  ShieldCheck,
  Smartphone,
  Sparkles,
  Users,
  Wrench,
  Zap,
} from "lucide-react";
import { useTranslation } from "react-i18next";

import { API_BASE_URL } from "../services";

export const Route = createFileRoute("/")({
  component: Home,
});

interface FeatureItem {
  icon: React.ReactNode;
  name: string;
  description: string;
}

interface FeatureCategory {
  title: string;
  icon: React.ReactNode;
  iconColor: string;
  features: FeatureItem[];
}

function Home() {
  const { t } = useTranslation("landing");

  const featureCategories: FeatureCategory[] = [
    {
      title: t("features.categories.authSecurity.title"),
      icon: <Shield className="h-6 w-6" />,
      iconColor: "text-primary",
      features: [
        {
          icon: <Key className="h-4 w-4" />,
          name: t("features.categories.authSecurity.emailPassword.name"),
          description: t("features.categories.authSecurity.emailPassword.description"),
        },
        {
          icon: <Users className="h-4 w-4" />,
          name: t("features.categories.authSecurity.oauth.name"),
          description: t("features.categories.authSecurity.oauth.description"),
        },
        {
          icon: <Smartphone className="h-4 w-4" />,
          name: t("features.categories.authSecurity.twoFactor.name"),
          description: t("features.categories.authSecurity.twoFactor.description"),
        },
        {
          icon: <MonitorCheck className="h-4 w-4" />,
          name: t("features.categories.authSecurity.sessionManagement.name"),
          description: t("features.categories.authSecurity.sessionManagement.description"),
        },
        {
          icon: <ScrollText className="h-4 w-4" />,
          name: t("features.categories.authSecurity.loginHistory.name"),
          description: t("features.categories.authSecurity.loginHistory.description"),
        },
        {
          icon: <ShieldCheck className="h-4 w-4" />,
          name: t("features.categories.authSecurity.ipBlocklist.name"),
          description: t("features.categories.authSecurity.ipBlocklist.description"),
        },
      ],
    },
    {
      title: t("features.categories.organizations.title"),
      icon: <Building2 className="h-6 w-6" />,
      iconColor: "text-violet-500",
      features: [
        {
          icon: <Building2 className="h-4 w-4" />,
          name: t("features.categories.organizations.multiTenant.name"),
          description: t("features.categories.organizations.multiTenant.description"),
        },
        {
          icon: <Users className="h-4 w-4" />,
          name: t("features.categories.organizations.teamManagement.name"),
          description: t("features.categories.organizations.teamManagement.description"),
        },
        {
          icon: <Shield className="h-4 w-4" />,
          name: t("features.categories.organizations.roleBasedAccess.name"),
          description: t("features.categories.organizations.roleBasedAccess.description"),
        },
        {
          icon: <Mail className="h-4 w-4" />,
          name: t("features.categories.organizations.invitations.name"),
          description: t("features.categories.organizations.invitations.description"),
        },
      ],
    },
    {
      title: t("features.categories.billing.title"),
      icon: <CreditCard className="h-6 w-6" />,
      iconColor: "text-success",
      features: [
        {
          icon: <CreditCard className="h-4 w-4" />,
          name: t("features.categories.billing.stripeIntegration.name"),
          description: t("features.categories.billing.stripeIntegration.description"),
        },
        {
          icon: <FileText className="h-4 w-4" />,
          name: t("features.categories.billing.subscriptionPlans.name"),
          description: t("features.categories.billing.subscriptionPlans.description"),
        },
        {
          icon: <Settings className="h-4 w-4" />,
          name: t("features.categories.billing.customerPortal.name"),
          description: t("features.categories.billing.customerPortal.description"),
        },
      ],
    },
    {
      title: t("features.categories.admin.title"),
      icon: <LayoutDashboard className="h-6 w-6" />,
      iconColor: "text-info",
      features: [
        {
          icon: <LayoutDashboard className="h-4 w-4" />,
          name: t("features.categories.admin.dashboard.name"),
          description: t("features.categories.admin.dashboard.description"),
        },
        {
          icon: <Users className="h-4 w-4" />,
          name: t("features.categories.admin.userManagement.name"),
          description: t("features.categories.admin.userManagement.description"),
        },
        {
          icon: <Flag className="h-4 w-4" />,
          name: t("features.categories.admin.featureFlags.name"),
          description: t("features.categories.admin.featureFlags.description"),
        },
        {
          icon: <ScrollText className="h-4 w-4" />,
          name: t("features.categories.admin.auditLogging.name"),
          description: t("features.categories.admin.auditLogging.description"),
        },
        {
          icon: <Bell className="h-4 w-4" />,
          name: t("features.categories.admin.announcements.name"),
          description: t("features.categories.admin.announcements.description"),
        },
        {
          icon: <Mail className="h-4 w-4" />,
          name: t("features.categories.admin.emailTemplates.name"),
          description: t("features.categories.admin.emailTemplates.description"),
        },
      ],
    },
    {
      title: t("features.categories.realtime.title"),
      icon: <Radio className="h-6 w-6" />,
      iconColor: "text-rose-500",
      features: [
        {
          icon: <Radio className="h-4 w-4" />,
          name: t("features.categories.realtime.websocket.name"),
          description: t("features.categories.realtime.websocket.description"),
        },
        {
          icon: <Globe className="h-4 w-4" />,
          name: t("features.categories.realtime.i18n.name"),
          description: t("features.categories.realtime.i18n.description"),
        },
        {
          icon: <Bell className="h-4 w-4" />,
          name: t("features.categories.realtime.pushNotifications.name"),
          description: t("features.categories.realtime.pushNotifications.description"),
        },
      ],
    },
    {
      title: t("features.categories.infrastructure.title"),
      icon: <HardDrive className="h-6 w-6" />,
      iconColor: "text-warning",
      features: [
        {
          icon: <HardDrive className="h-4 w-4" />,
          name: t("features.categories.infrastructure.fileStorage.name"),
          description: t("features.categories.infrastructure.fileStorage.description"),
        },
        {
          icon: <Activity className="h-4 w-4" />,
          name: t("features.categories.infrastructure.prometheusMetrics.name"),
          description: t("features.categories.infrastructure.prometheusMetrics.description"),
        },
        {
          icon: <MonitorCheck className="h-4 w-4" />,
          name: t("features.categories.infrastructure.healthMonitoring.name"),
          description: t("features.categories.infrastructure.healthMonitoring.description"),
        },
        {
          icon: <Lock className="h-4 w-4" />,
          name: t("features.categories.infrastructure.rateLimiting.name"),
          description: t("features.categories.infrastructure.rateLimiting.description"),
        },
        {
          icon: <Settings className="h-4 w-4" />,
          name: t("features.categories.infrastructure.userPreferences.name"),
          description: t("features.categories.infrastructure.userPreferences.description"),
        },
      ],
    },
    {
      title: t("features.categories.ai.title"),
      icon: <Sparkles className="h-6 w-6" />,
      iconColor: "text-fuchsia-500",
      features: [
        {
          icon: <MessageSquare className="h-4 w-4" />,
          name: t("features.categories.ai.chat.name"),
          description: t("features.categories.ai.chat.description"),
        },
        {
          icon: <Zap className="h-4 w-4" />,
          name: t("features.categories.ai.streaming.name"),
          description: t("features.categories.ai.streaming.description"),
        },
        {
          icon: <Image className="h-4 w-4" />,
          name: t("features.categories.ai.vision.name"),
          description: t("features.categories.ai.vision.description"),
        },
        {
          icon: <Brain className="h-4 w-4" />,
          name: t("features.categories.ai.embeddings.name"),
          description: t("features.categories.ai.embeddings.description"),
        },
        {
          icon: <Wrench className="h-4 w-4" />,
          name: t("features.categories.ai.functionCalling.name"),
          description: t("features.categories.ai.functionCalling.description"),
        },
      ],
    },
  ];

  return (
    <main className="bg-muted">
      {/* Hero Section */}
      <section className="px-4 py-16 md:py-24">
        <div className="mx-auto max-w-4xl text-center">
          <h1 className="text-foreground mb-6 text-4xl font-bold tracking-tight md:text-5xl lg:text-6xl">
            {t("hero.title")}
            <span className="text-primary block">{t("hero.titleHighlight")}</span>
          </h1>
          <p className="text-muted-foreground mx-auto mb-8 max-w-2xl text-lg md:text-xl">{t("hero.subtitle")}</p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              to="/demo"
              search={{}}
              className="bg-primary text-primary-foreground hover:bg-primary/90 inline-flex w-full items-center justify-center rounded-lg px-8 py-3 font-semibold shadow-lg transition-all duration-200 hover:scale-[1.02] hover:shadow-[0_0_20px_oklch(0.55_0.18_250/0.4)] sm:w-auto"
            >
              {t("hero.tryDemo")}
            </Link>
            <Link
              to="/pricing"
              search={{}}
              className="border-border bg-card text-card-foreground hover:bg-accent inline-flex w-full items-center justify-center rounded-lg border px-8 py-3 font-semibold shadow-sm transition-all duration-200 hover:scale-[1.02] hover:shadow-md sm:w-auto"
            >
              {t("hero.viewPricing")}
            </Link>
          </div>
        </div>
      </section>

      {/* Feature Categories */}
      <section className="px-4 py-16">
        <div className="mx-auto max-w-6xl">
          <div className="mb-12 text-center">
            <h2 className="text-foreground mb-4 text-3xl font-bold">{t("features.title")}</h2>
            <p className="text-muted-foreground mx-auto max-w-2xl">{t("features.subtitle")}</p>
          </div>

          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {featureCategories.map((category, index) => (
              <div
                key={category.title}
                className="group bg-card border-t-primary/60 animate-fade-in-up rounded-xl border-t-4 p-6 shadow-md transition-all duration-200 hover:-translate-y-1 hover:shadow-xl"
                style={{ animationDelay: `${index * 75}ms` }}
              >
                <div className="mb-4 flex items-center gap-3">
                  <div className={`${category.iconColor} transition-transform duration-200 group-hover:scale-110`}>
                    {category.icon}
                  </div>
                  <h3 className="text-card-foreground text-lg font-semibold">{category.title}</h3>
                </div>
                <div className="space-y-3">
                  {category.features.map((feature) => (
                    <div
                      key={feature.name}
                      className="flex items-start gap-2"
                    >
                      <div className="text-muted-foreground mt-0.5">{feature.icon}</div>
                      <div>
                        <p className="text-card-foreground font-medium">{feature.name}</p>
                        <p className="text-muted-foreground text-sm">{feature.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Tech Stack */}
      <section className="bg-card px-4 py-16">
        <div className="mx-auto max-w-5xl">
          <div className="mb-12 text-center">
            <h2 className="text-foreground mb-4 text-3xl font-bold">{t("techStack.title")}</h2>
            <p className="text-muted-foreground">{t("techStack.subtitle")}</p>
          </div>

          <div className="grid gap-6 md:grid-cols-3">
            <div className="border-primary/20 bg-primary/5 hover:border-primary/40 rounded-lg border p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-lg">
              <h3 className="text-primary mb-4 text-lg font-semibold">{t("techStack.frontend.title")}</h3>
              <ul className="text-card-foreground space-y-2">
                <li className="flex items-center gap-2">
                  <span className="text-primary">&#9679;</span> {t("techStack.frontend.react")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-primary">&#9679;</span> {t("techStack.frontend.tanstack")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-primary">&#9679;</span> {t("techStack.frontend.tailwind")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-primary">&#9679;</span> {t("techStack.frontend.zustand")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-primary">&#9679;</span> {t("techStack.frontend.websocket")}
                </li>
              </ul>
            </div>
            <div className="border-success/20 bg-success/5 hover:border-success/40 rounded-lg border p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-lg">
              <h3 className="text-success mb-4 text-lg font-semibold">{t("techStack.backend.title")}</h3>
              <ul className="text-card-foreground space-y-2">
                <li className="flex items-center gap-2">
                  <span className="text-success">&#9679;</span> {t("techStack.backend.go")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-success">&#9679;</span> {t("techStack.backend.postgresql")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-success">&#9679;</span> {t("techStack.backend.dragonfly")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-success">&#9679;</span> {t("techStack.backend.multiTenant")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-success">&#9679;</span> {t("techStack.backend.swagger")}
                </li>
              </ul>
            </div>
            <div className="border-warning/20 bg-warning/5 hover:border-warning/40 rounded-lg border p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-lg">
              <h3 className="text-warning mb-4 text-lg font-semibold">{t("techStack.devops.title")}</h3>
              <ul className="text-card-foreground space-y-2">
                <li className="flex items-center gap-2">
                  <span className="text-warning">&#9679;</span> {t("techStack.devops.docker")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-warning">&#9679;</span> {t("techStack.devops.prometheus")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-warning">&#9679;</span> {t("techStack.devops.grafana")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-warning">&#9679;</span> {t("techStack.devops.river")}
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-warning">&#9679;</span> {t("techStack.devops.github")}
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Getting Started */}
      <section className="px-4 py-16">
        <div className="mx-auto max-w-4xl">
          <div className="mb-12 text-center">
            <h2 className="text-foreground mb-4 text-3xl font-bold">{t("gettingStarted.title")}</h2>
            <p className="text-muted-foreground">{t("gettingStarted.subtitle")}</p>
          </div>

          <div className="bg-card rounded-xl p-8 shadow-md">
            <div className="space-y-4">
              <div className="flex items-start gap-4">
                <span className="bg-primary/10 text-primary flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                  1
                </span>
                <div>
                  <p className="text-card-foreground font-medium">{t("gettingStarted.steps.clone.title")}</p>
                  <code className="bg-muted mt-1 block rounded px-3 py-2 text-sm">
                    {t("gettingStarted.steps.clone.command")}
                  </code>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <span className="bg-primary/10 text-primary flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                  2
                </span>
                <div>
                  <p className="text-card-foreground font-medium">{t("gettingStarted.steps.configure.title")}</p>
                  <code className="bg-muted mt-1 block rounded px-3 py-2 text-sm">
                    {t("gettingStarted.steps.configure.command")}
                  </code>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <span className="bg-primary/10 text-primary flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                  3
                </span>
                <div>
                  <p className="text-card-foreground font-medium">{t("gettingStarted.steps.start.title")}</p>
                  <code className="bg-muted mt-1 block rounded px-3 py-2 text-sm">
                    {t("gettingStarted.steps.start.command")}
                  </code>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <span className="bg-success/10 text-success flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold">
                  4
                </span>
                <div>
                  <p className="text-card-foreground font-medium">{t("gettingStarted.steps.build.title")}</p>
                  <p className="text-muted-foreground mt-1 text-sm">{t("gettingStarted.steps.build.subtitle")}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-primary px-4 py-16">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="text-primary-foreground mb-4 text-3xl font-bold">{t("cta.title")}</h2>
          <p className="text-primary-foreground/80 mb-8 text-lg">{t("cta.subtitle")}</p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              to="/demo"
              search={{}}
              className="bg-background text-primary hover:bg-background/90 inline-flex w-full items-center justify-center rounded-lg px-8 py-3 font-semibold shadow-lg transition-all sm:w-auto"
            >
              {t("cta.exploreDemo")}
            </Link>
            <a
              href={`${API_BASE_URL}/swagger/`}
              target="_blank"
              rel="noopener noreferrer"
              className="border-primary-foreground text-primary-foreground hover:bg-primary-foreground/10 inline-flex w-full items-center justify-center rounded-lg border-2 px-8 py-3 font-semibold transition-all sm:w-auto"
            >
              {t("cta.apiDocs")}
            </a>
          </div>
        </div>
      </section>
    </main>
  );
}
