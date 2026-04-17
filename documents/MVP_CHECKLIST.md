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
- [x] Merge command is registered
- [x] Split command is registered
- [x] Compress command is registered
- [x] Convert command is registered
- [x] Delete command is registered (currently out of MVP scope)
- [x] Command help text updated for MVP intent
- [x] Merge flags added (-o/--output, -d/--dir)
- [x] Split flags added (-p/--pages, -o/--output)
- [x] Compress flags added (-o/--output)
- [x] Convert flags added (-o/--output)

## 3) Core PDF Logic (internal/pdf)

- [x] Merge implemented with pdfcpu in command layer
- [ ] SplitByPages implemented with pdfcpu
- [x] CompressFile implemented with pdfcpu in command layer
- [ ] ImagesToPDF implemented with pdfcpu
- [ ] Delete pages implemented with pdfcpu

Notes:
- Merge has real PDF logic and output reporting.
- Merge currently mishandles absolute --output paths because output is always joined with --dir.
- Compress now has real PDF optimization logic and output reporting.
- Compress now includes optional image modes with safer defaults: off, balanced, aggressive.
- Compress detects image-heavy PDFs early and suggests enabling experimental mode.
- Compress supports interactive y/n opt-in when image-heavy input is detected and --image-mode is off (enables balanced mode).
- Advanced image tuning flags exist but are hidden from standard help output.
- Split, convert, and delete currently print placeholder messages only.

## 4) Dependencies and Project Setup

- [x] pdfcpu dependency added and pinned
- [x] cobra dependency in use
- [ ] Module metadata finalized

## 5) Verification

- [ ] Build succeeds on Windows (blocked in this environment: go not available in pwsh)
- [ ] Build succeeds on macOS
- [x] Build succeeds on Linux (FedoraLinux-43)
- [x] merge command verified with real files (via source run in FedoraLinux-43)
- [ ] split command verified with real files
- [x] compress command verified with real files (via source run in FedoraLinux-43)
- [x] compress experimental image mode verified with real files (via source run in FedoraLinux-43)
- [ ] convert command verified with real files
- [ ] Confirmed zero network calls during processing

## 7) Code Audit Snapshot (2026-04-17)

### Functional Go Files

- main.go: Entry point calls cmd.Execute.
- cmd/root.go: Root CLI wiring, help template, ANSI formatting helpers.
- cmd/merge.go: Real merge flow using pdfcpu validation + merge + summary report.
- cmd/compress.go: Real compress flow using pdfcpu validation + optimize + summary report.
- cmd/compress.go: Includes optional experimental image preprocessing (re-encode/downsample) before optimize.
- cmd/helper.go: Output collision handling and file summary helpers.
- cmd/helper_test.go: Tests for resolveOutputPath edge cases.

### Placeholder or Partial Go Files

- cmd/split.go: Placeholder; prints input/pages/output and exits.
- cmd/convert.go: Placeholder; prints image count/output only.
- cmd/delete.go: Placeholder; prints delete called only.

### Runtime Alignment Note

- The checked-in pdforge binary does not reflect current source registration for delete and still behaves as a stub build for merge.
- Source run in FedoraLinux-43 confirms merge works when output is provided as filename plus --dir.

## 6) Documentation and Delivery

- [ ] README includes install + usage examples for all commands
- [ ] Privacy statement included in README
- [ ] Initial GitHub release prepared
