import { useEffect } from 'react';
import { Link } from '@tanstack/react-router';
import { toast } from 'sonner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../ui/alert-dialog';

// Import our new hooks and store
import { useUsers } from '../../hooks/queries/use-users';
import {
  useCreateUser,
  useDeleteUser,
} from '../../hooks/mutations/use-user-mutations';
import { useHealthCheck } from '../../hooks/queries/use-health';
import { useUserStore } from '../../stores/user-store';

// Import types from services
import { API_BASE_URL, type User } from '../../services';

export function Demo() {
  // Server state - handled by Tanstack Query
  const { data: users, isLoading: usersLoading } = useUsers();
  const { data: healthStatus, isLoading: healthLoading } = useHealthCheck();

  // Mutations
  const createUserMutation = useCreateUser();
  const deleteUserMutation = useDeleteUser();

  // Client state - handled by Zustand
  const {
    formData: newUser,
    setFormData: setNewUser,
    deleteDialogOpen,
    userToDelete,
    setDeleteDialog,
  } = useUserStore();

  // Password validation helpers
  const passwordValidation = {
    length: newUser.password?.length >= 8,
    uppercase: /[A-Z]/.test(newUser.password || ''),
    lowercase: /[a-z]/.test(newUser.password || ''),
  };

  const isPasswordValid = passwordValidation.length && passwordValidation.uppercase && passwordValidation.lowercase;

  // Test health check - now handled by useHealthCheck hook
  const testHealthCheck = () => {
    // The health check is automatically handled by the useHealthCheck hook
    // We just need to show a success message when it's successful
    if (healthStatus) {
      toast.success('Health check successful!', {
        description: `Status: ${healthStatus.status} - ${healthStatus.message}`,
      });
    }
  };

  // Users are now automatically fetched by the useUsers hook

  // Create a new user
  const handleCreateUser = (e: React.FormEvent) => {
    e.preventDefault();

    // Frontend validation
    if (!newUser.name.trim() || !newUser.email.trim() || !newUser.password.trim()) {
      toast.error('Validation Error', {
        description: 'Please fill in all fields including password',
      });
      return;
    }

    // Password validation
    if (newUser.password.length < 8) {
      toast.error('Password too short', {
        description: 'Password must be at least 8 characters long',
      });
      return;
    }

    if (!/[A-Z]/.test(newUser.password)) {
      toast.error('Password validation failed', {
        description: 'Password must contain at least one uppercase letter',
      });
      return;
    }

    if (!/[a-z]/.test(newUser.password)) {
      toast.error('Password validation failed', {
        description: 'Password must contain at least one lowercase letter',
      });
      return;
    }

    createUserMutation.mutate({
      name: newUser.name,
      email: newUser.email,
      password: newUser.password,
    });
  };

  // Open delete confirmation dialog
  const openDeleteDialog = (userId: number, userName: string) => {
    setDeleteDialog(true, { id: userId, name: userName });
  };

  // Delete a user
  const handleDeleteUser = () => {
    if (!userToDelete) return;

    deleteUserMutation.mutate(userToDelete.id, {
      onSuccess: () => {
        setDeleteDialog(false);
      },
    });
  };

  // Users are automatically loaded by the useUsers hook
  useEffect(() => {
    if (usersLoading) {
      toast.loading('Loading users...', {
        id: 'fetch-users',
      });
    } else {
      toast.dismiss('fetch-users');
      if ((users || []).length > 0) {
        toast.success('Users loaded successfully!', {
          description: `Found ${(users || []).length} user${(users || []).length !== 1 ? 's' : ''}`,
        });
      }
    }
  }, [usersLoading, users]);

  return (
    <main className='min-h-screen bg-gray-50 dark:bg-gray-900 py-8 px-4'>
      <div className='max-w-4xl mx-auto space-y-8'>
        {/* Header */}
        <header className='text-center'>
          <h1 className='text-4xl font-bold text-gray-900 dark:text-white mb-2'>
            React + Go Integration Test
          </h1>
          <p className='text-gray-600 dark:text-gray-300'>
            Test your backend API endpoints from this React frontend
          </p>
        </header>

        {/* Health Check Section */}
        <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-6'>
          <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
            🔍 Health Check
          </h2>
          <div className='space-y-4'>
            <button
              onClick={testHealthCheck}
              disabled={healthLoading}
              className='bg-blue-600 hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-600 disabled:bg-blue-300 dark:disabled:bg-blue-800 text-white font-medium py-2 px-4 rounded-lg transition-colors'
            >
              {healthLoading ? 'Testing...' : 'Test Health Check'}
            </button>

            {healthStatus && (
              <div className='p-3 bg-green-100 dark:bg-green-900 border border-green-300 dark:border-green-700 rounded-lg'>
                <p className='text-green-700 dark:text-green-300'>
                  ✅ Status: {healthStatus.status}
                </p>
                <p className='text-green-600 dark:text-green-400 text-sm'>
                  {healthStatus.message}
                </p>
              </div>
            )}
          </div>
        </section>

        {/* Create User Section */}
        <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-6'>
          <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
            👤 Create User
          </h2>
          <form onSubmit={handleCreateUser} className='space-y-4'>
            <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
              <div>
                <label
                  htmlFor='name'
                  className='block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1'
                >
                  Name
                </label>
                <input
                  type='text'
                  id='name'
                  value={newUser.name}
                  onChange={e =>
                    setNewUser({ ...newUser, name: e.target.value })
                  }
                  className='w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white'
                  placeholder='Enter user name'
                  disabled={createUserMutation.isPending}
                />
              </div>
              <div>
                <label
                  htmlFor='email'
                  className='block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1'
                >
                  Email
                </label>
                <input
                  type='email'
                  id='email'
                  value={newUser.email}
                  onChange={e =>
                    setNewUser({ ...newUser, email: e.target.value })
                  }
                  className='w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white'
                  placeholder='Enter email address'
                  disabled={createUserMutation.isPending}
                />
              </div>
              <div className='md:col-span-2'>
                <label
                  htmlFor='password'
                  className='block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1'
                >
                  Password
                </label>
                <input
                  type='password'
                  id='password'
                  value={newUser.password || ''}
                  onChange={e =>
                    setNewUser({ ...newUser, password: e.target.value })
                  }
                  className='w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white'
                  placeholder='Min 8 chars, 1 uppercase, 1 lowercase'
                  disabled={createUserMutation.isPending}
                />
                <div className='text-xs text-gray-500 dark:text-gray-400 mt-1 space-y-1'>
                  <p className='font-medium'>Password requirements:</p>
                  <div className='flex flex-col space-y-1'>
                    <div className={`flex items-center ${passwordValidation.length ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                      <span className={`mr-2 ${passwordValidation.length ? '✓' : '✗'}`}></span>
                      At least 8 characters
                    </div>
                    <div className={`flex items-center ${passwordValidation.uppercase ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                      <span className={`mr-2 ${passwordValidation.uppercase ? '✓' : '✗'}`}></span>
                      At least 1 uppercase letter
                    </div>
                    <div className={`flex items-center ${passwordValidation.lowercase ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                      <span className={`mr-2 ${passwordValidation.lowercase ? '✓' : '✗'}`}></span>
                      At least 1 lowercase letter
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <button
              type='submit'
              disabled={createUserMutation.isPending || !isPasswordValid}
              className='bg-green-600 hover:bg-green-700 dark:bg-green-700 dark:hover:bg-green-600 disabled:bg-green-300 dark:disabled:bg-green-800 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:cursor-not-allowed'
            >
              {createUserMutation.isPending ? 'Creating...' : isPasswordValid ? 'Create User' : 'Complete Password Requirements'}
            </button>
          </form>
        </section>

        {/* Users List Section */}
        <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-6'>
          <div className='flex justify-between items-center mb-4'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white'>
              📋 Users List
            </h2>
            <button
              onClick={() => {
                if ((users || []).length > 0) {
                  toast.info('Refreshing users list...');
                }
              }}
              disabled={usersLoading}
              className='bg-gray-600 hover:bg-gray-700 dark:bg-gray-700 dark:hover:bg-gray-600 disabled:bg-gray-300 dark:disabled:bg-gray-800 text-white font-medium py-2 px-4 rounded-lg transition-colors'
            >
              {usersLoading ? 'Loading...' : 'Refresh'}
            </button>
          </div>

          {usersLoading ? (
            <div className='text-center py-8'>
              <p className='text-gray-600 dark:text-gray-300'>
                Loading users...
              </p>
            </div>
          ) : (users || []).length === 0 ? (
            <div className='text-center py-8'>
              <p className='text-gray-500 dark:text-gray-400'>
                No users found. Create one above!
              </p>
            </div>
          ) : (
            <div className='overflow-x-auto'>
              <table className='w-full border-collapse'>
                <thead>
                  <tr className='border-b border-gray-200 dark:border-gray-600'>
                    <th className='text-left py-2 px-4 text-gray-900 dark:text-white font-medium'>
                      ID
                    </th>
                    <th className='text-left py-2 px-4 text-gray-900 dark:text-white font-medium'>
                      Name
                    </th>
                    <th className='text-left py-2 px-4 text-gray-900 dark:text-white font-medium'>
                      Email
                    </th>
                    <th className='text-left py-2 px-4 text-gray-900 dark:text-white font-medium'>
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {(() => {
                    console.log(
                      'Demo - Users array before map:',
                      users,
                      'Type:',
                      typeof users,
                      'Is array:',
                      Array.isArray(users)
                    );
                    return (users || []).map(user => {
                      const userUrl = `/users/${user.id}`;
                      console.log('Demo - User data:', {
                        id: user.id,
                        name: user.name,
                        email: user.email,
                        generatedUrl: userUrl,
                      });

                      return (
                        <tr
                          key={user.id}
                          className='border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
                        >
                          <td className='py-2 px-4 text-gray-700 dark:text-gray-300'>
                            {user.id}
                          </td>
                          <td className='py-2 px-4'>
                            <Link
                              to='/users/$userId'
                              params={{ userId: user.id.toString() }}
                              search={{}}
                              className='text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline font-medium'
                            >
                              {user.name}
                            </Link>
                          </td>
                          <td className='py-2 px-4'>
                            <Link
                              to='/users/$userId'
                              params={{ userId: user.id.toString() }}
                              search={{}}
                              className='text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline'
                            >
                              {user.email}
                            </Link>
                          </td>
                          <td className='py-2 px-4'>
                            <button
                              onClick={() =>
                                openDeleteDialog(user.id, user.name)
                              }
                              className='bg-red-600 hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600 text-white font-medium py-1 px-3 rounded text-sm transition-colors'
                            >
                              Delete
                            </button>
                          </td>
                        </tr>
                      );
                    });
                  })()}
                </tbody>
              </table>
            </div>
          )}
        </section>

        {/* Backend Status */}
        <section className='bg-white dark:bg-gray-800 rounded-lg shadow-md p-6'>
          <h2 className='text-2xl font-semibold text-gray-900 dark:text-white mb-4'>
            🔧 Backend Status
          </h2>
          <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
            <div className='p-4 bg-gray-100 dark:bg-gray-700 rounded-lg'>
              <h3 className='font-medium text-gray-900 dark:text-white mb-2'>
                API Endpoints
              </h3>
              <ul className='text-sm text-gray-600 dark:text-gray-300 space-y-1'>
                <li>GET /api/health - Health check</li>
                <li>GET /api/users - List all users</li>
                <li>POST /api/users - Create user</li>
                <li>GET /api/users/:id - Get user by ID</li>
                <li>PUT /api/users/:id - Update user</li>
                <li>DELETE /api/users/:id - Delete user</li>
              </ul>
            </div>
            <div className='p-4 bg-gray-100 dark:bg-gray-700 rounded-lg'>
              <h3 className='font-medium text-gray-900 dark:text-white mb-2'>
                Server Details
              </h3>
              <ul className='text-sm text-gray-600 dark:text-gray-300 space-y-1'>
                <li>Base URL: {API_BASE_URL}</li>
                <li>Database: PostgreSQL</li>
                <li>Framework: Go + Chi Router</li>
                <li>CORS: Enabled for React dev server</li>
              </ul>
            </div>
          </div>
        </section>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={deleteDialogOpen}
        onOpenChange={open => setDeleteDialog(open)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete user "{userToDelete?.name}"? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setDeleteDialog(false)}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteUser}
              className='bg-red-600 hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600'
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </main>
  );
}
