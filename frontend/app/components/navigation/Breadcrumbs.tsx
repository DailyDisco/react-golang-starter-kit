import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Link, useLocation } from "@tanstack/react-router";
import { ChevronRight, Home } from "lucide-react";
import { useTranslation } from "react-i18next";

// Map of route paths to translation keys
const routeTranslationKeys: Record<string, string> = {
  dashboard: "navigation.dashboard",
  billing: "navigation.billing",
  settings: "navigation.settings",
  profile: "navigation.profile",
  security: "navigation.security",
  preferences: "navigation.preferences",
  notifications: "navigation.notifications",
  privacy: "navigation.privacy",
  "login-history": "navigation.loginHistory",
  "connected-accounts": "navigation.connectedAccounts",
  admin: "navigation.admin",
  users: "navigation.users",
  "audit-logs": "navigation.auditLogs",
  "feature-flags": "navigation.featureFlags",
  health: "navigation.systemHealth",
  announcements: "navigation.announcements",
  "email-templates": "navigation.emailTemplates",
};

export function Breadcrumbs() {
  const { t } = useTranslation("common");
  const location = useLocation();

  // Split the pathname into segments
  const pathSegments = location.pathname.split("/").filter(Boolean);

  // Don't show breadcrumbs for root or single-segment paths
  if (pathSegments.length === 0) {
    return null;
  }

  // Build breadcrumb items
  const breadcrumbItems = pathSegments.map((segment, index) => {
    const path = "/" + pathSegments.slice(0, index + 1).join("/");
    const translationKey = routeTranslationKeys[segment];

    const label = translationKey
      ? (t(translationKey as any) as string)
      : segment.charAt(0).toUpperCase() + segment.slice(1);
    const isLast = index === pathSegments.length - 1;

    return {
      path,
      label,
      isLast,
    };
  });

  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink asChild>
            <Link
              to="/dashboard"
              search={{}}
              className="flex items-center gap-1"
            >
              <Home className="h-3.5 w-3.5" />
              <span className="sr-only">{t("navigation.home")}</span>
            </Link>
          </BreadcrumbLink>
        </BreadcrumbItem>

        {breadcrumbItems.map((item) => (
          <BreadcrumbItem key={item.path}>
            <BreadcrumbSeparator>
              <ChevronRight className="h-3.5 w-3.5" />
            </BreadcrumbSeparator>
            {item.isLast ? (
              <BreadcrumbPage>{item.label}</BreadcrumbPage>
            ) : (
              <BreadcrumbLink asChild>
                <Link
                  to={item.path}
                  search={{}}
                >
                  {item.label}
                </Link>
              </BreadcrumbLink>
            )}
          </BreadcrumbItem>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
