import { lazy, Suspense, useEffect, useState } from "react";

import { ThemeProvider } from "@/providers/theme-provider";
import type { RouterContext } from "@/router";
import { createRootRouteWithContext, Outlet, useLocation, useNavigate } from "@tanstack/react-router";
import { Toaster } from "sonner";

import { SessionExpiredModal } from "../components/auth/SessionExpiredModal";
import { CommandPalette } from "../components/command-palette";
import { ErrorFallback } from "../components/error";
import { OfflineBanner } from "../components/ui/offline-banner";
import { useLanguageSync } from "../hooks/useLanguageSync";
import { StandardLayout } from "../layouts";
import { initCSRFToken } from "../services/api/client";

// Lazy load devtools to exclude from production bundle
const ReactQueryDevtools = import.meta.env.DEV
  ? lazy(() =>
      import("@tanstack/react-query-devtools").then((m) => ({
        default: m.ReactQueryDevtools,
      }))
    )
  : () => null;

const TanStackRouterDevtools = import.meta.env.DEV
  ? lazy(() =>
      import("@tanstack/react-router-devtools").then((m) => ({
        default: m.TanStackRouterDevtools,
      }))
    )
  : () => null;

/**
 * Initialize CSRF token on app load.
 * This ensures the CSRF cookie is set before any state-changing requests.
 */
function CSRFInitializer() {
  useEffect(() => {
    // Initialize CSRF token on app mount
    // This ensures we have a valid CSRF token before making POST/PUT/DELETE requests
    void initCSRFToken();
  }, []);

  return null;
}

/**
 * Sync language preference between frontend and backend.
 * On login: syncs language from backend to frontend.
 * On language change: updates backend.
 */
function LanguageSyncInitializer() {
  useLanguageSync();
  return null;
}

// HydrateFallback component for better SSR UX
function HydrateFallback() {
  return (
    <div className="bg-background flex min-h-screen items-center justify-center">
      <div className="space-y-4 text-center">
        <div className="border-primary mx-auto h-8 w-8 animate-spin rounded-full border-b-2"></div>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    </div>
  );
}

// Auth guard component for protected routes
function AuthGuard({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    // Check for valid user data in localStorage
    const storedUser = localStorage.getItem("auth_user");
    let authenticated = false;

    if (storedUser) {
      try {
        JSON.parse(storedUser);
        authenticated = true;
      } catch {
        localStorage.removeItem("auth_user");
      }
    }

    if (!authenticated) {
      navigate({ to: "/login" });
    } else {
      setIsAuthenticated(true);
    }
    setIsChecking(false);
  }, [navigate]);

  if (isChecking) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <div className="space-y-4 text-center">
          <div className="border-primary mx-auto h-8 w-8 animate-spin rounded-full border-b-2"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}

// RootLayout component that applies layouts based on route groups
function RootLayout() {
  const location = useLocation();

  // Routes that use the AppLayout (sidebar layout for authenticated users)
  const appRoutes = ["/dashboard", "/settings", "/admin", "/billing"];
  const isAppRoute = appRoutes.some((route) => location.pathname.startsWith(route));

  // Check if we're in the layout-demo route
  const isLayoutDemo = location.pathname.startsWith("/layout-demo");

  // Layout-demo routes don't use any wrapper layout
  if (isLayoutDemo) {
    return <Outlet />;
  }

  // App routes use StandardLayout with auth guard
  if (isAppRoute) {
    return (
      <AuthGuard>
        <StandardLayout />
      </AuthGuard>
    );
  }

  // Public routes use StandardLayout (navbar + footer)
  return <StandardLayout />;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => (
    <>
      <ThemeProvider defaultTheme="system">
        {/* QueryClientProvider is handled by the SSR Query integration */}
        <CSRFInitializer />
        <LanguageSyncInitializer />
        <OfflineBanner />
        <Toaster />
        <SessionExpiredModal />
        <CommandPalette />
        <RootLayout />
        {import.meta.env.DEV && (
          <Suspense fallback={null}>
            <ReactQueryDevtools initialIsOpen={false} />
          </Suspense>
        )}
      </ThemeProvider>
      {import.meta.env.DEV && (
        <Suspense fallback={null}>
          <TanStackRouterDevtools />
        </Suspense>
      )}
    </>
  ),
  notFoundComponent: () => (
    <div className="bg-background flex min-h-screen flex-col">
      <header className="bg-card text-card-foreground border-b p-4">
        <h1 className="text-xl font-bold">Application Error</h1>
      </header>
      <main className="flex-1 p-4">
        <div className="container mx-auto">
          <h1 className="text-destructive mb-4 text-2xl font-bold">404</h1>
          <p className="text-muted-foreground mb-4">The requested page could not be found.</p>
        </div>
      </main>
    </div>
  ),
  errorComponent: ({ error, reset }: { error: Error; reset?: () => void }) => (
    <div className="bg-background flex min-h-screen flex-col">
      <header className="bg-card text-card-foreground border-b p-4">
        <h1 className="text-xl font-bold">Application Error</h1>
      </header>
      <main className="flex-1 p-4">
        <ErrorFallback
          error={error}
          resetError={reset}
          showStack={import.meta.env.DEV}
        />
      </main>
    </div>
  ),
});
