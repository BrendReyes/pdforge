# pdforge MVP Checklist

This checklist tracks implementation progress for the pdforge CLI MVP.

## Workflow Rule

- [x] Review this checklist on every project change and update item statuses as needed

## 1) Product and Branding

- [x] CLI root command uses the name pdforge
- [x] Root command description reflects privacy-first local processing
- [ ] Version flag and release metadata finalized

## 2) Command Surface (Cobra)

- [x] Root command is configured
- [x] Merge command is registered (`merge`)
- [x] Split command is registered (`split`)
- [x] Optimize command is registered (`optimize`)
- [x] Convert command is registered (`convert`)
- [x] Remove pages command is registered (`rmpage`)
- [x] Command help text updated for current command intent
- [x] Merge flags added (`-o/--output`, `-d/--dir`)
- [x] Split flags added (`-p/--page`, `-o/--output`, `-d/--dir`) plus extract/odd/even/verbose controls
- [x] Optimize flags added (`-o/--output`, `-d/--dir`)
- [x] Convert flags added (`-o/--output`, `-d/--dir`)
- [x] Remove pages flags added (`-p/--page`, `-o/--output`, `-d/--dir`)
- [ ] Backward-compatible aliases retained for old plan names (`compress`, `delete`)
- [ ] Convert command visible in standard help output (currently hidden)

## 3) Core PDF Logic (command layer with pdfcpu)

- [x] Merge implemented with pdfcpu (`api.MergeCreateFile`)
- [x] Split implemented with pdfcpu (`api.TrimFile`) for boundary and extract modes
- [x] Optimize implemented with pdfcpu (`api.OptimizeFile`)
- [x] Images-to-PDF conversion implemented with pdfcpu (`api.ImportImagesFile`)
- [x] Remove pages implemented with pdfcpu (`api.RemovePagesFile`)

Notes:
- Merge, split, optimize, convert, and remove pages now perform real file operations.
- `ensureOutputDirectory` supports interactive directory creation for CLI output paths.
- `resolveOutputPath` avoids overwrite collisions by appending ` (n)` suffixes.
- Web optimize exposes image-mode controls, but image preprocessing is still a placeholder (`preprocessPDFImages` currently returns input unchanged).

## 4) Dependencies and Project Setup

- [x] pdfcpu dependency added and pinned in `go.mod`
- [x] cobra dependency in use
- [ ] Module metadata finalized (copyright headers still use placeholder values)

## 5) Verification

- [ ] Build succeeds on Windows (blocked in this environment: `go` not available in pwsh)
- [ ] Build succeeds on macOS
- [x] Build succeeds on Linux (FedoraLinux-43)
- [x] merge command verified with real files (via source run in FedoraLinux-43)
- [ ] split command verified with real files
- [ ] optimize command verified with real files after command-surface rename
- [ ] rmpage command verified with real files
- [ ] convert command verified with real files
- [ ] Confirmed zero network calls during processing
- [x] Unit tests exist for output collision handling (`cmd/helper_test.go`)
- [x] Unit tests exist for page selector parsing and output collision behavior (`cmd/rmpage_test.go`)

## 6) Code Audit Snapshot (2026-04-24)

### Functional Go Files

- `main.go`: Entry point calls `cmd.Execute`.
- `cmd/root.go`: Root CLI wiring, help templates, ANSI formatting helpers.
- `cmd/merge.go`: Merge flow with validation, output handling, and summary report.
- `cmd/split.go`: Real split/extract flow with selector parsing and multi-output generation.
- `cmd/optimize.go`: Real optimize flow with pdfcpu optimization and reporting.
- `cmd/convert.go`: Real image-to-PDF flow via pdfcpu import.
- `cmd/rmpage.go`: Real page removal flow with selector parsing and reporting.
- `cmd/helper.go`: Output collision handling, interactive directory creation, and file reporting.
- `cmd/serve.go`: Local serve command, loopback startup, optional browser auto-open.
- `cmd/webserver.go`: Local web endpoints (`merge`, `split`, `remove`, `optimize`), CSRF/security headers, uploads, and downloads.

### Partial, Experimental, or Drift Areas

- `cmd/webserver.go`: `preprocessPDFImages` is explicitly a placeholder (no real image extraction/re-encode yet).
- `cmd/convert.go`: command is implemented but hidden from standard help (`Hidden: true`).
- `web/README.md`: endpoint naming is stale (`/api/compress` documented, implementation uses `/api/optimize`).

### Runtime Alignment Notes

- Source-level command names have moved to `optimize` and `rmpage`; older planning docs still reference `compress` and `delete`.
- This Windows environment cannot currently re-run `go build`/`go test` due missing `go` binary.

## 7) Documentation and Delivery

- [ ] Root README includes install + usage examples for all current commands (`merge`, `split`, `optimize`, `rmpage`, optional `convert`, `serve`)
- [ ] Privacy statement included in root README
- [ ] Initial GitHub release prepared

## 8) Web Implementation

- [x] `pdforge serve` launches a local browser dashboard
- [x] Merge, split, remove pages, and optimize workflows are wired to local pdfcpu actions
- [x] Web forms expose the core workflow fields for each implemented operation
- [x] Download links are generated for completed web jobs
- [x] Serve binds to loopback only (`127.0.0.1`) by default
- [x] CSRF tokens and security headers are enabled for browser requests
- [x] Multipart uploads are size-limited and temp files are cleaned up
- [x] coss ui React frontend scaffold created under `web/`
- [x] `pdforge serve` prefers `web/dist` when a frontend build exists
- [x] `pdforge serve` auto-opens the browser by default with a `--no-open` escape hatch

## 9) Experimental and Future Scopes

- [x] Start web dashboard phase with local `serve` backend + React frontend scaffold (partial implementation complete; full web product still in progress)
- [ ] Implement real image preprocessing pipeline for optimize image modes (currently placeholder in web server)
- [ ] Align CLI and web optimize flags for image tuning (`--image-mode`, `--image-max-dimension`, `--image-jpeg-quality`)
- [ ] Decide convert command product stance: keep hidden (experimental) or expose as first-class MVP command
- [ ] Decide whether to add compatibility aliases for legacy naming from older plans (`compress`, `delete`)
- [ ] Add and maintain root project README (the repository currently only has `web/README.md`)
- [ ] Finalize module/legal metadata and remove placeholder copyright headers
