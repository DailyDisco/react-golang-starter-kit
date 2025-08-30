import { useLocation, useNavigate } from '@tanstack/react-router';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import {
  createMockAuthStore,
  createMockMutation,
  createMockUser,
  renderWithProviders,
} from '../../test/test-utils';
import { LoginForm } from './LoginForm';

// Mock the hooks
const mockUseAuthStore = vi.fn();
const mockUseLogin = vi.fn();

vi.mock('../../stores/auth-store', () => ({
  useAuthStore: () => mockUseAuthStore(),
}));

vi.mock('../../hooks/mutations/use-auth-mutations', () => ({
  useLogin: () => mockUseLogin(),
}));

describe('LoginForm', () => {
  let user: ReturnType<typeof userEvent.setup>;

  beforeEach(() => {
    user = userEvent.setup();
    vi.clearAllMocks();

    // Default mock implementations
    mockUseAuthStore.mockReturnValue(createMockAuthStore());
    mockUseLogin.mockReturnValue(createMockMutation());
    useNavigate.mockReturnValue(vi.fn());
    useLocation.mockReturnValue({
      pathname: '/',
      search: {}, // Ensure search is an object
      hash: '',
      state: null,
      key: 'default',
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    // No need to call mockRestore on globally mocked functions.
    // useNavigate.mockRestore();
    // useLocation.mockRestore();
  });

  it('renders LoginForm with providers', () => {
    try {
      const { container } = renderWithProviders(<LoginForm />);
      expect(container.firstChild).toBeInTheDocument();
    } catch (error) {
      console.error('LoginForm with providers rendering error:', error);
      throw error;
    }
  });

  describe('Form Rendering', () => {
    it('renders all form fields correctly', () => {
      renderWithProviders(<LoginForm />);

      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: /sign in/i })
      ).toBeInTheDocument();
    });

    it('renders form title and description', () => {
      renderWithProviders(<LoginForm />);

      expect(screen.getByText('Sign in')).toBeInTheDocument();
      expect(
        screen.getByText(
          'Enter your email and password to sign in to your account'
        )
      ).toBeInTheDocument();
    });

    it('renders sign up link', () => {
      renderWithProviders(<LoginForm />);

      const signUpLink = screen.getByText('Sign up');
      expect(signUpLink).toBeInTheDocument();
      expect(signUpLink.closest('a')).toHaveAttribute('href', '/register');
    });
  });

  describe('Form Validation', () => {
    it('shows validation errors for empty form submission', async () => {
      renderWithProviders(<LoginForm />);

      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText('Please enter a valid email address')
        ).toBeInTheDocument();
        expect(screen.getByText('Password is required')).toBeInTheDocument();
      });
    });

    it('shows validation error for invalid email', async () => {
      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      await user.type(emailInput, 'invalid-email');

      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(
          screen.getByText('Please enter a valid email address')
        ).toBeInTheDocument();
      });
    });

    it('shows validation error for empty password', async () => {
      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      await user.type(emailInput, 'test@example.com');

      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Password is required')).toBeInTheDocument();
      });
    });

    it('accepts valid form data', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);

      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');

      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockMutation.mutate).toHaveBeenCalledWith({
          email: 'test@example.com',
          password: 'password123',
        });
      });
    });
  });

  describe('Password Visibility Toggle', () => {
    it('toggles password visibility', async () => {
      renderWithProviders(<LoginForm />);

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
  });

  describe('Loading States', () => {
    it('disables form fields and button when mutation is pending', () => {
      const mockMutation = createMockMutation({ isPending: true });
      mockUseLogin.mockReturnValue(mockMutation);

      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      expect(emailInput).toBeDisabled();
      expect(passwordInput).toBeDisabled();
      expect(submitButton).toBeDisabled();
    });

    it('shows loading spinner when mutation is pending', () => {
      const mockMutation = createMockMutation({ isPending: true });
      mockUseLogin.mockReturnValue(mockMutation);

      renderWithProviders(<LoginForm />);

      const loadingSpinner = document.querySelector('.animate-spin');
      expect(loadingSpinner).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('displays error message on login failure', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);

      renderWithProviders(<LoginForm />);

      // Fill out the form
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'wrongpassword');

      // Submit the form
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      // Simulate error response
      const error = new Error('Invalid credentials');
      mockMutation.mutate.mock.calls[0][1].onError(error);

      await waitFor(() => {
        expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
      });
    });

    it('clears previous error on new submission', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);

      renderWithProviders(<LoginForm />);

      // First submission with error
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');

      const submitButton = screen.getByRole('button', { name: /sign in/i });
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

  describe('Success Handling and Navigation', () => {
    it('navigates to home page when no redirect location is specified', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);
      useNavigate.mockReturnValue(vi.fn());
      useLocation.mockReturnValue({
        pathname: '/',
        search: '',
        hash: '',
        state: null,
        key: 'default',
      });

      renderWithProviders(<LoginForm />);

      // Fill out the form
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');

      // Submit the form
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      // Simulate success response
      const mockUser = createMockUser({ email: 'test@example.com' });
      mockMutation.mutate.mock.calls[0][1].onSuccess({
        user: mockUser,
        token: 'mock-token',
      });

      await waitFor(() => {
        expect(useNavigate).toHaveBeenCalledWith({
          to: '/',
          replace: true,
        });
      });
    });

    it('navigates to intended page when redirect location is specified', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);
      useNavigate.mockReturnValue(vi.fn());
      useLocation.mockReturnValue({
        pathname: '/',
        search: '',
        hash: '',
        state: { from: { pathname: '/dashboard' } },
        key: 'default',
      });

      renderWithProviders(<LoginForm />);

      // Fill out the form
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');

      // Submit the form
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      // Simulate success response
      const mockUser = createMockUser({ email: 'test@example.com' });
      mockMutation.mutate.mock.calls[0][1].onSuccess({
        user: mockUser,
        token: 'mock-token',
      });

      await waitFor(() => {
        expect(useNavigate).toHaveBeenCalledWith({
          to: '/dashboard',
          replace: true,
        });
      });
    });

    it('navigates to home page when redirect location has no pathname', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);
      useNavigate.mockReturnValue(vi.fn());
      useLocation.mockReturnValue({
        pathname: '/',
        search: '',
        hash: '',
        state: { from: {} },
        key: 'default',
      });

      renderWithProviders(<LoginForm />);

      // Fill out the form
      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      await user.type(emailInput, 'test@example.com');
      await user.type(passwordInput, 'password123');

      // Submit the form
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      // Simulate success response
      const mockUser = createMockUser({ email: 'test@example.com' });
      mockMutation.mutate.mock.calls[0][1].onSuccess({
        user: mockUser,
        token: 'mock-token',
      });

      await waitFor(() => {
        expect(useNavigate).toHaveBeenCalledWith({
          to: '/',
          replace: true,
        });
      });
    });
  });

  describe('Accessibility', () => {
    it('has proper ARIA labels and roles', () => {
      renderWithProviders(<LoginForm />);

      const form = screen.getByRole('form');
      expect(form).toBeInTheDocument();

      const inputs = screen.getAllByRole('textbox');
      expect(inputs.length).toBeGreaterThan(0);

      const button = screen.getByRole('button', { name: /sign in/i });
      expect(button).toBeInTheDocument();
    });

    it('associates labels with inputs correctly', () => {
      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);

      expect(emailInput).toHaveAttribute('id', 'email');
      expect(passwordInput).toHaveAttribute('id', 'password');
    });
  });

  describe('User Experience', () => {
    it('maintains focus on email input initially', () => {
      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      expect(emailInput).toHaveFocus();
    });

    it('allows keyboard navigation through form fields', async () => {
      renderWithProviders(<LoginForm />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/^password$/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      emailInput.focus();
      expect(emailInput).toHaveFocus();

      await user.tab();
      expect(passwordInput).toHaveFocus();

      await user.tab();
      expect(submitButton).toHaveFocus();
    });

    it('prevents default form submission when validation fails', async () => {
      const mockMutation = createMockMutation();
      mockUseLogin.mockReturnValue(mockMutation);

      renderWithProviders(<LoginForm />);

      const form = screen.getByRole('form');
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      // Try to submit without filling required fields
      await user.click(submitButton);

      // Mutation should not be called due to validation
      expect(mockMutation.mutate).not.toHaveBeenCalled();
    });
  });
});
