import { AlertCircle, Check } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

interface FloatingLabelInputProps
  extends Omit<React.ComponentProps<'input'>, 'placeholder'> {
  label: string;
  error?: string;
  success?: boolean;
  hint?: string;
}

function FloatingLabelInput({
  className,
  type = 'text',
  label,
  error,
  success,
  hint,
  id,
  ...props
}: FloatingLabelInputProps) {
  const [isFocused, setIsFocused] = React.useState(false);
  const [hasValue, setHasValue] = React.useState(false);
  const generatedId = React.useId();
  const inputId = id ?? generatedId;
  const hintId = `${inputId}-hint`;
  const errorId = `${inputId}-error`;

  const isFloating = isFocused || hasValue;

  const handleFocus = (e: React.FocusEvent<HTMLInputElement>) => {
    setIsFocused(true);
    props.onFocus?.(e);
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    setIsFocused(false);
    props.onBlur?.(e);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setHasValue(e.target.value.length > 0);
    props.onChange?.(e);
  };

  React.useEffect(() => {
    if (props.value !== undefined) {
      setHasValue(String(props.value).length > 0);
    }
    if (props.defaultValue !== undefined) {
      setHasValue(String(props.defaultValue).length > 0);
    }
  }, [props.value, props.defaultValue]);

  return (
    <div className='relative'>
      <div className='relative'>
        <input
          type={type}
          id={inputId}
          data-slot='floating-input'
          aria-invalid={!!error}
          aria-describedby={
            error ? errorId : hint ? hintId : undefined
          }
          className={cn(
            'peer flex h-12 w-full min-w-0 rounded-md border bg-transparent px-3 pt-5 pb-1 text-base shadow-xs transition-all duration-200 outline-none md:text-sm',
            'file:text-foreground placeholder:text-transparent selection:bg-primary selection:text-primary-foreground',
            'dark:bg-input/30 border-input',
            'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
            'disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50',
            error &&
              'border-destructive ring-destructive/20 dark:ring-destructive/40 animate-shake',
            success &&
              'border-success ring-success/20 dark:ring-success/30',
            className
          )}
          placeholder={label}
          onFocus={handleFocus}
          onBlur={handleBlur}
          onChange={handleChange}
          {...props}
        />
        <label
          htmlFor={inputId}
          className={cn(
            'text-muted-foreground pointer-events-none absolute left-3 transition-all duration-200 ease-out',
            isFloating
              ? 'top-1.5 text-xs font-medium'
              : 'top-1/2 -translate-y-1/2 text-base md:text-sm',
            isFocused && 'text-primary',
            error && 'text-destructive',
            success && 'text-success'
          )}
        >
          {label}
        </label>

        {/* Success indicator */}
        {success && !error && (
          <div className='absolute right-3 top-1/2 -translate-y-1/2'>
            <Check className='text-success h-4 w-4 animate-scale-in' />
          </div>
        )}

        {/* Error indicator */}
        {error && (
          <div className='absolute right-3 top-1/2 -translate-y-1/2'>
            <AlertCircle className='text-destructive h-4 w-4' />
          </div>
        )}
      </div>

      {/* Hint text */}
      {hint && !error && (
        <p
          id={hintId}
          className='text-muted-foreground mt-1.5 text-xs'
        >
          {hint}
        </p>
      )}

      {/* Error message */}
      {error && (
        <p
          id={errorId}
          role='alert'
          className='text-destructive mt-1.5 flex items-center gap-1 text-xs animate-fade-in-up'
        >
          {error}
        </p>
      )}
    </div>
  );
}

export { FloatingLabelInput };
