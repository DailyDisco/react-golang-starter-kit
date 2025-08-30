import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/(auth)/_protected/dashboard')({
  component: ProtectedDashboard,
});

function ProtectedDashboard() {
  return (
    <div className='space-y-6'>
      <div>
        <h1 className='text-3xl font-bold'>Protected Dashboard</h1>
        <p className='text-muted-foreground mt-2'>
          This page is protected and requires authentication to access.
        </p>
      </div>

      <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6'>
        <div className='bg-card p-6 rounded-lg border'>
          <h3 className='font-semibold mb-2'>Welcome Back!</h3>
          <p className='text-sm text-muted-foreground'>
            You're viewing a protected area of the application.
          </p>
        </div>

        <div className='bg-card p-6 rounded-lg border'>
          <h3 className='font-semibold mb-2'>Your Profile</h3>
          <p className='text-sm text-muted-foreground'>
            Manage your account settings and preferences.
          </p>
        </div>

        <div className='bg-card p-6 rounded-lg border'>
          <h3 className='font-semibold mb-2'>Recent Activity</h3>
          <p className='text-sm text-muted-foreground'>
            View your latest actions and updates.
          </p>
        </div>
      </div>
    </div>
  );
}
