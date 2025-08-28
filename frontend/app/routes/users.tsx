import { useParams, Link, useNavigate } from 'react-router';
import { useState, useEffect } from 'react';
import { useUsers } from '../hooks/use-users';
import { type User, updateUser } from '../lib/api';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import {
  Loader2,
  ArrowLeft,
  Save,
  Edit3,
  User as UserIcon,
  Lock,
} from 'lucide-react';
import { toast } from 'sonner';

export function meta({ params }: { params: { userId: string } }) {
  return [
    { title: `User ${params.userId} - React + Go Starter Kit` },
    { name: 'description', content: 'View and edit user details' },
  ];
}

const UserDetailPage = () => {
  const { userId } = useParams();
  const navigate = useNavigate();
  const { users, loading: usersLoading, editUser } = useUsers();

  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [formData, setFormData] = useState({ name: '', email: '' });

  useEffect(() => {
    if (users.length > 0 && userId) {
      const userIdNum = parseInt(userId, 10);
      if (!isNaN(userIdNum)) {
        const foundUser = users.find(u => u.id === userIdNum);
        if (foundUser) {
          setUser(foundUser);
          setFormData({ name: foundUser.name, email: foundUser.email });
        }
      }
    }
  }, [users, userId]);

  const handleSave = async () => {
    if (!user) return;

    setSaving(true);
    try {
      const updatedUser = await updateUser({
        ...user,
        name: formData.name.trim(),
        email: formData.email.trim(),
      });

      await editUser(updatedUser);
      setUser(updatedUser);
      setEditMode(false);
      toast.success('User updated successfully');
    } catch (error) {
      // Error handling is done in the hook
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    if (user) {
      setFormData({ name: user.name, email: user.email });
    }
    setEditMode(false);
  };

  if (usersLoading) {
    return (
      <div className='flex items-center justify-center min-h-[400px]'>
        <Loader2 className='h-8 w-8 animate-spin' />
      </div>
    );
  }

  if (!user) {
    return (
      <div className='text-center py-12'>
        <UserIcon className='h-12 w-12 mx-auto text-gray-400 mb-4' />
        <h3 className='text-lg font-medium text-gray-900 dark:text-white mb-2'>
          User not found
        </h3>
        <p className='text-gray-600 dark:text-gray-300 mb-4'>
          The user with ID {userId} could not be found
        </p>
        <Link to='/demo'>
          <Button>
            <ArrowLeft className='h-4 w-4 mr-2' />
            Back to Demo
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <main className='bg-gray-50 dark:bg-gray-900 py-12 px-4'>
      <div className='max-w-2xl mx-auto'>
        {/* Back Button */}
        <div className='mb-6'>
          <Link to='/demo'>
            <Button variant='outline' size='sm'>
              <ArrowLeft className='h-4 w-4 mr-2' />
              Back to Demo
            </Button>
          </Link>
        </div>

        {/* Header */}
        <header className='mb-8'>
          <h1 className='text-3xl font-bold text-gray-900 dark:text-white mb-2'>
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
                  <Edit3 className='h-4 w-4 mr-2' />
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
                    saving || !formData.name.trim() || !formData.email.trim()
                  }
                  className='flex-1'
                >
                  {saving ? (
                    <Loader2 className='h-4 w-4 mr-2 animate-spin' />
                  ) : (
                    <Save className='h-4 w-4 mr-2' />
                  )}
                  Save
                </Button>
                <Button
                  variant='outline'
                  onClick={handleCancel}
                  disabled={saving}
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
};

export default function UserDetail() {
  return <UserDetailPage />;
}
