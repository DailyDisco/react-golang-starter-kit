import { QueryClient } from '@tanstack/react-query';
import { createRouter } from '@tanstack/react-router';
import { setupRouterSsrQueryIntegration } from '@tanstack/react-router-ssr-query';
import { routeTree } from './routeTree.gen';
import { queryClient } from './lib/query-client';

// Router types are registered in router.types.ts

export function createAppRouter() {
    const router = createRouter({
        routeTree,
        // Expose the QueryClient via router context for use in loaders
        context: {
            queryClient,
        } as any, // Cast to any to avoid type issues
        scrollRestoration: true,
        defaultPreload: 'intent',
        // Add default error boundary and loading components
        defaultErrorComponent: ({ error }: { error: Error }) => (
            <div className="min-h-screen flex items-center justify-center">
                <div className="text-center">
                    <h2 className="text-2xl font-bold text-red-600 mb-4">Something went wrong!</h2>
                    <p className="text-gray-600">{error.message}</p>
                </div>
            </div>
        ),
        defaultPendingComponent: () => (
            <div className="min-h-screen flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
        ),
    } as any); // Cast to any to avoid router type issues

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
