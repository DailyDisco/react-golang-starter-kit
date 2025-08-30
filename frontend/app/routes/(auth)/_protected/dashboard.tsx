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

      <div className='grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3'>
        <div className='bg-card rounded-lg border p-6'>
          <h3 className='mb-2 font-semibold'>Welcome Back!</h3>
          <p className='text-muted-foreground text-sm'>
            You're viewing a protected area of the application.
          </p>
        </div>

        <div className='bg-card rounded-lg border p-6'>
          <h3 className='mb-2 font-semibold'>Your Profile</h3>
          <p className='text-muted-foreground text-sm'>
            Manage your account settings and preferences.
          </p>
        </div>

        <div className='bg-card rounded-lg border p-6'>
          <h3 className='mb-2 font-semibold'>Recent Activity</h3>
          <p className='text-muted-foreground text-sm'>
            View your latest actions and updates.
          </p>
        </div>
      </div>
    </div>
  );
}
