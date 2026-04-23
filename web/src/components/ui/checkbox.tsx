import * as React from 'react';
import { cn } from '@/lib/cn';

export interface CheckboxProps extends React.InputHTMLAttributes<HTMLInputElement> {}

export const Checkbox = React.forwardRef<HTMLInputElement, CheckboxProps>(function Checkbox({ className, ...props }, ref) {
  return (
    <input
      ref={ref}
      type="checkbox"
      className={cn('h-4 w-4 rounded border-zinc-300 text-zinc-900 accent-zinc-900 focus:ring-zinc-500/20 dark:border-zinc-700 dark:bg-zinc-900 dark:text-white dark:accent-white', className)}
      {...props}
    />
  );
});