import { AlertCircle } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';
import { Label } from './label';

interface FormFieldWrapperProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Label text for the field */
  label?: string;
  /** Unique ID to link label with input */
  htmlFor?: string;
  /** Help text shown below the input */
  hint?: string;
  /** Error message (takes precedence over hint) */
  error?: string;
  /** Whether the field is required */
  required?: boolean;
  /** Whether the field is disabled */
  disabled?: boolean;
}

/**
 * Wrapper component for form fields with label, hint, and error display.
 * Use this for simple forms without React Hook Form integration.
 *
 * @example
 * ```tsx
 * <FormFieldWrapper
 *   label="Email"
 *   htmlFor="email"
 *   hint="We'll never share your email"
 *   error={errors.email}
 *   required
 * >
 *   <Input id="email" type="email" />
 * </FormFieldWrapper>
 * ```
 */
function FormFieldWrapper({
  label,
  htmlFor,
  hint,
  error,
  required,
  disabled,
  children,
  className,
  ...props
}: FormFieldWrapperProps) {
  return (
    <div
      className={cn('space-y-2', disabled && 'opacity-50', className)}
      {...props}
    >
      {label && (
        <Label
          htmlFor={htmlFor}
          className={cn(error && 'text-destructive')}
        >
          {label}
          {required && (
            <span className="text-destructive ml-1" aria-hidden="true">
              *
            </span>
          )}
        </Label>
      )}
      {children}
      {hint && !error && (
        <p className="text-xs text-muted-foreground">{hint}</p>
      )}
      {error && (
        <p
          className="text-xs text-destructive flex items-center gap-1.5 animate-fade-in"
          role="alert"
        >
          <AlertCircle className="h-3 w-3 shrink-0" aria-hidden="true" />
          {error}
        </p>
      )}
    </div>
  );
}

interface FormSectionProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Section title */
  title?: string;
  /** Section description */
  description?: string;
}

/**
 * Group related form fields into a section
 *
 * @example
 * ```tsx
 * <FormSection title="Personal Information" description="Your basic profile details">
 *   <FormFieldWrapper label="Name">...</FormFieldWrapper>
 *   <FormFieldWrapper label="Email">...</FormFieldWrapper>
 * </FormSection>
 * ```
 */
function FormSection({
  title,
  description,
  children,
  className,
  ...props
}: FormSectionProps) {
  return (
    <div className={cn('space-y-4', className)} {...props}>
      {(title || description) && (
        <div className="space-y-1">
          {title && (
            <h3 className="text-lg font-medium leading-none">{title}</h3>
          )}
          {description && (
            <p className="text-sm text-muted-foreground">{description}</p>
          )}
        </div>
      )}
      <div className="space-y-4">{children}</div>
    </div>
  );
}

interface FormActionsProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Align actions to start, center, or end */
  align?: 'start' | 'center' | 'end' | 'between';
}

/**
 * Container for form action buttons (submit, cancel, etc.)
 *
 * @example
 * ```tsx
 * <FormActions align="end">
 *   <Button variant="outline">Cancel</Button>
 *   <Button type="submit">Save</Button>
 * </FormActions>
 * ```
 */
function FormActions({
  align = 'end',
  children,
  className,
  ...props
}: FormActionsProps) {
  const alignmentClasses = {
    start: 'justify-start',
    center: 'justify-center',
    end: 'justify-end',
    between: 'justify-between',
  };

  return (
    <div
      className={cn(
        'flex flex-wrap items-center gap-3 pt-4',
        alignmentClasses[align],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export { FormFieldWrapper, FormSection, FormActions };
