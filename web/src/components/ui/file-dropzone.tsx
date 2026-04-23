import * as React from 'react';
import { Upload, X, FileText } from 'lucide-react';
import { cn } from '@/lib/cn';

export interface FileDropzoneProps {
  name: string;
  id?: string;
  accept?: string;
  multiple?: boolean;
  required?: boolean;
  onChange?: (files: File[]) => void;
  className?: string;
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

export function FileDropzone({
  name,
  id,
  accept,
  multiple = false,
  required = false,
  onChange,
  className,
}: FileDropzoneProps) {
  const [files, setFiles] = React.useState<File[]>([]);
  const [dragOver, setDragOver] = React.useState(false);
  const inputRef = React.useRef<HTMLInputElement>(null);
  const hiddenInputRef = React.useRef<HTMLInputElement>(null);

  const updateFiles = React.useCallback(
    (next: File[]) => {
      setFiles(next);
      onChange?.(next);

      // Sync to the hidden native input via DataTransfer so form submission works
      if (hiddenInputRef.current) {
        const dt = new DataTransfer();
        next.forEach((f) => dt.items.add(f));
        hiddenInputRef.current.files = dt.files;
      }
    },
    [onChange],
  );

  const handleDrop = React.useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const incoming = Array.from(e.dataTransfer.files);
      if (!multiple) {
        updateFiles(incoming.slice(0, 1));
      } else {
        updateFiles([...files, ...incoming]);
      }
    },
    [files, multiple, updateFiles],
  );

  const handleSelect = React.useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const incoming = Array.from(e.target.files ?? []);
      if (!multiple) {
        updateFiles(incoming.slice(0, 1));
      } else {
        updateFiles([...files, ...incoming]);
      }
      // Reset so the same file can be re-selected
      e.target.value = '';
    },
    [files, multiple, updateFiles],
  );

  const removeFile = React.useCallback(
    (index: number) => {
      updateFiles(files.filter((_, i) => i !== index));
    },
    [files, updateFiles],
  );

  return (
    <div className={cn('grid gap-2', className)}>
      {/* Native hidden input for form submission */}
      <input
        ref={hiddenInputRef}
        type="file"
        name={name}
        id={id}
        accept={accept}
        multiple={multiple}
        required={required && files.length === 0}
        className="sr-only"
        tabIndex={-1}
      />

      {/* Visual dropzone */}
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
        className={cn(
          'group relative flex min-h-[120px] cursor-pointer flex-col items-center justify-center rounded-2xl border-2 border-dashed transition-all duration-200',
          'border-zinc-300 bg-zinc-50/50 hover:border-zinc-400 hover:bg-zinc-100/50',
          'dark:border-zinc-700 dark:bg-zinc-900/30 dark:hover:border-zinc-500 dark:hover:bg-zinc-800/40',
          dragOver && 'dropzone-active',
        )}
      >
        <Upload
          className={cn(
            'mb-2 h-6 w-6 text-zinc-400 transition-transform duration-200 group-hover:scale-110',
            'dark:text-zinc-500',
          )}
        />
        <p className="text-sm font-medium text-zinc-500 dark:text-zinc-400">
          Drop {multiple ? 'files' : 'a file'} here, or{' '}
          <span className="font-semibold text-zinc-700 underline underline-offset-2 dark:text-zinc-200">
            browse
          </span>
        </p>
        <p className="mt-1 text-xs text-zinc-400 dark:text-zinc-500">
          {accept ? accept.replace(/\./g, '').toUpperCase() + ' files' : 'Any files'}
        </p>
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          multiple={multiple}
          onChange={handleSelect}
          className="sr-only"
          tabIndex={-1}
        />
      </div>

      {/* File list */}
      {files.length > 0 && (
        <ul className="grid gap-1.5">
          {files.map((file, i) => (
            <li
              key={`${file.name}-${file.size}-${i}`}
              className={cn(
                'flex items-center gap-2 rounded-xl border px-3 py-2 text-sm animate-fade-in',
                'border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900/60',
              )}
            >
              <FileText className="h-4 w-4 shrink-0 text-zinc-400 dark:text-zinc-500" />
              <span className="min-w-0 flex-1 truncate text-zinc-700 dark:text-zinc-300">
                {file.name}
              </span>
              <span className="shrink-0 text-xs text-zinc-400 dark:text-zinc-500">
                {formatSize(file.size)}
              </span>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  removeFile(i);
                }}
                className="shrink-0 rounded-lg p-1 text-zinc-400 transition hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
                aria-label={`Remove ${file.name}`}
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
