import * as React from 'react';
import { cn } from '@/lib/cn';

export function Table({ className, ...props }: React.TableHTMLAttributes<HTMLTableElement>) {
  return <table className={cn('w-full border-separate border-spacing-0 text-sm', className)} {...props} />;
}

export function TableHead({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <thead className={className} {...props} />;
}

export function TableBody({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <tbody className={className} {...props} />;
}

export function TableRow({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) {
  return <tr className={cn('hover:bg-zinc-50 dark:hover:bg-zinc-900/70', className)} {...props} />;
}

export function TableHeadCell({ className, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) {
  return <th className={cn('border-b border-zinc-200 px-4 py-3 text-left font-semibold text-zinc-600 dark:border-zinc-800 dark:text-zinc-400', className)} {...props} />;
}

export function TableCell({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) {
  return <td className={cn('border-b border-zinc-100 px-4 py-3 align-top text-zinc-700 dark:border-zinc-900 dark:text-zinc-300', className)} {...props} />;
}