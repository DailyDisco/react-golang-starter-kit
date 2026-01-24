import { useCallback } from "react";

import { ErrorBoundary } from "@/components/error/ErrorBoundary";
import { requireAuth } from "@/lib/guards";
import { OrganizationService } from "@/services/organizations/organizationService";
import { createFileRoute, Outlet, redirect, useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

/**
 * Layout route for organization-specific routes.
 * Enforces authentication and validates org membership.
 * Redirects to /403 if user doesn't have access to the organization.
 */
export const Route = createFileRoute("/(app)/org/$orgSlug")({
  beforeLoad: async (ctx) => {
    // First ensure user is authenticated
    const { user } = await requireAuth(ctx);

    // Get the org slug from params
    const orgSlug = ctx.params.orgSlug;

    try {
      // Verify user has access to this organization
      const org = await OrganizationService.getOrganization(orgSlug);

      // Return org context for child routes
      return { user, organization: org };
    } catch (error) {
      // User doesn't have access to this org
      throw redirect({
        to: "/403",
        search: { from: ctx.location.pathname },
      });
    }
  },
  component: OrgLayout,
});

function OrgLayout() {
  const router = useRouter();
  const { t } = useTranslation("errors");

  const handleReset = useCallback(() => {
    router.invalidate();
  }, [router]);

  return (
    <ErrorBoundary
      onReset={handleReset}
      fallback={
        <div className="flex min-h-[400px] flex-col items-center justify-center p-8">
          <h2 className="text-destructive mb-2 text-xl font-semibold">{t("orgError.title", "Organization Error")}</h2>
          <p className="text-muted-foreground mb-4">
            {t("orgError.message", "Something went wrong loading this organization.")}
          </p>
          <button
            onClick={handleReset}
            className="bg-primary text-primary-foreground hover:bg-primary/90 rounded px-4 py-2"
          >
            {t("actions.retry", "Try Again")}
          </button>
        </div>
      }
    >
      <Outlet />
    </ErrorBoundary>
  );
}
