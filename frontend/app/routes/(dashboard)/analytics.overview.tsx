import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/(dashboard)/analytics/overview')({
  component: AnalyticsOverview,
});

function AnalyticsOverview() {
  return (
    <div className='space-y-6'>
      <div className='grid grid-cols-1 lg:grid-cols-2 gap-6'>
        <div className='bg-muted/50 p-6 rounded-lg'>
          <h3 className='font-semibold mb-4'>Traffic Overview</h3>
          <div className='h-64 bg-muted rounded flex items-center justify-center'>
            <span className='text-muted-foreground'>
              Traffic Chart Placeholder
            </span>
          </div>
        </div>

        <div className='bg-muted/50 p-6 rounded-lg'>
          <h3 className='font-semibold mb-4'>User Demographics</h3>
          <div className='space-y-3'>
            <div className='flex justify-between'>
              <span>Desktop</span>
              <span>65%</span>
            </div>
            <div className='flex justify-between'>
              <span>Mobile</span>
              <span>30%</span>
            </div>
            <div className='flex justify-between'>
              <span>Tablet</span>
              <span>5%</span>
            </div>
          </div>
        </div>
      </div>

      <div className='bg-muted/50 p-6 rounded-lg'>
        <h3 className='font-semibold mb-4'>Recent Activity</h3>
        <div className='space-y-2'>
          <div className='flex justify-between py-2 border-b border-border'>
            <span>User registration</span>
            <span className='text-sm text-muted-foreground'>2 minutes ago</span>
          </div>
          <div className='flex justify-between py-2 border-b border-border'>
            <span>Page view</span>
            <span className='text-sm text-muted-foreground'>5 minutes ago</span>
          </div>
          <div className='flex justify-between py-2'>
            <span>Form submission</span>
            <span className='text-sm text-muted-foreground'>
              10 minutes ago
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
