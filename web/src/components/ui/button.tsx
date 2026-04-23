import * as React from 'react';
import { Loader2 } from 'lucide-react';
import { cn } from '@/lib/cn';

type ButtonVariant = 'default' | 'secondary' | 'outline' | 'ghost' | 'link' | 'destructive';
type ButtonSize = 'xs' | 'sm' | 'default' | 'lg' | 'xl' | 'icon';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  render?: React.ReactElement;
}

const variantStyles: Record<ButtonVariant, string> = {
  default: 'bg-zinc-900 text-white shadow-card hover:bg-zinc-800 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-200',
  secondary: 'bg-zinc-100 text-zinc-900 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-100 dark:hover:bg-zinc-700',
  outline: 'border border-zinc-200 bg-white text-zinc-900 hover:bg-zinc-50 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-100 dark:hover:bg-zinc-800',
  ghost: 'bg-transparent text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800',
  link: 'bg-transparent text-zinc-700 underline-offset-4 hover:underline dark:text-zinc-300',
  destructive: 'bg-zinc-800 text-white hover:bg-zinc-900 dark:bg-zinc-200 dark:text-zinc-900 dark:hover:bg-zinc-100',
};

const sizeStyles: Record<ButtonSize, string> = {
  xs: 'h-8 px-2.5 text-xs rounded-md',
  sm: 'h-9 px-3 text-sm rounded-md',
  default: 'h-10 px-4 text-sm rounded-xl',
  lg: 'h-11 px-5 text-sm rounded-xl',
  xl: 'h-12 px-6 text-base rounded-xl',
  icon: 'h-10 w-10 rounded-xl p-0',
};

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, variant = 'default', size = 'default', loading = false, disabled, children, ...props },
  ref,
) {
  const isDisabled = disabled || loading;

  if (props.render) {
    const rendered = React.cloneElement(props.render, {
      className: cn(
        'inline-flex items-center justify-center gap-2 font-medium transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
        'focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-950',
        variantStyles[variant],
        sizeStyles[size],
        props.render.props.className,
        className,
      ),
      children,
    });

    return rendered;
  }

  return (
    <button
      ref={ref}
      className={cn(
        'inline-flex items-center justify-center gap-2 font-medium transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
        'focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-950',
        variantStyles[variant],
        sizeStyles[size],
        className,
      )}
      disabled={isDisabled}
      aria-disabled={isDisabled}
      data-loading={loading ? 'true' : undefined}
      {...props}
    >
      {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
      {children}
    </button>
  );
});