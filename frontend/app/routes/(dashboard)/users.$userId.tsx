import { createFileRoute } from '@tanstack/react-router';
import { Button } from '../../components/ui/button';
import { ArrowLeft } from 'lucide-react';

export const Route = createFileRoute('/(dashboard)/users/$userId')({
  component: UserDetailPage,
  // Validate the userId parameter
  validateSearch: (search) => ({
    tab: search.tab as 'profile' | 'settings' | undefined,
  }),
  // Loader with parameter validation
  loader: async ({ params }) => {
    const userId = Number(params.userId);
    if (isNaN(userId)) {
      throw new Error('Invalid user ID');
    }

    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 200));
    return {
      user: {
        id: userId,
        name: `User ${userId}`,
        email: `user${userId}@example.com`,
        role: userId === 1 ? 'Admin' : 'User',
      }
    };
  },
});

function UserDetailPage() {
  const { userId } = Route.useParams();
  const data = Route.useLoaderData();
  const navigate = Route.useNavigate();

  return (
    <div className='max-w-2xl mx-auto py-8 px-4'>
      <div className='mb-6'>
        <Button
          variant='outline'
          onClick={() => navigate({ to: '/(dashboard)/users' })}
        >
          <ArrowLeft className='w-4 h-4 mr-2' />
          Back to Users
        </Button>
      </div>

      <div className='bg-card p-6 rounded-lg border'>
        <h1 className='text-2xl font-bold mb-4'>User Details</h1>

        <div className='space-y-3'>
          <div>
            <strong>ID:</strong> {data.user.id}
          </div>
          <div>
            <strong>Name:</strong> {data.user.name}
          </div>
          <div>
            <strong>Email:</strong> {data.user.email}
          </div>
          <div>
            <strong>Role:</strong> {data.user.role}
          </div>
        </div>

        <div className='mt-6'>
          <Button>Edit User</Button>
        </div>
      </div>
    </div>
  );
}