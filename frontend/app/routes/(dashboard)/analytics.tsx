import { createFileRoute, Link, Outlet } from '@tanstack/react-router';
import { BarChart3, TrendingUp, Users } from 'lucide-react';

import { Button } from '../../components/ui/button';

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

      <div className='grid grid-cols-1 gap-6 md:grid-cols-3'>
        <div className='bg-card rounded-lg border p-6'>
          <div className='flex items-center gap-3'>
            <Users className='h-8 w-8 text-blue-500' />
            <div>
              <p className='text-muted-foreground text-sm'>Total Users</p>
              <p className='text-2xl font-bold'>1,234</p>
            </div>
          </div>
        </div>

        <div className='bg-card rounded-lg border p-6'>
          <div className='flex items-center gap-3'>
            <BarChart3 className='h-8 w-8 text-green-500' />
            <div>
              <p className='text-muted-foreground text-sm'>Page Views</p>
              <p className='text-2xl font-bold'>45,678</p>
            </div>
          </div>
        </div>

        <div className='bg-card rounded-lg border p-6'>
          <div className='flex items-center gap-3'>
            <TrendingUp className='h-8 w-8 text-purple-500' />
            <div>
              <p className='text-muted-foreground text-sm'>Conversion Rate</p>
              <p className='text-2xl font-bold'>12.5%</p>
            </div>
          </div>
        </div>
      </div>

      <div className='bg-card rounded-lg border p-6'>
        <h2 className='mb-4 text-lg font-semibold'>Detailed Analytics</h2>
        <Outlet />
      </div>
    </div>
  );
}
