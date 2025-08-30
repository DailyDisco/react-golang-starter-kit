import { createFileRoute } from '@tanstack/react-router';

import { Button } from '../../components/ui/button';

export const Route = createFileRoute('/(dashboard)/settings')({
  component: SettingsPage,
  // Add a loader for data fetching
  loader: async () => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 100));
    return { userSettings: { theme: 'dark', notifications: true } };
  },
});

function SettingsPage() {
  const data = Route.useLoaderData();

  return (
    <div className='mx-auto max-w-2xl px-4 py-8'>
      <h1 className='mb-6 text-2xl font-bold'>Settings</h1>

      <div className='space-y-6'>
        <div className='bg-card rounded-lg border p-6'>
          <h2 className='mb-4 text-lg font-semibold'>User Preferences</h2>
          <p className='text-muted-foreground mb-4'>
            Theme: {data.userSettings.theme}
            <br />
            Notifications:{' '}
            {data.userSettings.notifications ? 'Enabled' : 'Disabled'}
          </p>
          <Button>Update Settings</Button>
        </div>
      </div>
    </div>
  );
}
