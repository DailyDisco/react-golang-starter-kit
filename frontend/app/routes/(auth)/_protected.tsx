import {
  createFileRoute,
  Link,
  Outlet,
  redirect,
  useLocation,
} from '@tanstack/react-router';
import { BarChart3, Home, LogOut, Settings, User, Users } from 'lucide-react';

import { Button } from '../../components/ui/button';
import { useAuth } from '../../hooks/auth/useAuth';

export const Route = createFileRoute('/(auth)/_protected')({
  component: ProtectedLayout,
  beforeLoad: async ({ location }: { location: { href: string } }) => {
    // Check authentication status
    // const { isAuthenticated } = useAuth();
    // For demo purposes, we'll simulate authentication check
    const isAuthenticated = localStorage.getItem('isAuthenticated') === 'true';

    if (!isAuthenticated) {
      throw redirect({
        to: '/(auth)/login',
        search: {
          redirect: location.href,
        },
      });
    }
  },
});

function ProtectedLayout() {
  const { logout, user } = useAuth();
  const location = useLocation();

  const dashboardNav = [
    {
      name: 'Dashboard',
      href: '/(auth)/_protected/(dashboard)/users',
      icon: Home,
    },
    { name: 'Users', href: '/(dashboard)/users', icon: Users },
    { name: 'Analytics', href: '/(dashboard)/analytics', icon: BarChart3 },
    { name: 'Settings', href: '/(dashboard)/settings', icon: Settings },
    { name: 'Profile', href: '/profile', icon: User },
  ];

  const isActive = (href: string) => {
    return (
      location.pathname.startsWith(href) ||
      (href === '/(auth)/_protected/(dashboard)/users' &&
        location.pathname === '/')
    );
  };

  return (
    <div className='bg-background flex min-h-screen'>
      {/* Sidebar Navigation */}
      <aside className='bg-card border-border w-64 border-r'>
        <div className='border-border border-b p-6'>
          <div className='flex items-center gap-3'>
            <div className='bg-primary flex h-10 w-10 items-center justify-center rounded-lg'>
              <span className='text-primary-foreground text-lg font-bold'>
                RG
              </span>
            </div>
            <div>
              <h2 className='text-lg font-semibold'>Dashboard</h2>
              <p className='text-muted-foreground text-sm'>Welcome back!</p>
            </div>
          </div>
        </div>

        <nav className='p-4'>
          <ul className='space-y-2'>
            {dashboardNav.map(item => {
              const Icon = item.icon;
              return (
                <li key={item.name}>
                  <Link
                    to={item.href}
                    search={{}}
                    className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                      isActive(item.href)
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                    }`}
                  >
                    <Icon className='h-4 w-4' />
                    {item.name}
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>

        {/* User Info & Logout */}
        <div className='border-border bg-card/50 absolute right-0 bottom-0 left-0 border-t p-4'>
          <div className='mb-3 flex items-center gap-3'>
            <div className='bg-muted flex h-8 w-8 items-center justify-center rounded-full'>
              <span className='text-xs font-medium'>
                {user?.name
                  ?.split(' ')
                  .map(n => n[0])
                  .join('')
                  .toUpperCase() || 'U'}
              </span>
            </div>
            <div className='min-w-0 flex-1'>
              <p className='truncate text-sm font-medium'>
                {user?.name || 'User'}
              </p>
              <p className='text-muted-foreground truncate text-xs'>
                {user?.email || ''}
              </p>
            </div>
          </div>
          <Button
            variant='outline'
            size='sm'
            onClick={logout}
            className='w-full justify-start gap-2'
          >
            <LogOut className='h-4 w-4' />
            Sign Out
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <div className='flex flex-1 flex-col'>
        <div className='bg-muted/30 border-b'>
          <div className='p-6'>
            <h1 className='text-foreground text-2xl font-semibold'>
              Dashboard
            </h1>
          </div>
        </div>
        <main className='flex-1 overflow-auto p-6'>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
