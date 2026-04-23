import * as React from 'react';
import { cn } from '@/lib/cn';

export function Field({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('grid gap-1.5', className)} {...props} />;
}

export function Fieldset({ className, ...props }: React.FieldsetHTMLAttributes<HTMLFieldSetElement>) {
  return <fieldset className={cn('grid gap-4 rounded-2xl border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900', className)} {...props} />;
}
