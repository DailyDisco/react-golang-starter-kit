import React from "react";

import { AnnouncementsContainer } from "@/components/announcements";
import { UpgradePrompt } from "@/components/billing";
import { ErrorBoundary } from "@/components/error/ErrorBoundary";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Link, Outlet, useLocation, useRouter } from "@tanstack/react-router";
import { Home } from "lucide-react";

import { logger } from "../lib/logger";
import { Footer } from "./Footer";
import { Navbar } from "./Navbar";

// User-friendly route labels mapping
const ROUTE_LABELS: Record<string, string> = {
  "/": "Home",
  "/about": "About",
  "/blog": "Blog",
  "/demo": "Demo",
  "/search": "Search",
  "/settings": "Settings",
  "/profile": "Profile",
  "/analytics": "Analytics",
  "/analytics/overview": "Overview",
  "/users": "Users",
  "/layout-demo": "Layout Demo",
  "/api-docs": "API Docs",
  "/dashboard": "Dashboard",
  "/login": "Login",
  "/register": "Register",
  "/changelog": "Changelog",
  // App routes
  "/billing": "Billing",
  "/admin": "Admin",
  "/admin/users": "Users",
  "/admin/audit-logs": "Audit Logs",
  "/admin/feature-flags": "Feature Flags",
  "/admin/health": "System Health",
  "/admin/announcements": "Announcements",
  "/admin/email-templates": "Email Templates",
  "/admin/settings": "Admin Settings",
  // Settings sub-routes
  "/settings/profile": "Profile",
  "/settings/security": "Security",
  "/settings/preferences": "Preferences",
  "/settings/notifications": "Notifications",
  "/settings/privacy": "Privacy",
  "/settings/login-history": "Login History",
  "/settings/connected-accounts": "Connected Accounts",
};

// Helper function to get user-friendly label
const getRouteLabel = (pathname: string): string => {
  // Check for exact matches first
  if (ROUTE_LABELS[pathname]) {
    return ROUTE_LABELS[pathname];
  }

  // Handle dynamic routes (e.g., /users/123 -> Users)
  const segments = pathname.split("/").filter(Boolean);
  if (segments.length > 0) {
    const lastSegment = segments[segments.length - 1];

    // If it's a dynamic segment (contains numbers), use the parent route
    if (/^\d+$/.test(lastSegment) && segments.length > 1) {
      const parentPath = "/" + segments.slice(0, -1).join("/");
      return ROUTE_LABELS[parentPath] || lastSegment;
    }

    // Try to find a matching route
    return ROUTE_LABELS["/" + lastSegment] || lastSegment.charAt(0).toUpperCase() + lastSegment.slice(1);
  }

  return "Home";
};

export default function StandardLayout() {
  const location = useLocation();
  const router = useRouter();
  const isOnHomePage = location.pathname === "/";

  // Handle error boundary reset by invalidating router
  const handleErrorReset = () => {
    router.invalidate();
  };

  // Log errors to our structured logger
  const handleError = (error: Error) => {
    logger.error("Page error caught by ErrorBoundary", error, {
      pathname: location.pathname,
    });
  };

  // Generate breadcrumbs from current path
  const generateBreadcrumbs = () => {
    if (isOnHomePage) {
      return [];
    }

    const pathSegments = location.pathname.split("/").filter(Boolean);
    const breadcrumbs = [];

    // Build cumulative paths
    for (let i = 0; i < pathSegments.length; i++) {
      const path = "/" + pathSegments.slice(0, i + 1).join("/");
      const label = getRouteLabel(path);

      breadcrumbs.push({
        label,
        to: path,
      });
    }

    return breadcrumbs;
  };

  const breadcrumbs = generateBreadcrumbs();

  return (
    <div className="flex min-h-screen flex-col">
      {/* Skip to main content link for accessibility */}
      <a
        href="#main-content"
        className="focus:bg-primary focus:text-primary-foreground focus:ring-ring sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:rounded-md focus:px-4 focus:py-2 focus:ring-2 focus:outline-none"
      >
        Skip to main content
      </a>
      <Navbar />
      <AnnouncementsContainer />
      <UpgradePrompt className="mx-auto max-w-screen-2xl px-4 pt-4 sm:px-6 lg:px-8" />
      <div className="bg-background/95 supports-backdrop-filter:bg-background/60 border-b backdrop-blur">
        <div className="mx-auto max-w-screen-2xl px-4 py-2 sm:px-6 lg:px-8">
          <Breadcrumb>
            <BreadcrumbList>
              {/* Always show Home */}
              <BreadcrumbItem>
                {isOnHomePage ? (
                  <BreadcrumbPage>
                    <Home className="size-3.5" />
                    Home
                  </BreadcrumbPage>
                ) : (
                  <BreadcrumbLink asChild>
                    <Link to="/">
                      <Home className="size-3.5" />
                      Home
                    </Link>
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>

              {/* Show breadcrumbs for non-home pages */}
              {!isOnHomePage &&
                breadcrumbs.map((crumb, index) => (
                  <React.Fragment key={crumb.to}>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      {index === breadcrumbs.length - 1 ? (
                        <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                      ) : (
                        <BreadcrumbLink asChild>
                          <Link to={crumb.to}>{crumb.label}</Link>
                        </BreadcrumbLink>
                      )}
                    </BreadcrumbItem>
                  </React.Fragment>
                ))}
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </div>
      <main
        id="main-content"
        className="flex-1"
      >
        <div className="mx-auto max-w-screen-2xl px-4 py-6 sm:px-6 lg:px-8">
          <ErrorBoundary
            onError={handleError}
            onReset={handleErrorReset}
            showStack={import.meta.env.DEV}
          >
            <Outlet />
          </ErrorBoundary>
        </div>
      </main>
      <Footer />
    </div>
  );
}
