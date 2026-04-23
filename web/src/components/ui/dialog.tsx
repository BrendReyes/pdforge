import * as React from 'react';
import { cn } from '@/lib/cn';

export function Dialog({ open, children }: { open: boolean; children: React.ReactNode }) {
  if (!open) return null;
  return <div className="fixed inset-0 z-50 flex items-center justify-center bg-zinc-950/50 p-4">{children}</div>;
}

export function DialogPanel({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('w-full max-w-lg rounded-2xl border border-zinc-200 bg-white p-6 shadow-soft', className)} {...props} />;
}

export function DialogTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn('text-lg font-semibold text-zinc-900', className)} {...props} />;
}

export function DialogDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn('mt-2 text-sm text-zinc-500', className)} {...props} />;
}