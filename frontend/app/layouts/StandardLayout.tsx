import { Link, Outlet, useLocation } from '@tanstack/react-router';
import { Navbar } from './Navbar';
import { Footer } from './Footer';
import {
  Breadcrumb,
  BreadcrumbLink,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import React from 'react';

// User-friendly route labels mapping
const ROUTE_LABELS: Record<string, string> = {
  '/': 'Home',
  '/about': 'About',
  '/blog': 'Blog',
  '/demo': 'Demo',
  '/search': 'Search',
  '/settings': 'Settings',
  '/profile': 'Profile',
  '/analytics': 'Analytics',
  '/analytics/overview': 'Overview',
  '/users': 'Users',
  '/layout-demo': 'Layout Demo',
  '/api-docs': 'API Docs',
  '/dashboard': 'Dashboard',
  '/login': 'Login',
  '/register': 'Register',
};

// Helper function to get user-friendly label
const getRouteLabel = (pathname: string): string => {
  // Check for exact matches first
  if (ROUTE_LABELS[pathname]) {
    return ROUTE_LABELS[pathname];
  }

  // Handle dynamic routes (e.g., /users/123 -> Users)
  const segments = pathname.split('/').filter(Boolean);
  if (segments.length > 0) {
    const lastSegment = segments[segments.length - 1];

    // If it's a dynamic segment (contains numbers), use the parent route
    if (/^\d+$/.test(lastSegment) && segments.length > 1) {
      const parentPath = '/' + segments.slice(0, -1).join('/');
      return ROUTE_LABELS[parentPath] || lastSegment;
    }

    // Try to find a matching route
    return (
      ROUTE_LABELS['/' + lastSegment] ||
      lastSegment.charAt(0).toUpperCase() + lastSegment.slice(1)
    );
  }

  return 'Home';
};

export default function StandardLayout() {
  const location = useLocation();
  const isOnHomePage = location.pathname === '/';

  // Generate breadcrumbs from current path
  const generateBreadcrumbs = () => {
    if (isOnHomePage) {
      return [];
    }

    const pathSegments = location.pathname.split('/').filter(Boolean);
    const breadcrumbs = [];

    // Build cumulative paths
    for (let i = 0; i < pathSegments.length; i++) {
      const path = '/' + pathSegments.slice(0, i + 1).join('/');
      const label = getRouteLabel(path);

      breadcrumbs.push({
        label,
        to: path,
      });
    }

    return breadcrumbs;
  };

  const breadcrumbs = generateBreadcrumbs();

  return (
    <div className='flex min-h-screen flex-col'>
      <Navbar />
      <div className='bg-muted/30 border-b'>
        <div className='mx-auto max-w-7xl px-4 py-3 sm:px-6 lg:px-8'>
          <Breadcrumb>
            <BreadcrumbList>
              {/* Always show Home */}
              <BreadcrumbItem>
                {isOnHomePage ? (
                  <BreadcrumbPage>Home</BreadcrumbPage>
                ) : (
                  <BreadcrumbLink asChild>
                    <Link to='/'>Home</Link>
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>

              {/* Show breadcrumbs for non-home pages */}
              {!isOnHomePage &&
                breadcrumbs.map((crumb, index) => (
                  <React.Fragment key={crumb.to}>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      {index === breadcrumbs.length - 1 ? (
                        <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                      ) : (
                        <BreadcrumbLink asChild>
                          <Link to={crumb.to}>{crumb.label}</Link>
                        </BreadcrumbLink>
                      )}
                    </BreadcrumbItem>
                  </React.Fragment>
                ))}
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </div>
      <main className='flex-1'>
        <Outlet />
      </main>
      <Footer />
    </div>
  );
}
