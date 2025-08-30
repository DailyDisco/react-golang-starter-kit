import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import { useAuth } from '../../providers/AuthContext';
import { Button } from '../../components/ui/button';
import { Link, useLocation } from '@tanstack/react-router';
import { Users, BarChart3, Settings, Home, User, LogOut } from 'lucide-react';
import { Breadcrumbs } from '../../components/ui/breadcrumbs';

export const Route = createFileRoute('/(auth)/_protected')({
  component: ProtectedLayout,
  beforeLoad: async ({ location }) => {
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
    <div className='min-h-screen bg-background flex'>
      {/* Sidebar Navigation */}
      <aside className='w-64 bg-card border-r border-border'>
        <div className='p-6 border-b border-border'>
          <div className='flex items-center gap-3'>
            <div className='w-10 h-10 bg-primary rounded-lg flex items-center justify-center'>
              <span className='text-primary-foreground font-bold text-lg'>
                RG
              </span>
            </div>
            <div>
              <h2 className='font-semibold text-lg'>Dashboard</h2>
              <p className='text-sm text-muted-foreground'>Welcome back!</p>
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
                    className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                      isActive(item.href)
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                    }`}
                  >
                    <Icon className='w-4 h-4' />
                    {item.name}
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>

        {/* User Info & Logout */}
        <div className='absolute bottom-0 left-0 right-0 p-4 border-t border-border bg-card/50'>
          <div className='flex items-center gap-3 mb-3'>
            <div className='w-8 h-8 bg-muted rounded-full flex items-center justify-center'>
              <span className='text-xs font-medium'>
                {user?.name
                  ?.split(' ')
                  .map(n => n[0])
                  .join('')
                  .toUpperCase() || 'U'}
              </span>
            </div>
            <div className='flex-1 min-w-0'>
              <p className='text-sm font-medium truncate'>
                {user?.name || 'User'}
              </p>
              <p className='text-xs text-muted-foreground truncate'>
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
            <LogOut className='w-4 h-4' />
            Sign Out
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <div className='flex-1 flex flex-col'>
        <div className='border-b bg-muted/30'>
          <div className='p-6'>
            <Breadcrumbs />
          </div>
        </div>
        <main className='flex-1 p-6 overflow-auto'>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
