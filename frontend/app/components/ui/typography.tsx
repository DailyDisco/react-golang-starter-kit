import { cn } from '@/lib/utils';
import * as React from 'react';

/**
 * Typography components for consistent text styling across the application.
 * Use these components to maintain visual hierarchy and readability.
 */

// ==========================================================================
// PAGE-LEVEL TYPOGRAPHY
// ==========================================================================

interface PageTitleProps extends React.HTMLAttributes<HTMLHeadingElement> {
  /** Optional subtitle displayed below the title */
  subtitle?: string;
}

/**
 * Primary page heading - use once per page
 *
 * @example
 * ```tsx
 * <PageTitle subtitle="Manage your account settings">
 *   Settings
 * </PageTitle>
 * ```
 */
function PageTitle({
  className,
  children,
  subtitle,
  ...props
}: PageTitleProps) {
  return (
    <div className="animate-fade-in">
      <h1
        className={cn(
          'text-3xl font-bold tracking-tight text-foreground md:text-4xl',
          className
        )}
        {...props}
      >
        {children}
      </h1>
      {subtitle && (
        <p className="mt-2 text-lg text-muted-foreground animation-delay-100 animate-fade-in">
          {subtitle}
        </p>
      )}
    </div>
  );
}

/**
 * Standalone page description paragraph
 */
function PageDescription({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn(
        'text-lg text-muted-foreground animate-fade-in animation-delay-100',
        className
      )}
      {...props}
    >
      {children}
    </p>
  );
}

// ==========================================================================
// SECTION-LEVEL TYPOGRAPHY
// ==========================================================================

interface SectionTitleProps extends React.HTMLAttributes<HTMLHeadingElement> {
  /** Render as different heading level */
  as?: 'h2' | 'h3' | 'h4';
}

/**
 * Section heading within a page
 *
 * @example
 * ```tsx
 * <SectionTitle>Personal Information</SectionTitle>
 * ```
 */
function SectionTitle({
  className,
  children,
  as: Component = 'h2',
  ...props
}: SectionTitleProps) {
  return (
    <Component
      className={cn(
        'text-xl font-semibold tracking-tight text-foreground md:text-2xl',
        className
      )}
      {...props}
    >
      {children}
    </Component>
  );
}

/**
 * Description text for a section
 */
function SectionDescription({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn('text-sm text-muted-foreground', className)}
      {...props}
    >
      {children}
    </p>
  );
}

// ==========================================================================
// CONTENT TYPOGRAPHY
// ==========================================================================

/**
 * Card or block title
 */
function CardTitle({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn(
        'text-lg font-semibold leading-none tracking-tight text-foreground',
        className
      )}
      {...props}
    >
      {children}
    </h3>
  );
}

/**
 * Card or block description
 */
function CardDescription({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn('text-sm text-muted-foreground', className)}
      {...props}
    >
      {children}
    </p>
  );
}

// ==========================================================================
// UTILITY TYPOGRAPHY
// ==========================================================================

/**
 * Caption/label text - small and uppercase
 *
 * @example
 * ```tsx
 * <Caption>Last updated</Caption>
 * ```
 */
function Caption({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        'text-xs font-medium uppercase tracking-wider text-muted-foreground',
        className
      )}
      {...props}
    >
      {children}
    </span>
  );
}

/**
 * Inline code styling
 */
function InlineCode({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLElement>) {
  return (
    <code
      className={cn(
        'relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm',
        className
      )}
      {...props}
    >
      {children}
    </code>
  );
}

/**
 * Lead paragraph - larger introductory text
 */
function Lead({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn('text-xl text-muted-foreground', className)}
      {...props}
    >
      {children}
    </p>
  );
}

/**
 * Muted/secondary text
 */
function Muted({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p
      className={cn('text-sm text-muted-foreground', className)}
      {...props}
    >
      {children}
    </p>
  );
}

/**
 * Large text for emphasis
 */
function Large({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('text-lg font-semibold', className)}
      {...props}
    >
      {children}
    </div>
  );
}

/**
 * Small text
 */
function Small({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLElement>) {
  return (
    <small
      className={cn('text-sm font-medium leading-none', className)}
      {...props}
    >
      {children}
    </small>
  );
}

export {
  PageTitle,
  PageDescription,
  SectionTitle,
  SectionDescription,
  CardTitle,
  CardDescription,
  Caption,
  InlineCode,
  Lead,
  Muted,
  Large,
  Small,
};
