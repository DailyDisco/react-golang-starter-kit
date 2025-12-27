import { createFileRoute, Link } from "@tanstack/react-router";
import {
  Bell,
  CreditCard,
  FileText,
  Flag,
  HardDrive,
  Key,
  LayoutDashboard,
  Lock,
  Mail,
  MonitorCheck,
  ScrollText,
  Settings,
  Shield,
  ShieldCheck,
  Smartphone,
  Users,
} from "lucide-react";

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

const featureCategories: FeatureCategory[] = [
  {
    title: "Authentication & Security",
    icon: <Shield className="h-6 w-6" />,
    iconColor: "text-blue-600 dark:text-blue-400",
    features: [
      { icon: <Key className="h-4 w-4" />, name: "Email/Password Auth", description: "JWT-based authentication" },
      { icon: <Users className="h-4 w-4" />, name: "OAuth Integration", description: "Google & GitHub SSO" },
      { icon: <Smartphone className="h-4 w-4" />, name: "Two-Factor Auth", description: "TOTP with backup codes" },
      { icon: <MonitorCheck className="h-4 w-4" />, name: "Session Management", description: "View & revoke sessions" },
      { icon: <ScrollText className="h-4 w-4" />, name: "Login History", description: "Track login activity" },
      { icon: <ShieldCheck className="h-4 w-4" />, name: "IP Blocklist", description: "Block malicious IPs" },
    ],
  },
  {
    title: "Billing & Monetization",
    icon: <CreditCard className="h-6 w-6" />,
    iconColor: "text-green-600 dark:text-green-400",
    features: [
      { icon: <CreditCard className="h-4 w-4" />, name: "Stripe Integration", description: "Complete billing system" },
      {
        icon: <FileText className="h-4 w-4" />,
        name: "Subscription Plans",
        description: "Dynamic pricing from Stripe",
      },
      { icon: <Settings className="h-4 w-4" />, name: "Customer Portal", description: "Self-service management" },
    ],
  },
  {
    title: "Admin & Management",
    icon: <LayoutDashboard className="h-6 w-6" />,
    iconColor: "text-purple-600 dark:text-purple-400",
    features: [
      { icon: <LayoutDashboard className="h-4 w-4" />, name: "Admin Dashboard", description: "Stats & analytics" },
      { icon: <Users className="h-4 w-4" />, name: "User Management", description: "Impersonate & manage users" },
      { icon: <Flag className="h-4 w-4" />, name: "Feature Flags", description: "Rollout control & targeting" },
      { icon: <ScrollText className="h-4 w-4" />, name: "Audit Logging", description: "Track all actions" },
      { icon: <Bell className="h-4 w-4" />, name: "Announcements", description: "Site-wide banners" },
      { icon: <Mail className="h-4 w-4" />, name: "Email Templates", description: "Customizable emails" },
    ],
  },
  {
    title: "Infrastructure",
    icon: <HardDrive className="h-6 w-6" />,
    iconColor: "text-orange-600 dark:text-orange-400",
    features: [
      { icon: <HardDrive className="h-4 w-4" />, name: "File Storage", description: "S3 or database storage" },
      { icon: <MonitorCheck className="h-4 w-4" />, name: "Health Monitoring", description: "Real-time system status" },
      { icon: <Lock className="h-4 w-4" />, name: "Rate Limiting", description: "Built-in protection" },
      { icon: <Settings className="h-4 w-4" />, name: "User Preferences", description: "Theme, timezone, language" },
    ],
  },
];

