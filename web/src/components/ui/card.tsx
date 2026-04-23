import * as React from 'react';
import { cn } from '@/lib/cn';

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {}
export interface CardFrameProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Card({ className, ...props }: CardProps) {
  return <div className={cn('rounded-2xl border border-zinc-200 bg-white shadow-soft dark:border-zinc-800/60 dark:bg-zinc-900', className)} {...props} />;
}

export function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('grid gap-2 border-b border-zinc-200 px-5 py-4 dark:border-zinc-800/60', className)} {...props} />;
}

export function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn('text-base font-semibold text-zinc-900 dark:text-zinc-100', className)} {...props} />;
}

export function CardDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn('text-sm text-zinc-500 dark:text-zinc-400', className)} {...props} />;
}

export function CardAction({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('ml-auto flex items-center justify-end gap-2', className)} {...props} />;
}

export function CardPanel({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('px-5 py-4', className)} {...props} />;
}

export function CardFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('border-t border-zinc-200 px-5 py-4 dark:border-zinc-800/60', className)} {...props} />;
}

export function CardFrame({ className, ...props }: CardFrameProps) {
  return <div className={cn('rounded-3xl border border-zinc-200 bg-white shadow-soft dark:border-zinc-800/60 dark:bg-zinc-900', className)} {...props} />;
}

export function CardFrameHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('grid gap-2 px-5 py-4 md:grid-cols-[1fr_auto]', className)} {...props} />;
}

export function CardFrameTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h3 className={cn('text-lg font-semibold text-zinc-900 dark:text-zinc-100', className)} {...props} />;
}

export function CardFrameDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn('text-sm text-zinc-500 dark:text-zinc-400', className)} {...props} />;
}

export function CardFrameAction({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('md:self-center md:justify-self-end', className)} {...props} />;
}

export function CardFrameFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('border-t border-zinc-200 px-5 py-4 dark:border-zinc-800/60', className)} {...props} />;
}