import React, { useState, useEffect } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import type { User } from '../../services';

interface UserFormProps {
  onSubmit: (name: string, email: string, id?: number) => void;
  initialData?: User | null;
  isLoading?: boolean;
}

export const UserForm: React.FC<UserFormProps> = ({
  onSubmit,
  initialData,
  isLoading,
}) => {
  const [name, setName] = useState(initialData?.name || '');
  const [email, setEmail] = useState(initialData?.email || '');

  useEffect(() => {
    if (initialData) {
      setName(initialData.name);
      setEmail(initialData.email);
    } else {
      setName('');
      setEmail('');
    }
  }, [initialData]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(name, email, initialData?.id);
    if (!initialData) {
      // Clear form only for new user creation
      setName('');
      setEmail('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className='flex flex-col gap-3'>
      <Input
        type='text'
        placeholder='Name'
        value={name}
        onChange={e => setName(e.target.value)}
        required
        disabled={isLoading}
      />
      <Input
        type='email'
        placeholder='Email'
        value={email}
        onChange={e => setEmail(e.target.value)}
        required
        disabled={isLoading}
      />
      <Button type='submit' disabled={isLoading}>
        {initialData ? 'Update User' : 'Create User'}
      </Button>
    </form>
  );
};
