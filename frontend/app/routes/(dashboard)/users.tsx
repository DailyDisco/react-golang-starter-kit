import {
  createFileRoute,
  Link,
  useNavigate,
  useParams,
} from '@tanstack/react-router';
import {
  ArrowLeft,
  Edit3,
  Loader2,
  Lock,
  Save,
  User as UserIcon,
} from 'lucide-react';
import { useEffect } from 'react';
import { toast } from 'sonner';

import { Button } from '../../components/ui/button';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '../../components/ui/card';
import { Input } from '../../components/ui/input';
import { Label } from '../../components/ui/label';
import { useUpdateUser } from '../../hooks/mutations/use-user-mutations';
import { useUser, useUsers } from '../../hooks/queries/use-users';
import { useUserStore } from '../../stores/user-store';

export const Route = createFileRoute('/(dashboard)/users')({
  component: UserDetailPage,
});

function UserDetailPage() {
  const { userId } = useParams({ from: '/users/$userId' });
  const navigate = useNavigate();

  // Server state - handled by Tanstack Query
  const { data: users, isLoading: usersLoading } = useUsers();
  const { data: user, isLoading: userLoading } = useUser();

  // Client state - handled by Zustand
  const { editMode, formData, setEditMode, setFormData, setSelectedUser } =
    useUserStore();

  // Mutations
  const updateUserMutation = useUpdateUser();

  // Sync user data to client state when it loads
  useEffect(() => {
    if (user && !editMode) {
      setFormData({ name: user.name, email: user.email });
      setSelectedUser(user.id);
    }
  }, [user, editMode, setFormData, setSelectedUser]);

  // Set selected user when userId changes
  useEffect(() => {
    if (userId) {
      setSelectedUser(Number(userId));
    }
  }, [userId, setSelectedUser]);

  const handleSave = () => {
    if (!user) return;

    updateUserMutation.mutate(
      {
        ...user,
        name: formData.name.trim(),
        email: formData.email.trim(),
      },
      {
        onSuccess: () => {
          setEditMode(false);
        },
      }
    );
  };

  const handleCancel = () => {
    if (user) {
      setFormData({ name: user.name, email: user.email });
    }
    setEditMode(false);
  };

  if (usersLoading || userLoading) {
    return (
      <div className='flex min-h-[400px] items-center justify-center'>
        <Loader2 className='h-8 w-8 animate-spin' />
      </div>
    );
  }

  if (!user) {
    return (
      <div className='py-12 text-center'>
        <UserIcon className='mx-auto mb-4 h-12 w-12 text-gray-400' />
        <h3 className='mb-2 text-lg font-medium text-gray-900 dark:text-white'>
          User not found
        </h3>
        <p className='mb-4 text-gray-600 dark:text-gray-300'>
          The user with ID {userId} could not be found
        </p>
        <Link to='/demo' search={{}}>
          <Button>
            <ArrowLeft className='mr-2 h-4 w-4' />
            Back to Demo
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <main className='bg-gray-50 px-4 py-12 dark:bg-gray-900'>
      <div className='mx-auto max-w-2xl'>
        {/* Back Button */}
        <div className='mb-6'>
          <Link to='/demo' search={{}}>
            <Button variant='outline' size='sm'>
              <ArrowLeft className='mr-2 h-4 w-4' />
              Back to Demo
            </Button>
          </Link>
        </div>

        {/* Header */}
        <header className='mb-8'>
          <h1 className='mb-2 text-3xl font-bold text-gray-900 dark:text-white'>
            User Details
          </h1>
          <p className='text-gray-600 dark:text-gray-300'>
            View and edit user information
          </p>
        </header>

        {/* User Card */}
        <Card className='shadow-lg'>
          <CardHeader>
            <CardTitle className='flex items-center justify-between'>
              <span>{editMode ? 'Edit User' : 'User Information'}</span>
              {!editMode && (
                <Button onClick={() => setEditMode(true)}>
                  <Edit3 className='mr-2 h-4 w-4' />
                  Edit
                </Button>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-6'>
            {/* User ID (read-only) */}
            <div>
              <Label htmlFor='userId' className='flex items-center gap-2'>
                <Lock className='h-3 w-3 text-gray-500' />
                User ID
              </Label>
              <Input
                id='userId'
                value={user.id}
                disabled
                className='bg-gray-100 dark:bg-gray-800'
              />
            </div>

            {/* Name */}
            <div>
              <Label htmlFor='name'>Name</Label>
              {editMode ? (
                <Input
                  id='name'
                  value={formData.name}
                  onChange={e =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  placeholder='Enter user name'
                />
              ) : (
                <Input
                  id='name'
                  value={user.name}
                  disabled
                  className='bg-gray-100 dark:bg-gray-800'
                />
              )}
            </div>

            {/* Email */}
            <div>
              <Label htmlFor='email'>Email</Label>
              {editMode ? (
                <Input
                  id='email'
                  type='email'
                  value={formData.email}
                  onChange={e =>
                    setFormData({ ...formData, email: e.target.value })
                  }
                  placeholder='Enter email address'
                />
              ) : (
                <Input
                  id='email'
                  type='email'
                  value={user.email}
                  disabled
                  className='bg-gray-100 dark:bg-gray-800'
                />
              )}
            </div>

            {/* Action Buttons */}
            {editMode && (
              <div className='flex gap-2'>
                <Button
                  onClick={handleSave}
                  disabled={
                    updateUserMutation.isPending ||
                    !formData.name.trim() ||
                    !formData.email.trim()
                  }
                  className='flex-1'
                >
                  {updateUserMutation.isPending ? (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  ) : (
                    <Save className='mr-2 h-4 w-4' />
                  )}
                  Save
                </Button>
                <Button
                  variant='outline'
                  onClick={handleCancel}
                  disabled={updateUserMutation.isPending}
                >
                  Cancel
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </main>
  );
}
