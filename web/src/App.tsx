import { useCallback, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  AlertCircle,
  Anvil,
  CheckCircle2,
  Download,
  Layers3,
  Minimize2,
  Scissors,
  Trash2,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardDescription, CardHeader, CardPanel, CardTitle } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Field, Fieldset } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select } from '@/components/ui/select';
import { FileDropzone } from '@/components/ui/file-dropzone';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { ToastProvider, useToast } from '@/components/ui/toast';
import { fetchCsrfToken } from '@/lib/csrf';

// ── Types ────────────────────────────────────────────────────────────────────

type WorkflowKey = 'merge' | 'split' | 'remove' | 'optimize';

type WorkflowItem = {
  key: WorkflowKey;
  label: string;
  icon: LucideIcon;
  description: string;
};

const workflows: WorkflowItem[] = [
  { key: 'merge', label: 'Merge', icon: Layers3, description: 'Combine multiple PDFs into one file.' },
  { key: 'split', label: 'Split', icon: Scissors, description: 'Extract ranges, odd/even pages, or sections.' },
  { key: 'remove', label: 'Remove Pages', icon: Trash2, description: 'Drop selected pages from a PDF.' },
  { key: 'optimize', label: 'Optimize', icon: Minimize2, description: 'Optimize PDFs with image controls.' },
];

type ApiResultItem = {
  name: string;
  size: string;
  pages: number;
  downloadURL: string;
};

type ApiResult = {
  action: string;
  summary: string;
  items: ApiResultItem[];
};

type ApiResponse = {
  ok?: boolean;
  alert?: string;
  error?: string;
  result?: ApiResult;
};

type SubmitResult = {
  alert: string;
  result: ApiResult;
};

// ── Helpers ──────────────────────────────────────────────────────────────────

async function submitForm(
  url: string,
  form: HTMLFormElement,
  csrfToken: string,
): Promise<SubmitResult> {
  const data = new FormData(form);
  data.set('csrf_token', csrfToken);

  const res = await fetch(url, {
    method: 'POST',
    body: data,
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
    },
  });

  const payload = (await res.json().catch(() => null)) as ApiResponse | null;

  if (!res.ok) {
    const message = payload?.alert || payload?.error || `Server error ${res.status}`;
    throw new Error(message);
  }

  if (!payload?.ok || !payload.result || payload.result.items.length === 0) {
    const message = payload?.alert || payload?.error || 'Operation did not produce downloadable output.';
    throw new Error(message);
  }

  return {
    alert: payload.alert ?? 'Operation completed successfully.',
    result: payload.result,
  };
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

async function downloadResultItem(item: ApiResultItem) {
  const response = await fetch(item.downloadURL, {
    method: 'GET',
    credentials: 'same-origin',
  });

  if (!response.ok) {
    throw new Error(`Download failed for ${item.name}`);
  }

  const disposition = response.headers.get('Content-Disposition') ?? '';
  const filenameMatch = disposition.match(/filename="?([^";\n]+)"?/);
  const filename = filenameMatch?.[1] ?? item.name;
  const blob = await response.blob();
  downloadBlob(blob, filename);
}

// ── App ──────────────────────────────────────────────────────────────────────

