import { createFileRoute, Outlet, Link } from '@tanstack/react-router';
import { Button } from '../../components/ui/button';
import { BarChart3, Users, TrendingUp } from 'lucide-react';

export const Route = createFileRoute('/(dashboard)/analytics')({
  component: AnalyticsLayout,
});

function AnalyticsLayout() {
  return (
    <div className='space-y-6'>
      <div className='flex items-center justify-between'>
        <h1 className='text-2xl font-bold'>Analytics Dashboard</h1>
        <div className='flex gap-2'>
          <Button variant='outline' size='sm'>
            Export
          </Button>
          <Button size='sm'>Refresh</Button>
        </div>
      </div>

      <div className='grid grid-cols-1 md:grid-cols-3 gap-6'>
        <div className='bg-card p-6 rounded-lg border'>
          <div className='flex items-center gap-3'>
            <Users className='w-8 h-8 text-blue-500' />
            <div>
              <p className='text-sm text-muted-foreground'>Total Users</p>
              <p className='text-2xl font-bold'>1,234</p>
            </div>
          </div>
        </div>

        <div className='bg-card p-6 rounded-lg border'>
          <div className='flex items-center gap-3'>
            <BarChart3 className='w-8 h-8 text-green-500' />
            <div>
              <p className='text-sm text-muted-foreground'>Page Views</p>
              <p className='text-2xl font-bold'>45,678</p>
            </div>
          </div>
        </div>

        <div className='bg-card p-6 rounded-lg border'>
          <div className='flex items-center gap-3'>
            <TrendingUp className='w-8 h-8 text-purple-500' />
            <div>
              <p className='text-sm text-muted-foreground'>Conversion Rate</p>
              <p className='text-2xl font-bold'>12.5%</p>
            </div>
          </div>
        </div>
      </div>

      <div className='bg-card p-6 rounded-lg border'>
        <h2 className='text-lg font-semibold mb-4'>Detailed Analytics</h2>
        <Outlet />
      </div>
    </div>
  );
}
