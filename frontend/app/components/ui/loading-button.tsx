import { Loader2 } from 'lucide-react';
import * as React from 'react';

import { Button, buttonVariants } from './button';
import type { VariantProps } from 'class-variance-authority';

interface LoadingButtonProps
  extends React.ComponentProps<'button'>,
    VariantProps<typeof buttonVariants> {
  /** Whether the button is in a loading state */
  isLoading?: boolean;
  /** Text to show while loading (defaults to children) */
  loadingText?: string;
  /** Custom loading spinner (defaults to Loader2) */
  loadingIcon?: React.ReactNode;
  /** Position of the loading icon */
  iconPosition?: 'left' | 'right';
  asChild?: boolean;
}

/**
 * Button with built-in loading state
 *
 * @example
 * ```tsx
 * <LoadingButton
 *   isLoading={mutation.isPending}
 *   loadingText="Saving..."
 * >
 *   Save Changes
 * </LoadingButton>
 * ```
 */
function LoadingButton({
  children,
  isLoading = false,
  loadingText,
  loadingIcon,
  iconPosition = 'left',
  disabled,
  ...props
}: LoadingButtonProps) {
  const spinner = loadingIcon ?? (
    <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
  );

  const content = isLoading ? (
    <>
      {iconPosition === 'left' && spinner}
      <span>{loadingText ?? children}</span>
      {iconPosition === 'right' && spinner}
    </>
  ) : (
    children
  );

  return (
    <Button
      disabled={disabled || isLoading}
      aria-busy={isLoading}
      aria-disabled={disabled || isLoading}
      {...props}
    >
      {content}
    </Button>
  );
}

export { LoadingButton };
