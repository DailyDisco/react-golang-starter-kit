import { useCallback } from "react";

import { ErrorBoundary } from "@/components/error/ErrorBoundary";
import { requireAuth } from "@/lib/guards";
import { createFileRoute, Outlet, useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

/**
 * Layout route for all authenticated app routes.
 * Enforces authentication via beforeLoad guard - redirects to /login if unauthenticated.
 * Child routes (dashboard, billing, settings, org, admin) inherit this protection.
 */
export const Route = createFileRoute("/(app)")({
  beforeLoad: async (ctx) => requireAuth(ctx),
  component: AppLayout,
});

function AppLayout() {
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
          <h2 className="text-destructive mb-2 text-xl font-semibold">{t("appError.title", "Something went wrong")}</h2>
          <p className="text-muted-foreground mb-4">
            {t("appError.message", "An unexpected error occurred. Please try again.")}
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
