import { useCallback } from "react";

import { ErrorBoundary } from "@/components/error/ErrorBoundary";
import { requireAdmin } from "@/lib/guards";
import { createFileRoute, Outlet, useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

/**
 * Layout route for all admin routes.
 * Enforces admin role via beforeLoad guard.
 * Redirects to / if user is authenticated but not admin.
 * Redirects to /login if user is not authenticated.
 */
export const Route = createFileRoute("/(app)/admin")({
  beforeLoad: async (ctx) => requireAdmin(ctx),
  component: AdminLayout,
});

function AdminLayout() {
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
          <h2 className="text-destructive mb-2 text-xl font-semibold">{t("adminError.title", "Admin Error")}</h2>
          <p className="text-muted-foreground mb-4">
            {t("adminError.message", "Something went wrong in the admin panel.")}
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