function Home() {
  return (
    <main className="bg-gray-50 dark:bg-gray-900">
      {/* Hero Section */}
      <section className="px-4 py-16 md:py-24">
        <div className="mx-auto max-w-4xl text-center">
          <h1 className="mb-6 text-4xl font-bold tracking-tight text-gray-900 md:text-5xl lg:text-6xl dark:text-white">
            Production-Ready
            <span className="block text-blue-600 dark:text-blue-400">React + Go Starter Kit</span>
          </h1>
          <p className="mx-auto mb-8 max-w-2xl text-lg text-gray-600 md:text-xl dark:text-gray-300">
            Launch your SaaS faster with authentication, billing, admin panel, and more built-in. Everything you need to
            go from idea to production.
          </p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              to="/demo"
              search={{}}
              className="inline-flex w-full items-center justify-center rounded-lg bg-blue-600 px-8 py-3 font-semibold text-white shadow-lg transition-all hover:bg-blue-700 hover:shadow-xl sm:w-auto dark:bg-blue-700 dark:hover:bg-blue-600"
            >
              Try Interactive Demo
            </Link>
            <Link
              to="/pricing"
              search={{}}
              className="inline-flex w-full items-center justify-center rounded-lg border border-gray-300 bg-white px-8 py-3 font-semibold text-gray-700 shadow-sm transition-all hover:bg-gray-50 sm:w-auto dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
            >
              View Pricing
            </Link>
          </div>
        </div>
      </section>

      {/* Feature Categories */}
      <section className="px-4 py-16">
        <div className="mx-auto max-w-6xl">
          <div className="mb-12 text-center">
            <h2 className="mb-4 text-3xl font-bold text-gray-900 dark:text-white">Everything You Need</h2>
            <p className="mx-auto max-w-2xl text-gray-600 dark:text-gray-400">
              Enterprise-grade features built in from day one. Focus on your product, not boilerplate.
            </p>
          </div>

          <div className="grid gap-8 md:grid-cols-2">
            {featureCategories.map((category) => (
              <div
                key={category.title}
                className="rounded-xl bg-white p-6 shadow-md transition-shadow hover:shadow-lg dark:bg-gray-800"
              >
                <div className="mb-4 flex items-center gap-3">
                  <div className={category.iconColor}>{category.icon}</div>
                  <h3 className="text-xl font-semibold text-gray-900 dark:text-white">{category.title}</h3>
                </div>
                <div className="grid gap-3 sm:grid-cols-2">
                  {category.features.map((feature) => (
                    <div
                      key={feature.name}
                      className="flex items-start gap-2"
                    >
                      <div className="mt-0.5 text-gray-400">{feature.icon}</div>
                      <div>
                        <p className="font-medium text-gray-900 dark:text-white">{feature.name}</p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">{feature.description}</p>
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
      <section className="bg-white px-4 py-16 dark:bg-gray-800">
        <div className="mx-auto max-w-4xl">
          <div className="mb-12 text-center">
            <h2 className="mb-4 text-3xl font-bold text-gray-900 dark:text-white">Modern Tech Stack</h2>
            <p className="text-gray-600 dark:text-gray-400">
              Built with the latest technologies for performance and developer experience.
            </p>
          </div>

          <div className="grid gap-8 md:grid-cols-2">
            <div className="rounded-lg border border-blue-100 bg-blue-50/50 p-6 dark:border-blue-900 dark:bg-blue-950/30">
              <h3 className="mb-4 text-lg font-semibold text-blue-600 dark:text-blue-400">Frontend</h3>
              <ul className="space-y-2 text-gray-700 dark:text-gray-300">
                <li className="flex items-center gap-2">
                  <span className="text-blue-500">&#9679;</span> React 19 with TypeScript
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-blue-500">&#9679;</span> TanStack Router & Query
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-blue-500">&#9679;</span> TailwindCSS + ShadCN/UI
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-blue-500">&#9679;</span> Vite for fast builds
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-blue-500">&#9679;</span> Zustand state management
                </li>
              </ul>
            </div>
            <div className="rounded-lg border border-green-100 bg-green-50/50 p-6 dark:border-green-900 dark:bg-green-950/30">
              <h3 className="mb-4 text-lg font-semibold text-green-600 dark:text-green-400">Backend</h3>
              <ul className="space-y-2 text-gray-700 dark:text-gray-300">
                <li className="flex items-center gap-2">
                  <span className="text-green-500">&#9679;</span> Go 1.25 with Chi router
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-500">&#9679;</span> PostgreSQL + GORM
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-500">&#9679;</span> Redis caching
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-500">&#9679;</span> Docker & Docker Compose
                </li>
                <li className="flex items-center gap-2">
                  <span className="text-green-500">&#9679;</span> Swagger API docs
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
            <h2 className="mb-4 text-3xl font-bold text-gray-900 dark:text-white">Get Started in Minutes</h2>
            <p className="text-gray-600 dark:text-gray-400">Clone, configure, and start building your next project.</p>
          </div>

          <div className="rounded-xl bg-white p-8 shadow-md dark:bg-gray-800">
            <div className="space-y-4">
              <div className="flex items-start gap-4">
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-semibold text-blue-600 dark:bg-blue-900 dark:text-blue-400">
                  1
                </span>
                <div>
                  <p className="font-medium text-gray-900 dark:text-white">Clone the repository</p>
                  <code className="mt-1 block rounded bg-gray-100 px-3 py-2 text-sm dark:bg-gray-700">
                    git clone https://github.com/your-repo/react-golang-starter-kit.git
                  </code>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-semibold text-blue-600 dark:bg-blue-900 dark:text-blue-400">
                  2
                </span>
                <div>
                  <p className="font-medium text-gray-900 dark:text-white">Configure environment</p>
                  <code className="mt-1 block rounded bg-gray-100 px-3 py-2 text-sm dark:bg-gray-700">
                    cp .env.example .env && nano .env
                  </code>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 text-sm font-semibold text-blue-600 dark:bg-blue-900 dark:text-blue-400">
                  3
                </span>
                <div>
                  <p className="font-medium text-gray-900 dark:text-white">Start everything with Docker</p>
                  <code className="mt-1 block rounded bg-gray-100 px-3 py-2 text-sm dark:bg-gray-700">
                    docker-compose up
                  </code>
                </div>
              </div>
              <div className="flex items-start gap-4">
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-green-100 text-sm font-semibold text-green-600 dark:bg-green-900 dark:text-green-400">
                  4
                </span>
                <div>
                  <p className="font-medium text-gray-900 dark:text-white">Start building!</p>
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Frontend at localhost:5173, Backend at localhost:8080, API docs at /swagger
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-blue-600 px-4 py-16 dark:bg-blue-900">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="mb-4 text-3xl font-bold text-white">Ready to Build Something Great?</h2>
          <p className="mb-8 text-lg text-blue-100">
            Explore the demo to see all features in action, or dive into the documentation.
          </p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              to="/demo"
              search={{}}
              className="inline-flex w-full items-center justify-center rounded-lg bg-white px-8 py-3 font-semibold text-blue-600 shadow-lg transition-all hover:bg-gray-100 sm:w-auto"
            >
              Explore Demo
            </Link>
            <a
              href={`${API_BASE_URL}/swagger/`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex w-full items-center justify-center rounded-lg border-2 border-white px-8 py-3 font-semibold text-white transition-all hover:bg-white/10 sm:w-auto"
            >
              API Documentation
            </a>
          </div>
        </div>
      </section>
    </main>
  );
}
