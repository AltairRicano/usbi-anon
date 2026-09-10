import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

// Clases compartidas entre <Button> y <LinkButton> (mismo look, distinto
// elemento raíz) para que ambos se mantengan visualmente idénticos sin
// duplicar el mapeo de variant/size en dos archivos.
export function buttonClasses(variant: ButtonVariant, size: ButtonSize, className?: string) {
  return cn(
    'inline-flex items-center justify-center gap-2 rounded-xl font-medium',
    'transition-[transform,box-shadow] hover:scale-[1.02] active:scale-95',
    'disabled:pointer-events-none disabled:opacity-50',
    'min-h-[44px] min-w-[44px]',
    'focus-visible:outline-none focus-visible:ring-2',
    'focus-visible:ring-[--color-primary] focus-visible:ring-offset-2',
    {
      'bg-[var(--color-primary)] text-[var(--color-primary-foreground)] shadow-sm hover:bg-[var(--color-primary-hover)] hover:shadow-md':
        variant === 'primary',
      'bg-[var(--color-secondary-dark)] text-white shadow-sm hover:opacity-90 hover:shadow-md':
        variant === 'secondary',
      'border-2 border-[var(--color-primary)] text-[var(--color-primary)] hover:bg-[var(--color-primary)] hover:text-[var(--color-primary-foreground)]':
        variant === 'outline',
      'bg-[var(--color-error)] text-white shadow-sm hover:opacity-90 hover:shadow-md':
        variant === 'danger',
      'text-[--color-foreground] hover:bg-black/5 dark:hover:bg-white/10':
        variant === 'ghost',
    },
    {
      'h-9 px-3 text-sm':   size === 'sm',
      'h-11 px-4 text-base': size === 'md',
      'h-14 px-8 text-lg':   size === 'lg',
    },
    className
  );
}