function AppInner() {
  const [csrfToken, setCsrfToken] = useState('');
  const [tokenError, setTokenError] = useState('');
  const [loadingToken, setLoadingToken] = useState(true);
  const [activeWorkflow, setActiveWorkflow] = useState<WorkflowKey>('merge');

  useEffect(() => {
    let mounted = true;

    fetchCsrfToken()
      .then((token) => {
        if (!mounted) return;
        setCsrfToken(token);
        setTokenError('');
      })
      .catch(() => {
        if (!mounted) return;
        setTokenError('Unable to initialize secure form submission token.');
      })
      .finally(() => {
        if (mounted) setLoadingToken(false);
      });

    return () => {
      mounted = false;
    };
  }, []);

  const active = workflows.find((w) => w.key === activeWorkflow) ?? workflows[0];
  const statusLabel = loadingToken ? 'Initializing…' : csrfToken ? 'Ready' : 'Token missing';
  const statusDotColor = loadingToken
    ? 'bg-zinc-400 animate-pulse-subtle'
    : csrfToken
      ? 'bg-zinc-900 dark:bg-zinc-100'
      : 'bg-zinc-400';

  return (
    <main className="min-h-screen">
      <div className="mx-auto w-full max-w-[1440px] px-4 py-4 sm:px-6 lg:px-8">
        {/* ── Top bar ────────────────────────────────────────── */}
        <header className="mb-4 flex items-center justify-between gap-3 animate-fade-in">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-zinc-900 text-white dark:bg-white dark:text-zinc-900">
              <Anvil className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-sm font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                pdforge
              </h1>
              <p className="text-[11px] text-zinc-400 dark:text-zinc-500">Local PDF toolkit</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2 rounded-full border border-zinc-200 bg-white/70 px-3 py-1.5 backdrop-blur dark:border-zinc-800 dark:bg-zinc-900/70">
              <span className={`h-2 w-2 rounded-full ${statusDotColor}`} />
              <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                {statusLabel}
              </span>
            </div>
            <ThemeToggle />
          </div>
        </header>

        {/* ── Main grid ────────────────────────────────────── */}
        <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
          {/* Sidebar */}
          <aside className="glass-strong rounded-2xl border border-zinc-200/70 p-3 dark:border-zinc-800/70 animate-fade-in-up">
            <p className="mb-2 px-1 text-[11px] font-semibold uppercase tracking-widest text-zinc-400 dark:text-zinc-500">
              Workflows
            </p>
            <nav className="grid gap-1.5">
              {workflows.map((item) => {
                const Icon = item.icon;
                const isActive = item.key === activeWorkflow;
                return (
                  <button
                    key={item.key}
                    type="button"
                    onClick={() => setActiveWorkflow(item.key)}
                    className={[
                      'group rounded-xl px-3 py-2.5 text-left transition-all duration-200',
                      isActive
                        ? 'bg-zinc-900 text-white shadow-card dark:bg-white dark:text-zinc-900'
                        : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800/60',
                    ].join(' ')}
                  >
                    <span className="flex items-center gap-2.5">
                      <Icon className="h-4 w-4" />
                      <span className="text-sm font-semibold">{item.label}</span>
                    </span>
                    <p
                      className={[
                        'mt-0.5 pl-[26px] text-[11px] leading-relaxed transition-colors',
                        isActive
                          ? 'text-zinc-300 dark:text-zinc-600'
                          : 'text-zinc-400 dark:text-zinc-500',
                      ].join(' ')}
                    >
                      {item.description}
                    </p>
                  </button>
                );
              })}
            </nav>

            {/* Session card */}
            <div className="mt-4 rounded-xl border border-zinc-200/60 bg-white/50 p-3 dark:border-zinc-800/60 dark:bg-zinc-900/40">
              <div className="flex items-center justify-between">
                <p className="text-xs font-semibold text-zinc-500 dark:text-zinc-400">Session</p>
                <Badge variant="secondary">127.0.0.1</Badge>
              </div>
              <p className="mt-1 text-[11px] text-zinc-400 dark:text-zinc-500">
                All files stay on your machine. Zero cloud uploads.
              </p>
            </div>
          </aside>

          {/* Content */}
          <section
            key={activeWorkflow}
            className="glass-strong rounded-2xl border border-zinc-200/70 p-4 dark:border-zinc-800/70 sm:p-5 animate-fade-in-up"
          >
            <div className="mb-5 flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
                  <active.icon className="h-4 w-4" />
                </div>
                <div>
                  <h2 className="text-base font-bold text-zinc-900 dark:text-zinc-100">
                    {active.label}
                  </h2>
                  <p className="text-xs text-zinc-400 dark:text-zinc-500">{active.description}</p>
                </div>
              </div>
            </div>

            {tokenError ? (
              <Card className="mb-4 border-zinc-300 bg-zinc-50/50 dark:border-zinc-700 dark:bg-zinc-900/40">
                <CardPanel className="flex items-start gap-2">
                  <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-zinc-500 dark:text-zinc-400" />
                  <p className="text-sm text-zinc-700 dark:text-zinc-300">{tokenError}</p>
                </CardPanel>
              </Card>
            ) : null}

            {activeWorkflow === 'merge' && <MergePanel csrfToken={csrfToken} />}
            {activeWorkflow === 'split' && <SplitPanel csrfToken={csrfToken} />}
            {activeWorkflow === 'remove' && <RemovePanel csrfToken={csrfToken} />}
            {activeWorkflow === 'optimize' && <OptimizePanel csrfToken={csrfToken} />}
          </section>
        </div>
      </div>
    </main>
  );
}

