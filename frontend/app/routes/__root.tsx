import { ThemeProvider } from "@/providers/theme-provider";
import type { RouterContext } from "@/router";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { createRootRouteWithContext, Outlet, useLocation } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { Toaster } from "sonner";

import { ErrorFallback } from "../components/error";
import { OfflineBanner } from "../components/ui/offline-banner";
import { StandardLayout } from "../layouts";

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

// RootLayout component that applies layouts based on route groups
function RootLayout() {
  const location = useLocation();

  // Check if we're in the layout-demo route
  const isLayoutDemo = location.pathname.startsWith("/layout-demo");

  if (isLayoutDemo) {
    // For layout-demo routes, just render the outlet without StandardLayout
    return <Outlet />;
  }

  // For all other routes, use StandardLayout
  return <StandardLayout />;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => (
    <>
      <ThemeProvider defaultTheme="system">
        {/* QueryClientProvider is handled by the SSR Query integration */}
        <OfflineBanner />
        <Toaster />
        <RootLayout />
        <ReactQueryDevtools initialIsOpen={false} />
      </ThemeProvider>
      <TanStackRouterDevtools />
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
