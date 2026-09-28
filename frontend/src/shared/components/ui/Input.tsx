import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react';
import { cn } from '../../utils';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, leftIcon, rightIcon, id, ...props }, ref) => {
    const inputId = id ?? `input-${label.toLowerCase().replace(/\s+/g, '-')}`;

    return (
      <div className="flex flex-col gap-1 w-full">
        <label htmlFor={inputId} className="text-sm font-medium text-[var(--color-foreground)]">
          {label}
        </label>

        <div className="relative flex items-center">
          {leftIcon && (
            <span className="absolute left-3 text-[var(--color-muted)]" aria-hidden="true">
              {leftIcon}
            </span>
          )}

          <input
            ref={ref}
            id={inputId}
            aria-describedby={error ? `${inputId}-error` : undefined}
            aria-invalid={error ? 'true' : 'false'}
            className={cn(
              'w-full rounded-lg border px-4 py-2 text-base',
              'min-h-[44px]',
              'border-[var(--color-border)] bg-[var(--color-background)]',
              'placeholder:text-[var(--color-muted)]',
              'focus-visible:outline-none focus-visible:ring-2',
              'focus-visible:ring-[var(--color-primary)] focus-visible:ring-offset-2',
              'transition-colors',
              error && 'border-[var(--color-error)] focus-visible:ring-[var(--color-error)]',
              leftIcon && 'pl-10',
              rightIcon && 'pr-10',
              className
            )}
            {...props}
          />

          {rightIcon && <span className="absolute right-3 text-[var(--color-muted)]">{rightIcon}</span>}
        </div>

        {error && (
          <p id={`${inputId}-error`} role="alert" className="text-sm text-[var(--color-error)] flex items-center gap-1">
            <span aria-hidden="true">⚠</span>
            {error}
          </p>
        )}
      </div>
    );
  }
);
Input.displayName = 'Input';