function App() {
  return (
    <ToastProvider>
      <AppInner />
    </ToastProvider>
  );
}

// ── useFormSubmit hook ────────────────────────────────────────────────────────

function useFormSubmit(endpoint: string, csrfToken: string) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ message: string } | null>(null);

  const handleSubmit = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setLoading(true);
      setResult(null);
      try {
        const { alert, result: operationResult } = await submitForm(endpoint, e.currentTarget, csrfToken);

        for (const item of operationResult.items) {
          await downloadResultItem(item);
        }

        const summaryMessage =
          operationResult.items.length === 1
            ? `Saved as ${operationResult.items[0].name}`
            : `Saved ${operationResult.items.length} files`;

        setResult({ message: summaryMessage });
        toast({ variant: 'success', title: 'Done!', description: alert });
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'Something went wrong';
        toast({ variant: 'error', title: 'Operation failed', description: msg });
      } finally {
        setLoading(false);
      }
    },
    [endpoint, csrfToken, toast],
  );

  return { loading, result, handleSubmit };
}

// ── Result banner ────────────────────────────────────────────────────────────

function ResultBanner({ message }: { message: string }) {
  return (
    <div className="flex items-center gap-2 rounded-xl border border-zinc-300 bg-zinc-50/50 px-3 py-2 text-sm animate-fade-in dark:border-zinc-700 dark:bg-zinc-900/40">
      <CheckCircle2 className="h-4 w-4 text-zinc-600 dark:text-zinc-300" />
      <span className="text-zinc-700 dark:text-zinc-300">{message}</span>
      <Download className="ml-auto h-4 w-4 text-zinc-400" />
    </div>
  );
}

// ── Workflow Panels ──────────────────────────────────────────────────────────

function MergePanel({ csrfToken }: { csrfToken: string }) {
  const { loading, result, handleSubmit } = useFormSubmit('/api/merge', csrfToken);

  return (
    <form className="grid gap-4 xl:grid-cols-[1fr_340px]" onSubmit={handleSubmit}>
      <input type="hidden" name="csrf_token" value={csrfToken} />

      <Fieldset>
        <legend className="px-1 text-sm font-semibold text-zinc-700 dark:text-zinc-300">Input</legend>
        <div className="grid gap-4">
          <Field>
            <Label htmlFor="merge-files">PDF files</Label>
            <FileDropzone id="merge-files" name="files" multiple accept=".pdf" required />
            <p className="text-[11px] text-zinc-400 dark:text-zinc-500">
              Files are merged in the order listed above.
            </p>
          </Field>

          {result && <ResultBanner message={result.message} />}

          <Button type="submit" loading={loading}>
            <Layers3 className="h-4 w-4" />
            Run merge
          </Button>
        </div>
      </Fieldset>

      <div className="grid gap-4 self-start">
        <FlagsCard title="Flags" subtitle="No extra flags needed for merge.">
          <p className="text-xs text-zinc-400 dark:text-zinc-500">
            Use output controls below for naming and destination.
          </p>
        </FlagsCard>
        <OutputCard workflow="merge" outputName="output" outputId="merge-output" dirName="dir" dirId="merge-dir" />
      </div>
    </form>
  );
}

