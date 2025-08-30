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
    <div className='max-w-2xl mx-auto py-8 px-4'>
      <h1 className='text-2xl font-bold mb-6'>Settings</h1>

      <div className='space-y-6'>
        <div className='bg-card p-6 rounded-lg border'>
          <h2 className='text-lg font-semibold mb-4'>User Preferences</h2>
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
