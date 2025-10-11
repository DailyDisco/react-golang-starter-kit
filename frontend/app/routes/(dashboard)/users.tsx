import { createFileRoute, Link, Outlet } from '@tanstack/react-router';
import { Loader2, User as UserIcon, Users as UsersIcon } from 'lucide-react';

import { Button } from '../../components/ui/button';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '../../components/ui/card';
import { useUsers } from '../../hooks/queries/use-users';

export const Route = createFileRoute('/(dashboard)/users')({
  component: UsersPage,
});

function UsersPage() {
  // Server state - handled by Tanstack Query
  const { data: users, isLoading: usersLoading } = useUsers();

  if (usersLoading) {
    return (
      <div className='flex min-h-[400px] items-center justify-center'>
        <Loader2 className='h-8 w-8 animate-spin' />
      </div>
    );
  }

  return (
    <main className='bg-gray-50 px-4 py-12 dark:bg-gray-900'>
      <div className='mx-auto max-w-4xl'>
        {/* Header */}
        <header className='mb-8'>
          <div className='flex items-center gap-3'>
            <UsersIcon className='h-8 w-8' />
            <div>
              <h1 className='text-3xl font-bold text-gray-900 dark:text-white'>
                Users
              </h1>
              <p className='text-gray-600 dark:text-gray-300'>
                Manage user accounts
              </p>
            </div>
          </div>
        </header>

        {/* Users List */}
        <div className='space-y-4'>
          {users && users.length > 0 ? (
            users.map(user => (
              <Card key={user.id} className='shadow-md'>
                <CardHeader>
                  <CardTitle className='flex items-center justify-between'>
                    <div className='flex items-center gap-3'>
                      <div className='bg-muted flex h-10 w-10 items-center justify-center rounded-full'>
                        <UserIcon className='h-5 w-5' />
                      </div>
                      <div>
                        <div>{user.name}</div>
                        <div className='text-muted-foreground text-sm font-normal'>
                          {user.email}
                        </div>
                      </div>
                    </div>
                    <Link
                      to='/users/$userId'
                      params={{ userId: String(user.id) }}
                      search={{ tab: undefined }}
                    >
                      <Button variant='outline' size='sm'>
                        View Details
                      </Button>
                    </Link>
                  </CardTitle>
                </CardHeader>
              </Card>
            ))
          ) : (
            <Card className='shadow-md'>
              <CardContent className='py-12 text-center'>
                <UserIcon className='mx-auto mb-4 h-12 w-12 text-gray-400' />
                <h3 className='mb-2 text-lg font-medium text-gray-900 dark:text-white'>
                  No users found
                </h3>
                <p className='text-gray-600 dark:text-gray-300'>
                  There are no users to display
                </p>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Outlet for child routes like /users/$userId */}
        <Outlet />
      </div>
    </main>
  );
}