function SplitPanel({ csrfToken }: { csrfToken: string }) {
  const { loading, result, handleSubmit } = useFormSubmit('/api/split', csrfToken);

  return (
    <form className="grid gap-4 xl:grid-cols-[1fr_340px]" onSubmit={handleSubmit}>
      <input type="hidden" name="csrf_token" value={csrfToken} />

      <Fieldset>
        <legend className="px-1 text-sm font-semibold text-zinc-700 dark:text-zinc-300">Input</legend>
        <div className="grid gap-4">
          <Field>
            <Label htmlFor="split-file">PDF file</Label>
            <FileDropzone id="split-file" name="file" accept=".pdf" required />
          </Field>
          <Field>
            <Label htmlFor="split-selector">Page selector</Label>
            <Input id="split-selector" name="selector" placeholder="8 or 6,8-10,11" />
          </Field>

          {result && <ResultBanner message={result.message} />}

          <Button type="submit" loading={loading}>
            <Scissors className="h-4 w-4" />
            Run split
          </Button>
        </div>
      </Fieldset>

      <div className="grid gap-4 self-start">
        <AdvancedFlagsCard title="Flags" subtitle="Advanced split options">
          <Field>
            <Label htmlFor="split-page">--page flag override</Label>
            <Input id="split-page" name="page" placeholder="same format as selector" />
          </Field>
          <div className="grid gap-2">
            <OptionCheck label="Extract mode (-e)" name="extract" />
            <OptionCheck label="Odd pages (--odd)" name="odd" />
            <OptionCheck label="Even pages (--even)" name="even" />
            <OptionCheck label="Verbose (--verbose)" name="verbose" />
          </div>
        </AdvancedFlagsCard>
        <OutputCard workflow="split" outputName="output" outputId="split-output" dirName="dir" dirId="split-dir" />
      </div>
    </form>
  );
}

function RemovePanel({ csrfToken }: { csrfToken: string }) {
  const { loading, result, handleSubmit } = useFormSubmit('/api/remove', csrfToken);

  return (
    <form className="grid gap-4 xl:grid-cols-[1fr_340px]" onSubmit={handleSubmit}>
      <input type="hidden" name="csrf_token" value={csrfToken} />

      <Fieldset>
        <legend className="px-1 text-sm font-semibold text-zinc-700 dark:text-zinc-300">Input</legend>
        <div className="grid gap-4">
          <Field>
            <Label htmlFor="remove-file">PDF file</Label>
            <FileDropzone id="remove-file" name="file" accept=".pdf" required />
          </Field>
          <Field>
            <Label htmlFor="remove-page">Page selector</Label>
            <Input id="remove-page" name="page" placeholder="1,6-11,17" required />
          </Field>

          {result && <ResultBanner message={result.message} />}

          <Button type="submit" loading={loading}>
            <Trash2 className="h-4 w-4" />
            Run page removal
          </Button>
        </div>
      </Fieldset>

      <div className="grid gap-4 self-start">
        <FlagsCard title="Flags" subtitle="Remove pages only needs selector input.">
          <p className="text-xs text-zinc-400 dark:text-zinc-500">No additional flags required.</p>
        </FlagsCard>
        <OutputCard workflow="remove" outputName="output" outputId="remove-output" dirName="dir" dirId="remove-dir" />
      </div>
    </form>
  );
}

