import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from 'react-router';

import type { Route } from './+types/root';
import './app.css';

export const links: Route.LinksFunction = () => [
  { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
  {
    rel: 'preconnect',
    href: 'https://fonts.gstatic.com',
    crossOrigin: 'anonymous',
  },
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap',
  },
];

import { Toaster } from 'sonner';
import { ThemeProvider } from '@/providers/theme-provider';

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang='en'>
      <head>
        <meta charSet='utf-8' />
        <meta name='viewport' content='width=device-width, initial-scale=1' />
        <Meta />
        <Links />
      </head>
      <body>
        <ThemeProvider
          attribute='class'
          defaultTheme='system'
          enableSystem
          disableTransitionOnChange
        >
          <Toaster />
          {children}
        </ThemeProvider>
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  // Layout logic is now handled by the routes configuration
  // Each route group has its own layout component
  // This component only handles the HTML structure and global providers
  return <Outlet />;
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let message = 'Oops!';
  let details = 'An unexpected error occurred.';
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? '404' : 'Error';
    details =
      error.status === 404
        ? 'The requested page could not be found.'
        : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  // Error boundary should show a simple layout without complex dependencies
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
}
