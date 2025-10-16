import { QueryClient } from "@tanstack/react-query";
import { createRouter } from "@tanstack/react-router";
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query";

import { queryClient } from "./lib/query-client";
import { routeTree } from "./routeTree.gen";

// Define the router context shape for type safety
// This interface is used by createRootRouteWithContext in __root.tsx
export interface RouterContext {
  queryClient: QueryClient;
}

export function createAppRouter() {
  const router = createRouter({
    routeTree,
    // Expose the QueryClient via router context for use in loaders
    context: {
      queryClient,
    },
    scrollRestoration: true,
    defaultPreload: "intent",
    // Add default error boundary and loading components
    defaultErrorComponent: ({ error }: { error: Error }) => (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="mb-4 text-2xl font-bold text-red-600">Something went wrong!</h2>
          <p className="text-gray-600">{error.message}</p>
        </div>
      </div>
    ),
    defaultPendingComponent: () => (
      <div className="flex min-h-screen items-center justify-center">
        <div className="border-primary h-8 w-8 animate-spin rounded-full border-b-2"></div>
      </div>
    ),
  });

  setupRouterSsrQueryIntegration({
    router,
    queryClient,
    // Let the integration handle QueryClientProvider wrapping
    wrapQueryClient: true,
    // Handle redirects from queries/mutations
    handleRedirects: true,
  });

  return router;
}

// Declare module augmentation for TanStack Router
// This registers the router type globally so hooks like useNavigate, Link, etc. are fully typed
declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createAppRouter>;
  }
}