function OptimizePanel({ csrfToken }: { csrfToken: string }) {
  const { loading, result, handleSubmit } = useFormSubmit('/api/optimize', csrfToken);

  return (
    <form className="grid gap-4 xl:grid-cols-[1fr_340px]" onSubmit={handleSubmit}>
      <input type="hidden" name="csrf_token" value={csrfToken} />

      <Fieldset>
        <legend className="px-1 text-sm font-semibold text-zinc-700 dark:text-zinc-300">Input</legend>
        <div className="grid gap-4">
          <Field>
            <Label htmlFor="optimize-file">PDF file</Label>
            <FileDropzone id="optimize-file" name="file" accept=".pdf" required />
          </Field>

          {result && <ResultBanner message={result.message} />}

          <Button type="submit" loading={loading}>
            <Minimize2 className="h-4 w-4" />
            Run optimization
          </Button>
        </div>
      </Fieldset>

      <div className="grid gap-4 self-start">
        <AdvancedFlagsCard title="Flags" subtitle="Advanced optimization settings">
          <Field>
            <Label htmlFor="optimize-mode">Image mode (--image-mode)</Label>
            <Select id="optimize-mode" name="image_mode" defaultValue="off">
              <option value="off">off</option>
              <option value="readable">readable</option>
              <option value="balanced">balanced</option>
              <option value="aggressive">aggressive</option>
              <option value="experimental">experimental</option>
            </Select>
          </Field>
          <Field>
            <Label htmlFor="optimize-dimension">Max dimension (--image-max-dimension)</Label>
            <Input id="optimize-dimension" type="number" min={0} name="image_max_dimension" placeholder="0 = auto" />
          </Field>
          <Field>
            <Label htmlFor="optimize-quality">JPEG quality (--image-jpeg-quality)</Label>
            <Input id="optimize-quality" type="number" min={0} max={100} name="image_jpeg_quality" placeholder="0 = auto" />
          </Field>
        </AdvancedFlagsCard>
        <OutputCard workflow="optimize" outputName="output" outputId="optimize-output" dirName="dir" dirId="optimize-dir" />
      </div>
    </form>
  );
}

// ── Shared sub-components ────────────────────────────────────────────────────

function FlagsCard({ title, subtitle, children }: { title: string; subtitle: string; children: ReactNode }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">{title}</CardTitle>
        <CardDescription>{subtitle}</CardDescription>
      </CardHeader>
      <CardPanel>{children}</CardPanel>
    </Card>
  );
}

function AdvancedFlagsCard({ title, subtitle, children }: { title: string; subtitle: string; children: ReactNode }) {
  const [open, setOpen] = useState(false);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">{title}</CardTitle>
        <CardDescription>{subtitle}</CardDescription>
      </CardHeader>
      <CardPanel>
        <button
          type="button"
          onClick={() => setOpen((o) => !o)}
          className="flex w-full items-center justify-between rounded-xl border border-zinc-200 bg-zinc-50/50 px-3 py-2 text-left text-sm font-medium text-zinc-600 transition hover:bg-zinc-100 dark:border-zinc-800 dark:bg-zinc-900/40 dark:text-zinc-400 dark:hover:bg-zinc-800/60"
        >
          <span>{open ? 'Hide advanced options' : 'Show advanced options'}</span>
          <svg
            className={`h-4 w-4 transition-transform duration-200 ${open ? 'rotate-180' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        {open && (
          <div className="mt-3 grid gap-4 animate-fade-in">
            {children}
          </div>
        )}
      </CardPanel>
    </Card>
  );
}

function OutputCard({
  workflow,
  outputName,
  outputId,
  dirName,
  dirId,
}: {
  workflow: string;
  outputName: string;
  outputId: string;
  dirName: string;
  dirId: string;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">Output</CardTitle>
        <CardDescription>Naming and destination for {workflow} results.</CardDescription>
      </CardHeader>
      <CardPanel className="grid gap-4">
        <Field>
          <Label htmlFor={outputId}>Output file (-o / --output)</Label>
          <Input id={outputId} name={outputName} placeholder="optional output filename" />
        </Field>
        <Field>
          <Label htmlFor={dirId}>Output directory (-d / --dir)</Label>
          <Input id={dirId} name={dirName} placeholder="optional output directory" />
        </Field>
      </CardPanel>
    </Card>
  );
}

function OptionCheck({ label, name }: { label: string; name: string }) {
  return (
    <label className="flex items-center gap-2.5 rounded-xl border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-700 transition hover:border-zinc-300 hover:bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-900/60 dark:text-zinc-300 dark:hover:border-zinc-700 dark:hover:bg-zinc-800/50">
      <Checkbox name={name} />
      <span>{label}</span>
    </label>
  );
}

export default App;
