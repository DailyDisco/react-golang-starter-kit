import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
import { useAuth } from '../../providers/AuthContext';
import { Button } from '../../components/ui/button';
import { Link } from '@tanstack/react-router';

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
  const { logout } = useAuth();

  return (
    <div className='min-h-screen bg-background'>
      {/* Protected Header */}
      <header className='bg-card border-b px-6 py-4'>
        <div className='flex justify-between items-center'>
          <div className='flex items-center gap-4'>
            <h1 className='text-xl font-semibold'>Dashboard</h1>
            <nav className='flex gap-4'>
              <Link to='/(dashboard)/users' className='text-muted-foreground hover:text-foreground'>
                Users
              </Link>
              <Link to='/(dashboard)/settings' className='text-muted-foreground hover:text-foreground'>
                Settings
              </Link>
              <Link to='/(dashboard)/analytics' className='text-muted-foreground hover:text-foreground'>
                Analytics
              </Link>
            </nav>
          </div>
          <Button variant='outline' onClick={logout}>
            Logout
          </Button>
        </div>
      </header>

      {/* Protected Content */}
      <main className='p-6'>
        <Outlet />
      </main>
    </div>
  );
}