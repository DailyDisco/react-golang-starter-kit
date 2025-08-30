import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RegisterForm } from './RegisterForm';
import {
  renderWithProviders,
  createMockAuthStore,
  createMockMutation,
  createMockNavigate,
  createMockLocation,
  createMockUser,
} from '../../test/test-utils';

// Mock the hooks
const mockUseAuthStore = vi.fn();
const mockUseRegister = vi.fn();

vi.mock('../../stores/auth-store', () => ({
  useAuthStore: () => mockUseAuthStore(),
}));

vi.mock('../../hooks/mutations/use-auth-mutations', () => ({
  useRegister: () => mockUseRegister(),
}));

// Router context is provided by renderWithProviders

describe('RegisterForm', () => {
  let user: ReturnType<typeof userEvent.setup>;

  beforeEach(() => {
    user = userEvent.setup();
    vi.clearAllMocks();

    // Default mock implementations
    mockUseAuthStore.mockReturnValue(createMockAuthStore());
    mockUseRegister.mockReturnValue(createMockMutation());
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', () => {
    try {
      const { container } = renderWithProviders(<RegisterForm />);
      console.log('Container HTML:', container.innerHTML);
      expect(container.firstChild).toBeInTheDocument();
    } catch (error) {
      console.error('Rendering error:', error);
      throw error;
    }
  });

  it('renders a simple div', () => {
    const { container } = renderWithProviders(<div>Test</div>);
    expect(container.firstChild).toBeInTheDocument();
  });

  describe('Form Rendering', () => {
    it('renders all form fields correctly', () => {
      renderWithProviders(<RegisterForm />);

      expect(screen.getByLabelText(/full name/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: /create account/i })
      ).toBeInTheDocument();
    });

    it('renders form title and description', () => {
      renderWithProviders(<RegisterForm />);

      expect(screen.getByText('Create account')).toBeInTheDocument();
      expect(
        screen.getByText('Enter your information to create your account')
      ).toBeInTheDocument();
    });

    it('renders sign in link', () => {
      renderWithProviders(<RegisterForm />);

      const signInLink = screen.getByText('Sign in');
      expect(signInLink).toBeInTheDocument();
      expect(signInLink.closest('a')).toHaveAttribute('href', '/login');
    });
  });

  describe('Form Validation', () => {
    it('shows validation errors for empty form submission', async () => {
      renderWithProviders(<RegisterForm />);

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText('Name must be at least 2 characters')
        ).toBeInTheDocument();
        expect(
          screen.getByText('Please enter a valid email address')
        ).toBeInTheDocument();
        expect(
          screen.getByText('Password must be at least 8 characters')
        ).toBeInTheDocument();
      });
    });

    it('shows validation error for short name', async () => {
      renderWithProviders(<RegisterForm />);

      const nameInput = screen.getByLabelText(/full name/i);
      await user.type(nameInput, 'A');

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText('Name must be at least 2 characters')
        ).toBeInTheDocument();
      });
    });

    it('shows validation error for invalid email', async () => {
      renderWithProviders(<RegisterForm />);

      const emailInput = screen.getByLabelText(/email/i);
      await user.type(emailInput, 'invalid-email');

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText('Please enter a valid email address')
        ).toBeInTheDocument();
      });
    });

    it('shows validation error for short password', async () => {
      renderWithProviders(<RegisterForm />);

      const passwordInput = screen.getByLabelText(/^password$/i);
      await user.type(passwordInput, '1234567');

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText('Password must be at least 8 characters')
        ).toBeInTheDocument();
      });
    });

    it('shows validation error when passwords do not match', async () => {
      renderWithProviders(<RegisterForm />);

      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'differentpassword');

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText("Passwords don't match")).toBeInTheDocument();
      });
    });

    it('accepts valid form data', async () => {
      const mockMutation = createMockMutation();
      mockUseRegister.mockReturnValue(mockMutation);

      renderWithProviders(<RegisterForm />);

      const nameInput = screen.getByLabelText(/full name/i);
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

      await user.type(nameInput, 'John Doe');
      await user.type(emailInput, 'john@example.com');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockMutation.mutate).toHaveBeenCalledWith({
          name: 'John Doe',
          email: 'john@example.com',
          password: 'password123',
        });
      });
    });
  });

  describe('Password Visibility Toggle', () => {
    it('toggles password visibility for password field', async () => {
      renderWithProviders(<RegisterForm />);

      const passwordInput = screen.getByLabelText(/^password$/i);
      const toggleButton = passwordInput.parentElement?.querySelector('button');

      expect(passwordInput).toHaveAttribute('type', 'password');

      if (toggleButton) {
        await user.click(toggleButton);
        expect(passwordInput).toHaveAttribute('type', 'text');

        await user.click(toggleButton);
        expect(passwordInput).toHaveAttribute('type', 'password');
      }
    });

    it('toggles password visibility for confirm password field', async () => {
      renderWithProviders(<RegisterForm />);

      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);
      const toggleButton =
        confirmPasswordInput.parentElement?.querySelector('button');

      expect(confirmPasswordInput).toHaveAttribute('type', 'password');

      if (toggleButton) {
        await user.click(toggleButton);
        expect(confirmPasswordInput).toHaveAttribute('type', 'text');

        await user.click(toggleButton);
        expect(confirmPasswordInput).toHaveAttribute('type', 'password');
      }
    });
  });

  describe('Loading States', () => {
    it('disables form fields and button when mutation is pending', () => {
      const mockMutation = createMockMutation({ isPending: true });
      mockUseRegister.mockReturnValue(mockMutation);

      renderWithProviders(<RegisterForm />);

      const nameInput = screen.getByLabelText(/full name/i);
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);
      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });

      expect(nameInput).toBeDisabled();
      expect(emailInput).toBeDisabled();
      expect(passwordInput).toBeDisabled();
      expect(confirmPasswordInput).toBeDisabled();
      expect(submitButton).toBeDisabled();
    });

    it('shows loading spinner when mutation is pending', () => {
      const mockMutation = createMockMutation({ isPending: true });
      mockUseRegister.mockReturnValue(mockMutation);

      renderWithProviders(<RegisterForm />);

      const loadingSpinner = document.querySelector('.animate-spin');
      expect(loadingSpinner).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('displays error message on registration failure', async () => {
      const mockMutation = createMockMutation();
      mockUseRegister.mockReturnValue(mockMutation);

      renderWithProviders(<RegisterForm />);

      // Fill out the form
      const nameInput = screen.getByLabelText(/full name/i);
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

      await user.type(nameInput, 'John Doe');
      await user.type(emailInput, 'john@example.com');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');

      // Submit the form
      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      // Simulate error response
      const error = new Error('Email already exists');
      mockMutation.mutate.mock.calls[0][1].onError(error);

      await waitFor(() => {
        expect(screen.getByText('Email already exists')).toBeInTheDocument();
      });
    });

    it('clears previous error on new submission', async () => {
      const mockMutation = createMockMutation();
      mockUseRegister.mockReturnValue(mockMutation);

      renderWithProviders(<RegisterForm />);

      // First submission with error
      const nameInput = screen.getByLabelText(/full name/i);
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

      await user.type(nameInput, 'John Doe');
      await user.type(emailInput, 'john@example.com');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');

      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      // Simulate error
      mockMutation.mutate.mock.calls[0][1].onError(new Error('First error'));

      await waitFor(() => {
        expect(screen.getByText('First error')).toBeInTheDocument();
      });

      // Second submission should clear the error
      await user.click(submitButton);
      expect(screen.queryByText('First error')).not.toBeInTheDocument();
    });
  });

  describe('Success Handling', () => {
    it('navigates to home page on successful registration', async () => {
      const mockMutation = createMockMutation();
      const mockNavigate = createMockNavigate();
      mockUseRegister.mockReturnValue(mockMutation);
      mockUseNavigate.mockReturnValue(mockNavigate);

      renderWithProviders(<RegisterForm />);

      // Fill out the form
      const nameInput = screen.getByLabelText(/full name/i);
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

      await user.type(nameInput, 'John Doe');
      await user.type(emailInput, 'john@example.com');
      await user.type(passwordInput, 'password123');
      await user.type(confirmPasswordInput, 'password123');

      // Submit the form
      const submitButton = screen.getByRole('button', {
        name: /create account/i,
      });
      await user.click(submitButton);

      // Simulate success response
      const mockUser = createMockUser({
        name: 'John Doe',
        email: 'john@example.com',
      });
      mockMutation.mutate.mock.calls[0][1].onSuccess({
        user: mockUser,
        token: 'mock-token',
      });

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith({
          to: '/',
          search: undefined,
        });
      });
    });
  });

  describe('Accessibility', () => {
    it('has proper ARIA labels and roles', () => {
      renderWithProviders(<RegisterForm />);

      const form = screen.getByRole('form');
      expect(form).toBeInTheDocument();

      const inputs = screen.getAllByRole('textbox');
      expect(inputs.length).toBeGreaterThan(0);

      const button = screen.getByRole('button', { name: /create account/i });
      expect(button).toBeInTheDocument();
    });

    it('associates labels with inputs correctly', () => {
      renderWithProviders(<RegisterForm />);

      const nameInput = screen.getByLabelText(/full name/i);
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const confirmPasswordInput = screen.getByLabelText(/confirm password/i);

      expect(nameInput).toHaveAttribute('id', 'name');
      expect(emailInput).toHaveAttribute('id', 'email');
      expect(passwordInput).toHaveAttribute('id', 'password');
      expect(confirmPasswordInput).toHaveAttribute('id', 'confirmPassword');
    });
  });
});
