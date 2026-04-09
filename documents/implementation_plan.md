# Forge MVP — Implementation Plan

**Forge** is a local, open-source CLI tool for PDF manipulation. All processing stays on the user's machine — no uploads, no cloud, no privacy concerns.

## SPARK Alignment

| SPARK | Forge |
|-------|-------|
| **S**cope | CLI-first MVP with 4 core features: Merge, Split, Compress, Convert (images→PDF) |
| **P**lan | Go + `pdfcpu` (Apache 2.0) + `cobra` CLI framework |
| **A**llocation | Single developer; CLI first, web dashboard later |
| **R**isk | `pdfcpu` covers all 4 features natively; no external API dependency |
| **K**ey Metrics | All 4 commands work end-to-end with real files |

---

## Architecture

```
forge/
├── main.go                  # Entry point → cmd.Execute()
├── go.mod / go.sum
├── README.md
├── cmd/                     # CLI command definitions (Cobra)
│   ├── root.go              # Root command + global flags
│   ├── merge.go             # forge merge a.pdf b.pdf -o out.pdf
│   ├── split.go             # forge split file.pdf -p 1-3,5
│   ├── compress.go          # forge compress file.pdf -o out.pdf
│   └── convert.go           # forge convert img.jpg img.png -o out.pdf
└── internal/pdf/            # Core PDF logic (wraps pdfcpu)
    ├── merge.go
    ├── split.go
    ├── compress.go
    └── convert.go
```

---

## Proposed Changes

### Dependencies

#### [MODIFY] [go.mod](file:///c:/Users/Wayne%20Karlo/Documents/forge/go.mod)

Update module name to `github.com/brendreyes/forge` and add dependencies:
- `github.com/pdfcpu/pdfcpu` — PDF processing (merge, split, compress, image→PDF)
- `github.com/spf13/cobra` — CLI framework

---

### Project Entry Point

#### [NEW] [main.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/main.go)

Minimal entry point that calls `cmd.Execute()`.

---

### CLI Commands (`cmd/`)

#### [NEW] [root.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/cmd/root.go)

Root Cobra command with app description, version flag, and ASCII banner.

#### [NEW] [merge.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/cmd/merge.go)

`forge merge <file1.pdf> [file2.pdf ...] -o <output.pdf>`
- Requires 2+ input files
- `-o` / `--output` flag for destination (default: `merged.pdf`)

#### [NEW] [split.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/cmd/split.go)

`forge split <file.pdf> -p <pages> -o <output_dir>`
- `-p` / `--pages` flag: page specification (e.g., `1-3,5,8-10`)
- `-o` / `--output` flag: output directory (default: current directory)

#### [NEW] [compress.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/cmd/compress.go)

`forge compress <file.pdf> -o <output.pdf>`
- Shows before/after file sizes
- `-o` / `--output` flag (default: `<filename>_compressed.pdf`)

#### [NEW] [convert.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/cmd/convert.go)

`forge convert <image1> [image2 ...] -o <output.pdf>`
- Accepts `.jpg`, `.jpeg`, `.png`, `.tiff`, `.bmp`
- `-o` / `--output` flag (default: `converted.pdf`)

---

### Core PDF Logic (`internal/pdf/`)

#### [NEW] [merge.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/internal/pdf/merge.go)

`MergeFiles(inputPaths []string, outputPath string) error` — wraps `pdfcpu.MergeCreateFile()`.

#### [NEW] [split.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/internal/pdf/split.go)

`SplitByPages(inputPath, outputDir string, pages string) error` — parses page specs, extracts pages via `pdfcpu`.

#### [NEW] [compress.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/internal/pdf/compress.go)

`CompressFile(inputPath, outputPath string) error` — wraps `pdfcpu.OptimizeFile()`.

#### [NEW] [convert.go](file:///c:/Users/Wayne%20Karlo/Documents/forge/internal/pdf/convert.go)

`ImagesToPDF(imagePaths []string, outputPath string) error` — converts images to PDF via `pdfcpu.ImportImagesFile()`.

---

### Documentation

#### [NEW] [README.md](file:///c:/Users/Wayne%20Karlo/Documents/forge/README.md)

Project overview, installation instructions, usage examples for all 4 commands, and privacy mission statement.

---

## Verification Plan

### Automated CLI Testing

Run the built binary against real files to verify all 4 commands:

```bash
# Build
go build -o forge.exe .

# 1. Merge — combine 2 PDFs
forge.exe merge testdata/a.pdf testdata/b.pdf -o testdata/merged.pdf
# ✓ merged.pdf exists and is a valid PDF

# 2. Split — extract pages
forge.exe split testdata/merged.pdf -p 1 -o testdata/split_out
# ✓ output file(s) created in split_out/

# 3. Compress — optimize PDF
forge.exe compress testdata/a.pdf -o testdata/compressed.pdf
# ✓ compressed.pdf exists, file size ≤ original

# 4. Convert — images to PDF
forge.exe convert testdata/sample.jpg -o testdata/from_image.pdf
# ✓ from_image.pdf exists and is a valid PDF
```

I'll create small test PDF and image files in `testdata/` to run these tests.

### Manual Verification

After I run the automated tests above, I'll ask you to try the commands yourself on any PDF files you have locally to confirm everything works on your machine.
