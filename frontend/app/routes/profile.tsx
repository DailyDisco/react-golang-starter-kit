import { createFileRoute } from '@tanstack/react-router';
import { ProtectedRoute } from '../components/auth/ProtectedRoute';
import { UserProfile } from '../components/auth/UserProfile';

export const Route = createFileRoute('/profile')({
  component: ProfilePage,
});

function ProfilePage() {
  return (
    <ProtectedRoute>
      <div className='container mx-auto py-8 px-4'>
        <UserProfile />
      </div>
    </ProtectedRoute>
  );
}
