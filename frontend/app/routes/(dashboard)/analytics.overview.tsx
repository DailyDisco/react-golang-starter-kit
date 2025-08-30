import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/(dashboard)/analytics/overview')({
  component: AnalyticsOverview,
});

function AnalyticsOverview() {
  return (
    <div className='space-y-6'>
      <div className='grid grid-cols-1 gap-6 lg:grid-cols-2'>
        <div className='bg-muted/50 rounded-lg p-6'>
          <h3 className='mb-4 font-semibold'>Traffic Overview</h3>
          <div className='bg-muted flex h-64 items-center justify-center rounded'>
            <span className='text-muted-foreground'>
              Traffic Chart Placeholder
            </span>
          </div>
        </div>

        <div className='bg-muted/50 rounded-lg p-6'>
          <h3 className='mb-4 font-semibold'>User Demographics</h3>
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

      <div className='bg-muted/50 rounded-lg p-6'>
        <h3 className='mb-4 font-semibold'>Recent Activity</h3>
        <div className='space-y-2'>
          <div className='border-border flex justify-between border-b py-2'>
            <span>User registration</span>
            <span className='text-muted-foreground text-sm'>2 minutes ago</span>
          </div>
          <div className='border-border flex justify-between border-b py-2'>
            <span>Page view</span>
            <span className='text-muted-foreground text-sm'>5 minutes ago</span>
          </div>
          <div className='flex justify-between py-2'>
            <span>Form submission</span>
            <span className='text-muted-foreground text-sm'>
              10 minutes ago
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
