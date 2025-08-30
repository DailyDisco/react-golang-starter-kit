import { zodResolver } from '@hookform/resolvers/zod';
import { Calendar, Edit3, Loader2, Mail, Save, User, X } from 'lucide-react';
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { useAuth } from '../../hooks/auth/useAuth';
import { Alert, AlertDescription } from '../ui/alert';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../ui/card';
import { Input } from '../ui/input';
import { Label } from '../ui/label';

const profileSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Please enter a valid email address'),
});

type ProfileFormData = z.infer<typeof profileSchema>;

export function UserProfile() {
  const { user, updateUser, isLoading } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: user?.name || '',
      email: user?.email || '',
    },
  });

  const onSubmit = async (data: ProfileFormData) => {
    try {
      setError(null);
      await updateUser(data);
      setSuccess('Profile updated successfully!');
      setIsEditing(false);
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Update failed');
    }
  };

  const handleCancel = () => {
    reset({
      name: user?.name || '',
      email: user?.email || '',
    });
    setIsEditing(false);
    setError(null);
  };

  if (!user) {
    return (
      <div className='flex items-center justify-center p-8'>
        <Loader2 className='h-8 w-8 animate-spin' />
      </div>
    );
  }

  return (
    <Card className='mx-auto w-full max-w-2xl'>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <User className='h-5 w-5' />
          User Profile
        </CardTitle>
        <CardDescription>
          View and manage your account information
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-6'>
        {error && (
          <Alert variant='destructive'>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {success && (
          <Alert>
            <AlertDescription>{success}</AlertDescription>
          </Alert>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className='space-y-4'>
          <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
            <div className='space-y-2'>
              <Label htmlFor='name' className='flex items-center gap-2'>
                <User className='h-4 w-4' />
                Full Name
              </Label>
              {isEditing ? (
                <>
                  <Input id='name' {...register('name')} disabled={isLoading} />
                  {errors.name && (
                    <p className='text-sm text-red-500'>
                      {errors.name.message}
                    </p>
                  )}
                </>
              ) : (
                <p className='rounded-md border bg-gray-50 p-2 text-sm text-gray-900'>
                  {user.name}
                </p>
              )}
            </div>

            <div className='space-y-2'>
              <Label htmlFor='email' className='flex items-center gap-2'>
                <Mail className='h-4 w-4' />
                Email Address
              </Label>
              {isEditing ? (
                <>
                  <Input
                    id='email'
                    type='email'
                    {...register('email')}
                    disabled={isLoading}
                  />
                  {errors.email && (
                    <p className='text-sm text-red-500'>
                      {errors.email.message}
                    </p>
                  )}
                </>
              ) : (
                <div className='flex items-center gap-2 rounded-md border bg-gray-50 p-2'>
                  <span className='text-sm text-gray-900'>{user.email}</span>
                  {user.email_verified && (
                    <Badge variant='secondary' className='text-xs'>
                      Verified
                    </Badge>
                  )}
                </div>
              )}
            </div>
          </div>

          <div className='grid grid-cols-1 gap-4 md:grid-cols-3'>
            <div className='space-y-2'>
              <Label className='flex items-center gap-2'>
                <Calendar className='h-4 w-4' />
                Member Since
              </Label>
              <p className='rounded-md border bg-gray-50 p-2 text-sm text-gray-900'>
                {new Date(user.created_at).toLocaleDateString()}
              </p>
            </div>

            <div className='space-y-2'>
              <Label>Status</Label>
              <div className='rounded-md border bg-gray-50 p-2'>
                <Badge variant={user.is_active ? 'default' : 'destructive'}>
                  {user.is_active ? 'Active' : 'Inactive'}
                </Badge>
              </div>
            </div>

            <div className='space-y-2'>
              <Label>Email Status</Label>
              <div className='rounded-md border bg-gray-50 p-2'>
                <Badge variant={user.email_verified ? 'default' : 'secondary'}>
                  {user.email_verified ? 'Verified' : 'Unverified'}
                </Badge>
              </div>
            </div>
          </div>

          <div className='flex gap-2'>
            {isEditing ? (
              <>
                <Button type='submit' disabled={isLoading}>
                  {isLoading && (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  )}
                  <Save className='mr-2 h-4 w-4' />
                  Save Changes
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  onClick={handleCancel}
                  disabled={isLoading}
                >
                  <X className='mr-2 h-4 w-4' />
                  Cancel
                </Button>
              </>
            ) : (
              <Button
                type='button'
                variant='outline'
                onClick={() => setIsEditing(true)}
              >
                <Edit3 className='mr-2 h-4 w-4' />
                Edit Profile
              </Button>
            )}
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
