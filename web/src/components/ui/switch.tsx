import * as React from 'react';
import { cn } from '@/lib/cn';

export interface SwitchProps extends React.InputHTMLAttributes<HTMLInputElement> {}

export const Switch = React.forwardRef<HTMLInputElement, SwitchProps>(function Switch({ className, ...props }, ref) {
  return (
    <label className={cn('inline-flex cursor-pointer items-center', className)}>
      <input ref={ref} type="checkbox" className="peer sr-only" {...props} />
      <span className="relative h-6 w-11 rounded-full bg-zinc-300 transition peer-checked:bg-zinc-900 dark:bg-zinc-700 dark:peer-checked:bg-white">
        <span className="absolute left-1 top-1 h-4 w-4 rounded-full bg-white transition peer-checked:translate-x-5" />
      </span>
    </label>
  );
});