import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { Loader2 } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all duration-200 ease-out transform-gpu active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
  {
    variants: {
      variant: {
        default:
          'bg-primary text-primary-foreground shadow-sm hover:bg-primary/90 hover:shadow-md hover:scale-[1.02] active:shadow-sm active:scale-[0.98]',
        destructive:
          'bg-destructive text-white shadow-sm hover:bg-destructive/90 hover:shadow-md hover:scale-[1.02] active:shadow-sm active:scale-[0.98] focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60',
        outline:
          'border bg-background shadow-sm hover:bg-accent hover:text-accent-foreground hover:shadow-md hover:scale-[1.02] active:shadow-sm active:scale-[0.98] dark:bg-input/30 dark:border-input dark:hover:bg-input/50',
        secondary:
          'bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80 hover:shadow-md hover:scale-[1.02] active:shadow-sm active:scale-[0.98]',
        ghost:
          'hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50 active:bg-accent/80',
        link: 'text-primary underline-offset-4 hover:underline active:text-primary/80',
        gradient:
          'bg-gradient-to-r from-primary via-primary/90 to-primary/80 text-primary-foreground shadow-md hover:shadow-lg hover:scale-[1.02] active:scale-[0.98] hover:from-primary/90 hover:via-primary/80 hover:to-primary/70',
        glow:
          'bg-primary text-primary-foreground shadow-md hover:shadow-[0_0_20px_oklch(0.55_0.18_250/0.4)] dark:hover:shadow-[0_0_20px_oklch(0.65_0.2_250/0.5)] hover:scale-[1.02] active:scale-[0.98]',
      },
      size: {
        default: 'h-9 px-4 py-2 has-[>svg]:px-3',
        sm: 'h-8 rounded-md gap-1.5 px-3 has-[>svg]:px-2.5',
        lg: 'h-10 rounded-md px-6 has-[>svg]:px-4',
        icon: 'size-9',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

interface ButtonProps
  extends React.ComponentProps<'button'>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  loading?: boolean;
}

function Button({
  className,
  variant,
  size,
  asChild = false,
  loading = false,
  disabled,
  children,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : 'button';

  return (
    <Comp
      data-slot='button'
      className={cn(buttonVariants({ variant, size, className }))}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? (
        <>
          <Loader2 className='size-4 animate-spin' />
          <span className='sr-only'>Loading</span>
          {children}
        </>
      ) : (
        children
      )}
    </Comp>
  );
}

export { Button, buttonVariants };
