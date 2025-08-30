import { useEffect, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { toast } from 'sonner';
import { motion, AnimatePresence } from 'framer-motion';
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
import { Skeleton } from '../ui/skeleton';

// Import our new hooks and store
import { useUsers } from '../../hooks/queries/use-users';
import {
  useCreateUser,
  useDeleteUser,
} from '../../hooks/mutations/use-user-mutations';
import { useHealthCheck } from '../../hooks/queries/use-health';
import { useUserStore } from '../../stores/user-store';

// Import types from services
import { API_BASE_URL } from '../../services';

export function Demo() {
  // Framer Motion demo state
  const [showCards, setShowCards] = useState(false);

  // Server state - handled by Tanstack Query
  const { data: users, isLoading: usersLoading } = useUsers();
  const { data: healthStatus, isLoading: healthLoading } = useHealthCheck();

  // Mutations
  const createUserMutation = useCreateUser();
  const deleteUserMutation = useDeleteUser();

  // Client state - handled by Zustand
  const { formData: newUser, setFormData: setNewUser } = useUserStore();

  // Local state for delete dialog
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState<{
    id: number;
    name: string;
  } | null>(null);

  // Password validation helpers
  const passwordValidation = {
    length: newUser.password?.length >= 8,
    uppercase: /[A-Z]/.test(newUser.password || ''),
    lowercase: /[a-z]/.test(newUser.password || ''),
  };

  const isPasswordValid =
    passwordValidation.length &&
    passwordValidation.uppercase &&
    passwordValidation.lowercase;

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
    if (
      !newUser.name.trim() ||
      !newUser.email.trim() ||
      !newUser.password.trim()
    ) {
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
    setUserToDelete({ id: userId, name: userName });
    setDeleteDialogOpen(true);
  };

  // Delete a user
  const handleDeleteUser = () => {
    if (!userToDelete) return;

    deleteUserMutation.mutate(userToDelete.id, {
      onSuccess: () => {
        setDeleteDialogOpen(false);
        setUserToDelete(null);
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
    <main className='min-h-screen bg-gray-50 px-4 py-8 dark:bg-gray-900'>
      <div className='mx-auto max-w-4xl space-y-8'>
        {/* Header */}
        <motion.header
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className='text-center'
        >
          <h1 className='mb-2 text-4xl font-bold text-gray-900 dark:text-white'>
            React + Go Integration Test
          </h1>
          <p className='text-gray-600 dark:text-gray-300'>
            Test your backend API endpoints from this React frontend
          </p>
        </motion.header>

        {/* Health Check Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
            🔍 Health Check
          </h2>
          <div className='space-y-4'>
            <button
              onClick={testHealthCheck}
              disabled={healthLoading}
              className='rounded-lg bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700 disabled:bg-blue-300 dark:bg-blue-700 dark:hover:bg-blue-600 dark:disabled:bg-blue-800'
            >
              {healthLoading ? 'Testing...' : 'Test Health Check'}
            </button>

            {healthStatus && (
              <div className='rounded-lg border border-green-300 bg-green-100 p-3 dark:border-green-700 dark:bg-green-900'>
                <p className='text-green-700 dark:text-green-300'>
                  ✅ Status: {healthStatus.status}
                </p>
                <p className='text-sm text-green-600 dark:text-green-400'>
                  {healthStatus.message}
                </p>
              </div>
            )}
          </div>
        </section>

        {/* Create User Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
            👤 Create User
          </h2>
          <form onSubmit={handleCreateUser} className='space-y-4'>
            <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
              <div>
                <label
                  htmlFor='name'
                  className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
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
                  className='w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white'
                  placeholder='Enter user name'
                  disabled={createUserMutation.isPending}
                />
              </div>
              <div>
                <label
                  htmlFor='email'
                  className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
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
                  className='w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white'
                  placeholder='Enter email address'
                  disabled={createUserMutation.isPending}
                />
              </div>
              <div className='md:col-span-2'>
                <label
                  htmlFor='password'
                  className='mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300'
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
                  className='w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white'
                  placeholder='Min 8 chars, 1 uppercase, 1 lowercase'
                  disabled={createUserMutation.isPending}
                />
                <div className='mt-1 space-y-1 text-xs text-gray-500 dark:text-gray-400'>
                  <p className='font-medium'>Password requirements:</p>
                  <div className='flex flex-col space-y-1'>
                    <div
                      className={`flex items-center ${passwordValidation.length ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}
                    >
                      <span
                        className={`mr-2 ${passwordValidation.length ? '✓' : '✗'}`}
                      ></span>
                      At least 8 characters
                    </div>
                    <div
                      className={`flex items-center ${passwordValidation.uppercase ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}
                    >
                      <span
                        className={`mr-2 ${passwordValidation.uppercase ? '✓' : '✗'}`}
                      ></span>
                      At least 1 uppercase letter
                    </div>
                    <div
                      className={`flex items-center ${passwordValidation.lowercase ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}
                    >
                      <span
                        className={`mr-2 ${passwordValidation.lowercase ? '✓' : '✗'}`}
                      ></span>
                      At least 1 lowercase letter
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <button
              type='submit'
              disabled={createUserMutation.isPending || !isPasswordValid}
              className='rounded-lg bg-green-600 px-4 py-2 font-medium text-white transition-colors hover:bg-green-700 disabled:cursor-not-allowed disabled:bg-green-300 dark:bg-green-700 dark:hover:bg-green-600 dark:disabled:bg-green-800'
            >
              {createUserMutation.isPending
                ? 'Creating...'
                : isPasswordValid
                  ? 'Create User'
                  : 'Complete Password Requirements'}
            </button>
          </form>
        </section>

        {/* Users List Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <div className='mb-4 flex items-center justify-between'>
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
              className='rounded-lg bg-gray-600 px-4 py-2 font-medium text-white transition-colors hover:bg-gray-700 disabled:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 dark:disabled:bg-gray-800'
            >
              {usersLoading ? 'Loading...' : 'Refresh'}
            </button>
          </div>

          {usersLoading ? (
            <div className='py-8 text-center'>
              <p className='text-gray-600 dark:text-gray-300'>
                Loading users...
              </p>
            </div>
          ) : (users || []).length === 0 ? (
            <div className='py-8 text-center'>
              <p className='text-gray-500 dark:text-gray-400'>
                No users found. Create one above!
              </p>
            </div>
          ) : (
            <div className='overflow-x-auto'>
              <table className='w-full border-collapse'>
                <thead>
                  <tr className='border-b border-gray-200 dark:border-gray-600'>
                    <th className='px-4 py-2 text-left font-medium text-gray-900 dark:text-white'>
                      ID
                    </th>
                    <th className='px-4 py-2 text-left font-medium text-gray-900 dark:text-white'>
                      Name
                    </th>
                    <th className='px-4 py-2 text-left font-medium text-gray-900 dark:text-white'>
                      Email
                    </th>
                    <th className='px-4 py-2 text-left font-medium text-gray-900 dark:text-white'>
                      Actions
                    </th>
                  </tr>
                </thead>
                <AnimatePresence>
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
                        <motion.tr
                          key={user.id}
                          initial={{ opacity: 0, height: 0 }}
                          animate={{ opacity: 1, height: 'auto' }}
                          exit={{ opacity: 0, height: 0 }}
                          transition={{ duration: 0.3 }}
                          className='border-b border-gray-100 hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-700'
                        >
                          <td className='px-4 py-2 text-gray-700 dark:text-gray-300'>
                            {user.id}
                          </td>
                          <td className='px-4 py-2'>
                            <Link
                              to='/users/$userId'
                              params={{ userId: user.id.toString() }}
                              search={{}}
                              className='font-medium text-blue-600 hover:text-blue-800 hover:underline dark:text-blue-400 dark:hover:text-blue-300'
                            >
                              {user.name}
                            </Link>
                          </td>
                          <td className='px-4 py-2'>
                            <Link
                              to='/users/$userId'
                              params={{ userId: user.id.toString() }}
                              search={{}}
                              className='text-blue-600 hover:text-blue-800 hover:underline dark:text-blue-400 dark:hover:text-blue-300'
                            >
                              {user.email}
                            </Link>
                          </td>
                          <td className='px-4 py-2'>
                            <button
                              onClick={() =>
                                openDeleteDialog(user.id, user.name)
                              }
                              className='rounded bg-red-600 px-3 py-1 text-sm font-medium text-white transition-colors hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600'
                            >
                              Delete
                            </button>
                          </td>
                        </motion.tr>
                      );
                    });
                  })()}
                </AnimatePresence>
              </table>
            </div>
          )}
        </section>

        {/* Backend Status */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
            🔧 Backend Status
          </h2>
          <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            <div className='rounded-lg bg-gray-100 p-4 dark:bg-gray-700'>
              <h3 className='mb-2 font-medium text-gray-900 dark:text-white'>
                API Endpoints
              </h3>
              <ul className='space-y-1 text-sm text-gray-600 dark:text-gray-300'>
                <li>GET /api/health - Health check</li>
                <li>GET /api/users - List all users</li>
                <li>POST /api/users - Create user</li>
                <li>GET /api/users/:id - Get user by ID</li>
                <li>PUT /api/users/:id - Update user</li>
                <li>DELETE /api/users/:id - Delete user</li>
              </ul>
            </div>
            <div className='rounded-lg bg-gray-100 p-4 dark:bg-gray-700'>
              <h3 className='mb-2 font-medium text-gray-900 dark:text-white'>
                Server Details
              </h3>
              <ul className='space-y-1 text-sm text-gray-600 dark:text-gray-300'>
                <li>Base URL: {API_BASE_URL}</li>
                <li>Database: PostgreSQL</li>
                <li>Framework: Go + Chi Router</li>
                <li>CORS: Enabled for React dev server</li>
              </ul>
            </div>
          </div>
        </section>

        {/* Framer Motion Demo Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
            🎬 Framer Motion Demo
          </h2>
          <p className='mb-6 text-gray-600 dark:text-gray-300'>
            Experience a unique morphing animation that combines multiple
            techniques.
          </p>

          <div className='rounded-lg bg-gray-50 p-6 shadow-inner dark:bg-gray-700'>
            <h3 className='mb-4 text-xl font-semibold text-gray-900 dark:text-white'>
              🔄 Shape Morphing with Staggered Particles
            </h3>
            <p className='mb-6 text-gray-600 dark:text-gray-300'>
              Watch as geometric shapes transform while particles dance around
              them!
            </p>

            <div className='mb-8 flex justify-center'>
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => setShowCards(!showCards)}
                className='rounded-full bg-gradient-to-r from-purple-500 to-pink-500 px-6 py-3 font-medium text-white shadow-lg hover:from-purple-600 hover:to-pink-600'
              >
                {showCards ? 'Reset Animation' : 'Start Morphing Magic'}
              </motion.button>
            </div>

            <div className='relative flex h-96 items-center justify-center overflow-hidden'>
              <AnimatePresence>
                {showCards && (
                  <>
                    {/* Central Morphing Shape */}
                    <motion.div
                      initial={{ scale: 0, rotate: -180 }}
                      animate={{
                        scale: [0, 1.3, 1, 1.1, 1],
                        rotate: [0, 180, 360, 540, 720],
                        borderRadius: ['0%', '50%', '25%', '75%', '0%'],
                        boxShadow: [
                          '0 0 0px rgba(59, 130, 246, 0)',
                          '0 0 20px rgba(59, 130, 246, 0.5)',
                          '0 0 40px rgba(168, 85, 247, 0.8)',
                          '0 0 60px rgba(236, 72, 153, 1)',
                          '0 0 80px rgba(59, 130, 246, 0.6)',
                        ],
                      }}
                      exit={{ scale: 0, rotate: 180, opacity: 0 }}
                      transition={{
                        duration: 3,
                        ease: 'easeInOut',
                        times: [0, 0.2, 0.5, 0.8, 1],
                      }}
                      className='absolute z-10 h-28 w-28 bg-gradient-to-r from-blue-600 via-red-600 via-yellow-400 to-purple-600'
                      style={{
                        filter: 'drop-shadow(0 0 20px rgba(168, 85, 247, 0.6))',
                      }}
                    />

                    {/* Staggered Floating Particles */}
                    <motion.div
                      initial='hidden'
                      animate='visible'
                      exit='exit'
                      variants={{
                        visible: {
                          transition: {
                            staggerChildren: 0.15,
                            delayChildren: 0.5,
                          },
                        },
                        exit: {
                          transition: {
                            staggerChildren: 0.1,
                            staggerDirection: -1,
                          },
                        },
                      }}
                      className='absolute inset-0'
                    >
                      {[...Array(8)].map((_, i) => {
                        const angle = (i / 8) * 360;
                        const radius = 120;
                        const x = Math.cos((angle * Math.PI) / 180) * radius;
                        const y = Math.sin((angle * Math.PI) / 180) * radius;

                        return (
                          <motion.div
                            key={i}
                            variants={{
                              hidden: {
                                opacity: 0,
                                scale: 0,
                                x: 0,
                                y: 0,
                              },
                              visible: {
                                opacity: [0, 1, 0.8, 1],
                                scale: [0, 1.5, 1, 1.2],
                                x: [0, x, x * 0.8, x],
                                y: [0, y, y * 0.8, y],
                                transition: {
                                  duration: 3,
                                  repeat: Infinity,
                                  repeatType: 'reverse',
                                  ease: 'easeInOut',
                                },
                              },
                              exit: {
                                opacity: 0,
                                scale: 0,
                                x: 0,
                                y: 0,
                                transition: { duration: 0.5 },
                              },
                            }}
                            className={`absolute h-4 w-4 rounded-full ${
                              [
                                'bg-red-400',
                                'bg-orange-400',
                                'bg-yellow-400',
                                'bg-green-400',
                                'bg-blue-400',
                                'bg-indigo-400',
                                'bg-purple-400',
                                'bg-pink-400',
                              ][i]
                            } shadow-lg`}
                            style={{
                              filter: `drop-shadow(0 0 8px ${['#ef4444', '#f97316', '#eab308', '#22c55e', '#3b82f6', '#6366f1', '#a855f7', '#ec4899'][i]})`,
                              boxShadow: `0 0 10px ${['#ef4444', '#f97316', '#eab308', '#22c55e', '#3b82f6', '#6366f1', '#a855f7', '#ec4899'][i]}40`,
                            }}
                          />
                        );
                      })}
                    </motion.div>

                    {/* Pulsing Background Effect */}
                    <motion.div
                      initial={{ scale: 0, opacity: 0 }}
                      animate={{
                        scale: [0, 1.5, 2.5, 3.5],
                        opacity: [0, 0.4, 0.2, 0],
                        rotate: [0, 90, 180, 270],
                      }}
                      transition={{
                        duration: 5,
                        repeat: Infinity,
                        ease: 'easeInOut',
                      }}
                      className='absolute h-40 w-40 rounded-full bg-gradient-to-r from-purple-400 via-pink-400 to-blue-400 blur-2xl'
                    />

                    {/* Secondary Pulsing Rings */}
                    <motion.div
                      initial={{ scale: 0, opacity: 0 }}
                      animate={{
                        scale: [0, 2, 3, 4],
                        opacity: [0, 0.2, 0.1, 0],
                        borderWidth: [0, 2, 1, 0],
                      }}
                      transition={{
                        duration: 6,
                        repeat: Infinity,
                        ease: 'easeOut',
                        delay: 1,
                      }}
                      className='absolute h-32 w-32 rounded-full border border-purple-300'
                      style={{
                        filter: 'drop-shadow(0 0 15px rgba(168, 85, 247, 0.3))',
                      }}
                    />

                    {/* Energy Wave Effect */}
                    <motion.div
                      initial={{ scale: 0, opacity: 0 }}
                      animate={{
                        scale: [0, 1.8, 2.8],
                        opacity: [0, 0.3, 0],
                      }}
                      transition={{
                        duration: 4,
                        repeat: Infinity,
                        ease: 'easeOut',
                        delay: 0.5,
                      }}
                      className='absolute h-24 w-24 rounded-full bg-gradient-to-r from-cyan-400 to-purple-400 blur-lg'
                    />
                  </>
                )}
              </AnimatePresence>
            </div>

            <div className='mt-6 text-center text-sm text-gray-500 dark:text-gray-400'>
              ✨ Combines morphing, staggering, and continuous animations
            </div>
          </div>
        </section>

        {/* Skeleton Examples Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
            💀 Skeleton Loading Examples
          </h2>
          <p className='mb-6 text-gray-600 dark:text-gray-300'>
            Various skeleton loading patterns using the shadcn Skeleton
            component to improve perceived performance during data loading.
          </p>

          <div className='space-y-8'>
            {/* Basic Shapes */}
            <div>
              <h3 className='mb-3 text-lg font-medium text-gray-900 dark:text-white'>
                Basic Shapes
              </h3>
              <div className='flex flex-wrap items-center gap-4'>
                <div className='flex flex-col items-center space-y-2'>
                  <Skeleton className='h-4 w-32' />
                  <span className='text-sm text-gray-500 dark:text-gray-400'>
                    Rectangle
                  </span>
                </div>
                <div className='flex flex-col items-center space-y-2'>
                  <Skeleton className='h-12 w-12 rounded-full' />
                  <span className='text-sm text-gray-500 dark:text-gray-400'>
                    Circle
                  </span>
                </div>
                <div className='flex flex-col items-center space-y-2'>
                  <Skeleton className='h-8 w-20 rounded-lg' />
                  <span className='text-sm text-gray-500 dark:text-gray-400'>
                    Rounded
                  </span>
                </div>
                <div className='flex flex-col items-center space-y-2'>
                  <Skeleton className='h-3 w-16 rounded-sm' />
                  <span className='text-sm text-gray-500 dark:text-gray-400'>
                    Small
                  </span>
                </div>
              </div>
            </div>

            {/* Text Content Skeleton */}
            <div>
              <h3 className='mb-3 text-lg font-medium text-gray-900 dark:text-white'>
                Text Content
              </h3>
              <div className='space-y-3'>
                <div className='space-y-2'>
                  <Skeleton className='h-6 w-3/4' />
                  <Skeleton className='h-4 w-full' />
                  <Skeleton className='h-4 w-2/3' />
                </div>
                <div className='space-y-2'>
                  <Skeleton className='h-5 w-1/2' />
                  <Skeleton className='h-4 w-full' />
                  <Skeleton className='h-4 w-4/5' />
                  <Skeleton className='h-4 w-3/4' />
                </div>
              </div>
            </div>

            {/* Card Skeleton */}
            <div>
              <h3 className='mb-3 text-lg font-medium text-gray-900 dark:text-white'>
                Card Layout
              </h3>
              <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
                <div className='rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800'>
                  <div className='flex items-center space-x-4'>
                    <Skeleton className='h-12 w-12 rounded-full' />
                    <div className='space-y-2'>
                      <Skeleton className='h-4 w-32' />
                      <Skeleton className='h-3 w-24' />
                    </div>
                  </div>
                  <div className='mt-4 space-y-2'>
                    <Skeleton className='h-4 w-full' />
                    <Skeleton className='h-4 w-3/4' />
                  </div>
                </div>
                <div className='rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800'>
                  <div className='flex items-center space-x-4'>
                    <Skeleton className='h-12 w-12 rounded-full' />
                    <div className='space-y-2'>
                      <Skeleton className='h-4 w-40' />
                      <Skeleton className='h-3 w-28' />
                    </div>
                  </div>
                  <div className='mt-4 space-y-2'>
                    <Skeleton className='h-4 w-full' />
                    <Skeleton className='h-4 w-2/3' />
                  </div>
                </div>
              </div>
            </div>

            {/* User Profile Skeleton */}
            <div>
              <h3 className='mb-3 text-lg font-medium text-gray-900 dark:text-white'>
                User Profile
              </h3>
              <div className='rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800'>
                <div className='flex items-start space-x-6'>
                  <Skeleton className='h-20 w-20 rounded-full' />
                  <div className='flex-1 space-y-3'>
                    <Skeleton className='h-7 w-48' />
                    <Skeleton className='h-4 w-32' />
                    <div className='space-y-2'>
                      <Skeleton className='h-4 w-full' />
                      <Skeleton className='h-4 w-5/6' />
                      <Skeleton className='h-4 w-4/6' />
                    </div>
                    <div className='flex space-x-2 pt-2'>
                      <Skeleton className='h-8 w-20 rounded' />
                      <Skeleton className='h-8 w-24 rounded' />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete user "{userToDelete?.name}"? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setDeleteDialogOpen(false)}>
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
