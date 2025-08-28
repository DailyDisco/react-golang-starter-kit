import { ProtectedRoute } from '../components/auth/ProtectedRoute';
import { UserProfile } from '../components/auth/UserProfile';

export default function ProfilePage() {
    return (
        <ProtectedRoute>
            <div className="container mx-auto py-8 px-4">
                <UserProfile />
            </div>
        </ProtectedRoute>
    );
}
