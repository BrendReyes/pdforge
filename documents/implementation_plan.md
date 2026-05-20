# pdforge MVP — Implementation Plan

**pdforge** is a local, open-source CLI tool for PDF manipulation. All processing stays on the user's machine — no uploads, no cloud, no privacy concerns.

## SPARK Alignment

| SPARK | pdforge |
|-------|-------|
| **S**cope | CLI-first MVP with 4 core features: Merge, Split, Compress, Convert (images→PDF) |
| **P**lan | Go + `pdfcpu` (Apache 2.0) + `cobra` CLI framework |
| **A**llocation | Single developer; CLI first, web dashboard later |
| **R**isk | `pdfcpu` covers all 4 features natively; no external API dependency |
| **K**ey Metrics | All 4 commands work end-to-end with real files |

---

## Architecture

```
pdforge/
├── main.go                  # Entry point → cmd.Execute()
├── go.mod / go.sum
├── README.md
├── cmd/                     # CLI command definitions (Cobra)
│   ├── root.go              # Root command + global flags
│   ├── merge.go             # pdforge merge a.pdf b.pdf -o out.pdf
│   ├── split.go             # pdforge split file.pdf -p 1-3,5
│   ├── compress.go          # pdforge compress file.pdf -o out.pdf
│   └── convert.go           # pdforge convert img.jpg img.png -o out.pdf
└── internal/pdf/            # Core PDF logic (wraps pdfcpu)
    ├── merge.go
    ├── split.go
    ├── compress.go
    └── convert.go
```

---

## Proposed Changes

### Dependencies

#### [MODIFY] [go.mod](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/go.mod)

Update module name to `github.com/brendreyes/pdforge` and add dependencies:
- `github.com/pdfcpu/pdfcpu` — PDF processing (merge, split, compress, image→PDF)
- `github.com/spf13/cobra` — CLI framework

---

### Project Entry Point

#### [NEW] [main.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/main.go)

Minimal entry point that calls `cmd.Execute()`.

---

### CLI Commands (`cmd/`)

#### [NEW] [root.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/cmd/root.go)

Root Cobra command with app description, version flag, and ASCII banner.

#### [NEW] [merge.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/cmd/merge.go)

`pdforge merge <file1.pdf> [file2.pdf ...] -o <output.pdf>`
- Requires 2+ input files
- `-o` / `--output` flag for destination (default: `merged.pdf`)

#### [NEW] [split.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/cmd/split.go)

`pdforge split <file.pdf> -p <pages> -o <output_dir>`
- `-p` / `--pages` flag: page specification (e.g., `1-3,5,8-10`)
- `-o` / `--output` flag: output directory (default: current directory)

#### [NEW] [compress.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/cmd/compress.go)

`pdforge compress <file.pdf> -o <output.pdf>`
- Shows before/after file sizes
- `-o` / `--output` flag (default: `<filename>_compressed.pdf`)

#### [NEW] [convert.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/cmd/convert.go)

`pdforge convert <image1> [image2 ...] -o <output.pdf>`
- Accepts `.jpg`, `.jpeg`, `.png`, `.tiff`, `.bmp`
- `-o` / `--output` flag (default: `converted.pdf`)

---

### Core PDF Logic (`internal/pdf/`)

#### [NEW] [merge.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/internal/pdf/merge.go)

`MergeFiles(inputPaths []string, outputPath string) error` — wraps `pdfcpu.MergeCreateFile()`.

#### [NEW] [split.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/internal/pdf/split.go)

`SplitByPages(inputPath, outputDir string, pages string) error` — parses page specs, extracts pages via `pdfcpu`.

#### [NEW] [compress.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/internal/pdf/compress.go)

`CompressFile(inputPath, outputPath string) error` — wraps `pdfcpu.OptimizeFile()`.

#### [NEW] [convert.go](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/internal/pdf/convert.go)

`ImagesToPDF(imagePaths []string, outputPath string) error` — converts images to PDF via `pdfcpu.ImportImagesFile()`.

---

### Documentation

#### [NEW] [README.md](file:///c:/Users/Wayne%20Karlo/Documents/pdforge/README.md)

Project overview, installation instructions, usage examples for all 4 commands, and privacy mission statement.

---

## Verification Plan

### Automated CLI Testing

Run the built binary against real files to verify all 4 commands:

```bash
# Build
go build -o pdforge.exe .

# 1. Merge — combine 2 PDFs
pdforge.exe merge testdata/a.pdf testdata/b.pdf -o testdata/merged.pdf
# ✓ merged.pdf exists and is a valid PDF

# 2. Split — extract pages
pdforge.exe split testdata/merged.pdf -p 1 -o testdata/split_out
# ✓ output file(s) created in split_out/

# 3. Compress — optimize PDF
pdforge.exe compress testdata/a.pdf -o testdata/compressed.pdf
# ✓ compressed.pdf exists, file size ≤ original

# 4. Convert — images to PDF
pdforge.exe convert testdata/sample.jpg -o testdata/from_image.pdf
# ✓ from_image.pdf exists and is a valid PDF
```

I'll create small test PDF and image files in `testdata/` to run these tests.

### Manual Verification

After I run the automated tests above, I'll ask you to try the commands yourself on any PDF files you have locally to confirm everything works on your machine.

