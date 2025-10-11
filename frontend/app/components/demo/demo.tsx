import { Link } from '@tanstack/react-router';
import { AnimatePresence, motion } from 'framer-motion';
import { useState } from 'react';
import { toast } from 'sonner';

import {
  useCreateUser,
  useDeleteUser,
} from '../../hooks/mutations/use-user-mutations';
import {
  useFileUpload,
  useFileDelete,
} from '../../hooks/mutations/use-file-mutations';
import { useHealthCheck } from '../../hooks/queries/use-health';
// Import our new hooks and store
import { useUsers } from '../../hooks/queries/use-users';
import {
  useFiles,
  useStorageStatus,
  useFileDownload,
} from '../../hooks/queries/use-files';
// Import types from services
import { API_BASE_URL } from '../../services';
import { useUserStore } from '../../stores/user-store';
import { useFileStore } from '../../stores/file-store';
import { useAuthStore } from '../../stores/auth-store';
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

export function Demo() {
  // Server state - handled by Tanstack Query
  const { data: users, isLoading: usersLoading } = useUsers();
  const { data: healthStatus, isLoading: healthLoading } = useHealthCheck();
  const { data: files, isLoading: filesLoading } = useFiles();
  const { data: storageStatus } = useStorageStatus();

  // File download hook
  const { downloadFile } = useFileDownload();

  // Mutations
  const createUserMutation = useCreateUser();
  const deleteUserMutation = useDeleteUser();
  const fileUploadMutation = useFileUpload();
  const fileDeleteMutation = useFileDelete();

  // Client state - handled by Zustand
  const { formData: newUser, setFormData: setNewUser } = useUserStore();
  const {
    selectedFile,
    isDragOver,
    setSelectedFile,
    setIsDragOver,
    resetFileSelection,
  } = useFileStore();
  const { accessToken: authToken } = useAuthStore();

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
      const overallStatus = healthStatus.overall_status;
      const components = healthStatus.components;

      toast.success('Health check successful!', {
        description: `Overall: ${overallStatus.toUpperCase()}`,
      });

      // Show individual component statuses
      if (components && components.length > 0) {
        for (const component of components) {
          const statusIcon = component.status === 'healthy' ? '✅' : '❌';
          toast.info(`${statusIcon} ${component.name}: ${component.status}`, {
            description: component.message || 'No issues detected',
          });
        }
      }
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

  // File upload handlers
  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  };

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragOver(false);

    const file = event.dataTransfer.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  };

  const handleFileUpload = () => {
    if (!selectedFile) {
      toast.error('No file selected', {
        description: 'Please select a file to upload',
      });
      return;
    }

    fileUploadMutation.mutate(selectedFile, {
      onSuccess: () => {
        resetFileSelection();
      },
    });
  };

  // File delete handler
  const handleFileDelete = (fileId: number, _fileName: string) => {
    fileDeleteMutation.mutate(fileId);
  };

  // File download handler
  const handleFileDownload = (fileId: number, fileName: string) => {
    downloadFile(fileId, fileName);
  };

  // Users are automatically loaded by the useUsers hook

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
              <div className='space-y-3'>
                {/* Overall Status */}
                <div
                  className={`rounded-lg border p-3 ${
                    healthStatus.overall_status === 'healthy'
                      ? 'border-green-300 bg-green-100 dark:border-green-700 dark:bg-green-900'
                      : 'border-red-300 bg-red-100 dark:border-red-700 dark:bg-red-900'
                  }`}
                >
                  <p
                    className={`font-medium ${
                      healthStatus.overall_status === 'healthy'
                        ? 'text-green-700 dark:text-green-300'
                        : 'text-red-700 dark:text-red-300'
                    }`}
                  >
                    {healthStatus.overall_status === 'healthy' ? '✅' : '❌'}{' '}
                    Overall Status: {healthStatus.overall_status.toUpperCase()}
                  </p>
                  <p className='text-sm text-gray-600 dark:text-gray-400'>
                    Checked at:{' '}
                    {new Date(healthStatus.timestamp).toLocaleString()}
                  </p>
                </div>

                {/* Component Statuses */}
                {healthStatus.components &&
                  healthStatus.components.length > 0 && (
                    <div className='space-y-2'>
                      <h4 className='text-sm font-medium text-gray-700 dark:text-gray-300'>
                        Component Details:
                      </h4>
                      {healthStatus.components.map((component, index) => (
                        <div
                          key={index}
                          className={`rounded-lg border p-2 text-sm ${
                            component.status === 'healthy'
                              ? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950'
                              : 'border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950'
                          }`}
                        >
                          <div className='flex items-center justify-between'>
                            <span className='font-medium capitalize'>
                              {component.name}
                            </span>
                            <span
                              className={`font-medium ${
                                component.status === 'healthy'
                                  ? 'text-green-700 dark:text-green-300'
                                  : 'text-red-700 dark:text-red-300'
                              }`}
                            >
                              {component.status === 'healthy' ? '✅' : '❌'}{' '}
                              {component.status}
                            </span>
                          </div>
                          {component.message && (
                            <p className='mt-1 text-xs text-gray-600 dark:text-gray-400'>
                              {component.message}
                            </p>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
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

        {/* File Upload Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <h2 className='mb-4 text-2xl font-semibold text-gray-900 dark:text-white'>
            📁 File Upload Demo
          </h2>
          <p className='mb-6 text-gray-600 dark:text-gray-300'>
            Upload files to the server. Files are automatically stored in S3 if
            configured, otherwise stored in the database.{' '}
            <strong className='text-blue-600 dark:text-blue-400'>
              You must be logged in to upload files.
            </strong>
          </p>

          <div className='space-y-4'>
            {/* File Upload Area */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              className={`relative rounded-lg border-2 border-dashed p-8 text-center transition-colors ${
                isDragOver
                  ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                  : 'border-gray-300 dark:border-gray-600'
              }`}
            >
              <input
                type='file'
                onChange={handleFileSelect}
                className='absolute inset-0 h-full w-full cursor-pointer opacity-0'
                disabled={fileUploadMutation.isPending}
              />

              <div className='space-y-4'>
                <div className='mx-auto h-12 w-12 text-gray-400'>
                  <svg
                    fill='none'
                    stroke='currentColor'
                    viewBox='0 0 24 24'
                    className='h-full w-full'
                  >
                    <path
                      strokeLinecap='round'
                      strokeLinejoin='round'
                      strokeWidth={1.5}
                      d='M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12'
                    />
                  </svg>
                </div>

                <div>
                  <p className='text-lg font-medium text-gray-900 dark:text-white'>
                    {selectedFile
                      ? selectedFile.name
                      : 'Drop files here or click to browse'}
                  </p>
                  <p className='text-sm text-gray-500 dark:text-gray-400'>
                    {selectedFile
                      ? `${(selectedFile.size / 1024).toFixed(2)} KB • ${
                          selectedFile.type || 'Unknown type'
                        }`
                      : 'Supports any file type up to 10MB'}
                  </p>

                  {/* Storage Indicator */}
                  {storageStatus && storageStatus.storage_type && (
                    <div className='mt-3 flex items-center justify-center space-x-2'>
                      <div
                        className={`h-2 w-2 rounded-full ${
                          storageStatus.storage_type === 's3'
                            ? 'bg-green-500'
                            : 'bg-orange-500'
                        }`}
                      />
                      <span className='text-xs font-medium text-gray-600 dark:text-gray-400'>
                        Uploading to: {storageStatus.storage_type.toUpperCase()}
                      </span>
                    </div>
                  )}
                </div>

                {selectedFile && (
                  <button
                    onClick={(e: React.MouseEvent) => {
                      e.stopPropagation();
                      resetFileSelection();
                    }}
                    className='text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300'
                  >
                    Clear selection
                  </button>
                )}
              </div>
            </div>

            {/* Upload Button */}
            <button
              onClick={handleFileUpload}
              disabled={
                !selectedFile || fileUploadMutation.isPending || !authToken
              }
              className='w-full rounded-lg bg-green-600 px-4 py-3 font-medium text-white transition-colors hover:bg-green-700 disabled:cursor-not-allowed disabled:bg-green-300 dark:bg-green-700 dark:hover:bg-green-600 dark:disabled:bg-green-800'
            >
              {fileUploadMutation.isPending
                ? 'Uploading...'
                : !authToken
                  ? 'Please log in to upload files'
                  : selectedFile
                    ? `Upload ${selectedFile.name}`
                    : 'Select a file to upload'}
            </button>

            {/* Storage Status */}
            {storageStatus && storageStatus.storage_type && (
              <div className='rounded-lg border border-blue-300 bg-blue-50 p-4 dark:border-blue-700 dark:bg-blue-900/20'>
                <div className='flex items-center space-x-2'>
                  <div
                    className={`h-3 w-3 rounded-full ${
                      storageStatus.storage_type === 's3'
                        ? 'bg-green-500'
                        : 'bg-orange-500'
                    }`}
                  />
                  <p className='font-medium text-blue-700 dark:text-blue-300'>
                    Storage: {storageStatus.storage_type.toUpperCase()}
                  </p>
                </div>
                <p className='mt-1 text-sm text-blue-600 dark:text-blue-400'>
                  {storageStatus.message}
                </p>
              </div>
            )}
          </div>
        </section>

        {/* Files List Section */}
        <section className='rounded-lg bg-white p-6 shadow-md dark:bg-gray-800'>
          <div className='mb-4 flex items-center justify-between'>
            <h2 className='text-2xl font-semibold text-gray-900 dark:text-white'>
              📂 Uploaded Files
            </h2>
            <button
              onClick={() => {
                if (files && files.length > 0) {
                  toast.info('Refreshing files list...');
                }
              }}
              disabled={filesLoading}
              className='rounded-lg bg-gray-600 px-4 py-2 font-medium text-white transition-colors hover:bg-gray-700 disabled:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 dark:disabled:bg-gray-800'
            >
              {filesLoading ? 'Loading...' : 'Refresh'}
            </button>
          </div>

          {filesLoading ? (
            <div className='py-8 text-center'>
              <p className='text-gray-600 dark:text-gray-300'>
                Loading files...
              </p>
            </div>
          ) : !files || files.length === 0 ? (
            <div className='py-8 text-center'>
              <p className='text-gray-500 dark:text-gray-400'>
                {files === undefined
                  ? 'Please log in to view your files'
                  : 'No files uploaded yet. Upload one above!'}
              </p>
            </div>
          ) : (
            <div className='space-y-4'>
              <AnimatePresence>
                {files.map((file: any) => (
                  <motion.div
                    key={file.id}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    transition={{ duration: 0.3 }}
                    className='flex items-center justify-between rounded-lg border border-gray-200 p-4 dark:border-gray-700'
                  >
                    <div className='flex items-center space-x-4'>
                      <div className='flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900'>
                        <svg
                          className='h-6 w-6 text-blue-600 dark:text-blue-400'
                          fill='none'
                          stroke='currentColor'
                          viewBox='0 0 24 24'
                        >
                          <path
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            strokeWidth={2}
                            d='M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
                          />
                        </svg>
                      </div>
                      <div>
                        <p className='font-medium text-gray-900 dark:text-white'>
                          {file.file_name}
                        </p>
                        <p className='text-sm text-gray-500 dark:text-gray-400'>
                          {(file.file_size / 1024).toFixed(2)} KB •{' '}
                          {file.content_type} •{' '}
                          {file.storage_type.toUpperCase()}
                        </p>
                      </div>
                    </div>
                    <div className='flex items-center space-x-2'>
                      <button
                        onClick={() =>
                          handleFileDownload(file.id, file.file_name)
                        }
                        className='rounded bg-blue-600 px-3 py-1 text-sm font-medium text-white transition-colors hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-600'
                      >
                        Download
                      </button>
                      <button
                        onClick={() =>
                          handleFileDelete(file.id, file.file_name)
                        }
                        className='rounded bg-red-600 px-3 py-1 text-sm font-medium text-white transition-colors hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600'
                      >
                        Delete
                      </button>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          )}
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
                              search={{ tab: undefined }}
                              className='font-medium text-blue-600 hover:text-blue-800 hover:underline dark:text-blue-400 dark:hover:text-blue-300'
                            >
                              {user.name}
                            </Link>
                          </td>
                          <td className='px-4 py-2'>
                            <Link
                              to='/users/$userId'
                              params={{ userId: user.id.toString() }}
                              search={{ tab: undefined }}
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
                <li>POST /api/files/upload - Upload file</li>
                <li>GET /api/files - List files</li>
                <li>GET /api/files/:id/download - Download file</li>
                <li>DELETE /api/files/:id - Delete file</li>
              </ul>
            </div>
            <div className='rounded-lg bg-gray-100 p-4 dark:bg-gray-700'>
              <h3 className='mb-2 font-medium text-gray-900 dark:text-white'>
                Server Details
              </h3>
              <ul className='space-y-1 text-sm text-gray-600 dark:text-gray-300'>
                <li>Base URL: {API_BASE_URL}</li>
                <li>Database: PostgreSQL</li>
                <li>File Storage: S3 or Database</li>
                <li>Framework: Go + Chi Router</li>
                <li>CORS: Enabled for React dev server</li>
              </ul>
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
