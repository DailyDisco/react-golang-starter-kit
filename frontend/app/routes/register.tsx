import { createFileRoute } from '@tanstack/react-router';
import { RegisterForm } from '../components/auth/RegisterForm';

export const Route = createFileRoute('/register')({
  component: RegisterPage,
});

function RegisterPage() {
  return <RegisterForm />;
}
