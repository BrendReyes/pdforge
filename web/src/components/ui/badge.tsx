import * as React from 'react';
import { cn } from '@/lib/cn';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'secondary' | 'success' | 'warning' | 'destructive';
}

const variantStyles: Record<NonNullable<BadgeProps['variant']>, string> = {
  default: 'bg-zinc-900 text-white dark:bg-white dark:text-zinc-900',
  secondary: 'bg-zinc-100 text-zinc-800',
  success: 'bg-zinc-100 text-zinc-700 ring-1 ring-zinc-300 dark:bg-zinc-800 dark:text-zinc-300 dark:ring-zinc-700',
  warning: 'bg-zinc-100 text-zinc-600 ring-1 ring-zinc-300 dark:bg-zinc-800 dark:text-zinc-400 dark:ring-zinc-700',
  destructive: 'bg-zinc-200 text-zinc-800 ring-1 ring-zinc-400 dark:bg-zinc-700 dark:text-zinc-200 dark:ring-zinc-600',
};

export function Badge({ className, variant = 'default', ...props }: BadgeProps) {
  return <span className={cn('inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold', variantStyles[variant], className)} {...props} />;
}