import * as React from 'react';
import { cn } from '@/lib/cn';

export interface LabelProps extends React.LabelHTMLAttributes<HTMLLabelElement> {}

export function Label({ className, ...props }: LabelProps) {
  return <label className={cn('text-sm font-semibold text-zinc-700 dark:text-zinc-300', className)} {...props} />;
}