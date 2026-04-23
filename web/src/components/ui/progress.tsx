import { cn } from '@/lib/cn';

export function Progress({ value = 0, className }: { value?: number; className?: string }) {
  return (
    <div className={cn('h-2 w-full overflow-hidden rounded-full bg-zinc-200', className)}>
      <div className="h-full rounded-full bg-gradient-to-r from-zinc-700 to-zinc-400 dark:from-zinc-300 dark:to-zinc-500 transition-all" style={{ width: `${Math.max(0, Math.min(100, value))}%` }} />
    </div>
  );
}