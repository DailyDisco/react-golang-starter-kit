import { Check } from 'lucide-react';
import * as React from 'react';

import { cn } from '@/lib/utils';

interface Step {
  label: string;
  description?: string;
}

interface FormProgressProps {
  steps: Step[];
  currentStep: number;
  className?: string;
}

function FormProgress({ steps, currentStep, className }: FormProgressProps) {
  return (
    <nav aria-label='Progress' className={className}>
      <ol className='flex items-center'>
        {steps.map((step, index) => {
          const isCompleted = index < currentStep;
          const isCurrent = index === currentStep;
          const isPending = index > currentStep;

          return (
            <li
              key={step.label}
              className={cn(
                'relative flex-1',
                index !== steps.length - 1 && 'pr-8 sm:pr-20'
              )}
            >
              {/* Connector line */}
              {index !== steps.length - 1 && (
                <div
                  className='absolute top-4 left-0 -right-4 sm:-right-10 h-0.5'
                  aria-hidden='true'
                >
                  <div
                    className={cn(
                      'h-full w-full transition-colors duration-300',
                      isCompleted ? 'bg-primary' : 'bg-border'
                    )}
                  />
                </div>
              )}

              <div className='group relative flex flex-col items-center'>
                {/* Step circle */}
                <span
                  className={cn(
                    'relative z-10 flex h-8 w-8 items-center justify-center rounded-full border-2 transition-all duration-300',
                    isCompleted &&
                      'border-primary bg-primary text-primary-foreground',
                    isCurrent &&
                      'border-primary bg-background text-primary shadow-md shadow-primary/20',
                    isPending &&
                      'border-muted-foreground/30 bg-background text-muted-foreground'
                  )}
                >
                  {isCompleted ? (
                    <Check className='h-4 w-4 animate-scale-in' />
                  ) : (
                    <span
                      className={cn(
                        'text-sm font-medium',
                        isCurrent && 'font-semibold'
                      )}
                    >
                      {index + 1}
                    </span>
                  )}
                </span>

                {/* Step label */}
                <span
                  className={cn(
                    'mt-2 text-center text-xs font-medium transition-colors duration-200',
                    isCompleted && 'text-primary',
                    isCurrent && 'text-foreground',
                    isPending && 'text-muted-foreground'
                  )}
                >
                  {step.label}
                </span>

                {/* Step description (optional) */}
                {step.description && (
                  <span
                    className={cn(
                      'mt-0.5 text-center text-xs transition-colors duration-200',
                      isCompleted || isCurrent
                        ? 'text-muted-foreground'
                        : 'text-muted-foreground/60'
                    )}
                  >
                    {step.description}
                  </span>
                )}
              </div>
            </li>
          );
        })}
      </ol>
    </nav>
  );
}

interface FormProgressBarProps {
  steps: number;
  currentStep: number;
  className?: string;
}

function FormProgressBar({
  steps,
  currentStep,
  className,
}: FormProgressBarProps) {
  const progress = (currentStep / (steps - 1)) * 100;

  return (
    <div className={cn('space-y-2', className)}>
      <div className='flex justify-between text-xs text-muted-foreground'>
        <span>
          Step {currentStep + 1} of {steps}
        </span>
        <span>{Math.round(progress)}% complete</span>
      </div>
      <div className='bg-secondary h-2 overflow-hidden rounded-full'>
        <div
          className='bg-primary h-full rounded-full transition-all duration-500 ease-out'
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  );
}

export { FormProgress, FormProgressBar };
