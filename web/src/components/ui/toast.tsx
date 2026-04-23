import * as React from 'react';
import { CheckCircle2, XCircle, X } from 'lucide-react';
import { cn } from '@/lib/cn';

export type ToastVariant = 'success' | 'error';

export interface ToastItem {
  id: string;
  variant: ToastVariant;
  title: string;
  description?: string;
}

const ToastContext = React.createContext<{
  toast: (item: Omit<ToastItem, 'id'>) => void;
} | null>(null);

export function useToast() {
  const ctx = React.useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = React.useState<ToastItem[]>([]);

  const toast = React.useCallback((item: Omit<ToastItem, 'id'>) => {
    const id = crypto.randomUUID();
    setItems((prev) => [...prev, { ...item, id }]);
    setTimeout(() => {
      setItems((prev) => prev.filter((t) => t.id !== id));
    }, 5000);
  }, []);

  const dismiss = React.useCallback((id: string) => {
    setItems((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      {/* Toast viewport */}
      <div className="fixed bottom-5 right-5 z-50 grid gap-2 pointer-events-none" style={{ maxWidth: 380 }}>
        {items.map((item) => (
          <div
            key={item.id}
            className={cn(
              'pointer-events-auto flex items-start gap-3 rounded-2xl border px-4 py-3 shadow-card animate-slide-in-right',
              'bg-white dark:bg-zinc-900',
              item.variant === 'success'
                ? 'border-zinc-200 dark:border-zinc-800'
                : 'border-zinc-300 dark:border-zinc-700',
            )}
          >
            {item.variant === 'success' ? (
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-zinc-600 dark:text-zinc-300" />
            ) : (
              <XCircle className="mt-0.5 h-5 w-5 shrink-0 text-zinc-500 dark:text-zinc-400" />
            )}
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">{item.title}</p>
              {item.description && (
                <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">{item.description}</p>
              )}
            </div>
            <button
              type="button"
              onClick={() => dismiss(item.id)}
              className="shrink-0 rounded-lg p-1 text-zinc-400 transition hover:bg-zinc-100 hover:text-zinc-600 dark:hover:bg-zinc-800 dark:hover:text-zinc-300"
              aria-label="Dismiss"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
