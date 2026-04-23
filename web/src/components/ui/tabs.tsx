import * as React from 'react';
import { cn } from '@/lib/cn';

type TabsContextValue = {
  value: string;
  setValue: (value: string) => void;
};

const TabsContext = React.createContext<TabsContextValue | null>(null);

export function Tabs({ value, defaultValue, onValueChange, className, children }: { value?: string; defaultValue?: string; onValueChange?: (value: string) => void; className?: string; children: React.ReactNode; }) {
  const [innerValue, setInnerValue] = React.useState(defaultValue ?? '');
  const resolvedValue = value ?? innerValue;

  const setValue = React.useCallback((nextValue: string) => {
    if (value === undefined) {
      setInnerValue(nextValue);
    }
    onValueChange?.(nextValue);
  }, [onValueChange, value]);

  return <TabsContext.Provider value={{ value: resolvedValue, setValue }}><div className={className}>{children}</div></TabsContext.Provider>;
}

export function TabsList({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('inline-flex rounded-2xl border border-zinc-200 bg-zinc-50 p-1 dark:border-zinc-800 dark:bg-zinc-900', className)} {...props} />;
}

export function TabsTab({ value, className, children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { value: string }) {
  const context = React.useContext(TabsContext);
  if (!context) throw new Error('TabsTab must be used within Tabs');
  const active = context.value === value;
  return (
    <button
      type="button"
      aria-selected={active}
      className={cn(
        'rounded-xl px-4 py-2 text-sm font-semibold transition',
        active ? 'bg-white text-zinc-900 shadow-sm dark:bg-zinc-800 dark:text-zinc-100' : 'text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100',
        className,
      )}
      onClick={() => context.setValue(value)}
      {...props}
    >
      {children}
    </button>
  );
}

export function TabsPanel({ value, className, children, ...props }: React.HTMLAttributes<HTMLDivElement> & { value: string }) {
  const context = React.useContext(TabsContext);
  if (!context) throw new Error('TabsPanel must be used within Tabs');
  if (context.value !== value) return null;
  return <div className={cn('mt-4', className)} {...props}>{children}</div>;
}