import { createRootRoute, Outlet } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';

import { Toaster } from 'sonner';
import { ThemeProvider } from '@/providers/theme-provider';
import { AuthProvider } from '../providers/AuthContext';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

// HydrateFallback component for better SSR UX
function HydrateFallback() {
  return (
    <div className='min-h-screen flex items-center justify-center bg-background'>
      <div className='text-center space-y-4'>
        <div className='animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto'></div>
        <p className='text-muted-foreground'>Loading...</p>
      </div>
    </div>
  );
}

export const Route = createRootRoute({
  component: () => (
    <>
      <AuthProvider>
        <ThemeProvider defaultTheme='system'>
          {/* QueryClientProvider is handled by the SSR Query integration */}
          <Toaster />
          <Outlet />
          <ReactQueryDevtools initialIsOpen={false} />
        </ThemeProvider>
      </AuthProvider>
      <TanStackRouterDevtools />
    </>
  ),
  notFoundComponent: () => (
    <div className='min-h-screen flex flex-col bg-background'>
      <header className='bg-card text-card-foreground border-b p-4'>
        <h1 className='text-xl font-bold'>Application Error</h1>
      </header>
      <main className='flex-1 p-4'>
        <div className='container mx-auto'>
          <h1 className='text-2xl font-bold text-destructive mb-4'>404</h1>
          <p className='text-muted-foreground mb-4'>
            The requested page could not be found.
          </p>
        </div>
      </main>
    </div>
  ),
  errorComponent: ({ error }: { error: Error }) => {
    let message = 'Oops!';
    let details = 'An unexpected error occurred.';
    let stack: string | undefined;

    if (import.meta.env.DEV && error && error instanceof Error) {
      details = error.message;
      stack = error.stack;
    }

    return (
      <div className='min-h-screen flex flex-col bg-background'>
        <header className='bg-card text-card-foreground border-b p-4'>
          <h1 className='text-xl font-bold'>Application Error</h1>
        </header>
        <main className='flex-1 p-4'>
          <div className='container mx-auto'>
            <h1 className='text-2xl font-bold text-destructive mb-4'>
              {message}
            </h1>
            <p className='text-muted-foreground mb-4'>{details}</p>
            {stack && (
              <pre className='w-full p-4 overflow-x-auto bg-muted rounded'>
                <code className='text-muted-foreground'>{stack}</code>
              </pre>
            )}
          </div>
        </main>
      </div>
    );
  },
});
