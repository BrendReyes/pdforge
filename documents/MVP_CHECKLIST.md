# pdforge MVP Checklist

This checklist tracks implementation progress for the pdforge CLI MVP.

## Workflow Rule

- [ ] Review this checklist on every project change and update item statuses as needed

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
- [x] Command help text updated for MVP intent
- [ ] Merge flags added (-o/--output)
- [ ] Split flags added (-p/--pages, -o/--output)
- [ ] Compress flags added (-o/--output)
- [ ] Convert flags added (-o/--output)

## 3) Core PDF Logic (internal/pdf)

- [x] MergeFiles implemented with pdfcpu
- [ ] SplitByPages implemented with pdfcpu
- [ ] CompressFile implemented with pdfcpu
- [ ] ImagesToPDF implemented with pdfcpu

## 4) Dependencies and Project Setup

- [ ] pdfcpu dependency added and pinned
- [x] cobra dependency in use
- [ ] Module metadata finalized

## 5) Verification

- [ ] Build succeeds on Windows
- [ ] Build succeeds on macOS
- [ ] Build succeeds on Linux
- [ ] merge command verified with real files
- [ ] split command verified with real files
- [ ] compress command verified with real files
- [ ] convert command verified with real files
- [ ] Confirmed zero network calls during processing

## 6) Documentation and Delivery

- [ ] README includes install + usage examples for all commands
- [ ] Privacy statement included in README
- [ ] Initial GitHub release prepared
