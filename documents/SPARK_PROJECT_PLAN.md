# 🔥 SPARK Project Plan — pdforge

> **A local, open-source PDF toolkit that keeps your data yours.**

---

## S — Scope & Story

### Pain Point

- Users are **forced to upload sensitive PDFs** to online tools (iLovePDF, SmallPDF, etc.) just to merge, split, compress, or convert files
- **Privacy risk** — uploaded documents may be stored, logged, or intercepted
- No simple, free, offline tool that handles all common PDF operations in one place

### Target User

- Privacy-conscious individuals (lawyers, accountants, students, researchers)
- Developers and power users comfortable with CLI tools
- Anyone working with sensitive or confidential documents

### MVP Core Value

- **Process PDFs 100% locally** — no internet required, no file uploads, no data leaves your machine
- One command-line tool for the 4 most common PDF operations

### ⚠ Out-of-Scope (for MVP)

- ❌ Web dashboard UI (future phase)
- ❌ PDF editing (text/annotation manipulation)
- ❌ OCR / text extraction
- ❌ PDF form filling
- ❌ Digital signatures
- ❌ Cloud sync or multi-device support
- ❌ Batch processing via config files (future)

---

## P — Plan & Product

### Tech Stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| **Language** | Go | Fast compilation, single binary distribution, cross-platform |
| **PDF Engine** | `pdfcpu` (Apache 2.0) | Fully open-source, handles merge/split/compress/convert natively |
| **CLI Framework** | `cobra` | Industry standard for Go CLIs (used by Docker, kubectl, Hugo) |
| **Future Web UI** | TBD (MagicUI / Shadcn / HeroUI) | Decided later when web dashboard phase begins |

### Feature List

#### ✅ Must-Have (MVP)

- **Merge** — Combine 2+ PDF files into one
- **Split** — Extract specific pages or page ranges from a PDF
- **Compress** — Optimize PDF file size
- **Convert** — Convert images (JPG, PNG, TIFF, BMP) to PDF

#### 🔮 Future

- PDF → Image conversion
- Rotate / reorder pages
- Watermarking
- Password protection / encryption
- Web dashboard with drag-and-drop UI
- Batch processing via YAML/JSON config

### Architecture

```
pdforge/
├── main.go                  # Entry point
├── cmd/                     # CLI commands (Cobra)
│   ├── root.go              # Root command + version
│   ├── merge.go             # pdforge merge
│   ├── split.go             # pdforge split
│   ├── compress.go          # pdforge compress
│   └── convert.go           # pdforge convert
└── internal/pdf/            # Core logic (wraps pdfcpu)
    ├── merge.go
    ├── split.go
    ├── compress.go
    └── convert.go
```

---

## A — Allocation

### Roles

| Member | Role | Responsibilities |
|--------|------|------------------|
| Developer | Full-stack / Lead | Architecture, core PDF logic, CLI commands, testing, documentation |

### Milestones

| Week | Milestone | Deliverable |
|------|-----------|-------------|
| **Week 1** | Project Setup | Go project scaffolded, dependencies installed, directory structure created |
| **Week 1** | Core Merge & Split | `pdforge merge` and `pdforge split` working end-to-end |
| **Week 2** | Core Compress & Convert | `pdforge compress` and `pdforge convert` working end-to-end |
| **Week 2** | Polish & Ship | README, error handling, cross-platform binary builds |
| **Future** | Web Dashboard | Browser-based UI wrapping the CLI functionality |

---

## R — Risk

### Technical Risks

| Risk | Impact | Likelihood |
|------|--------|------------|
| `pdfcpu` doesn't support a specific PDF version | Some PDFs fail to process | Low — pdfcpu supports all PDF versions through 2.0 |
| Large PDFs cause memory issues | Slow or crashed processing | Medium — mitigated by pdfcpu's efficient streaming |
| Image conversion quality loss | Poor output PDFs | Low — pdfcpu handles image import natively |

### Schedule Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Feature creep (adding too many features) | MVP never ships | Strict out-of-scope list enforced |
| Unfamiliar with pdfcpu API | Slower development | Library has excellent docs + examples |

### Solutions (Plan B)

- **If `pdfcpu` can't handle a specific feature** → Fall back to `unipdf` (AGPL) or shell out to `ghostscript`
- **If running behind schedule** → Cut Convert (images→PDF) from MVP; ship with Merge, Split, Compress only
- **If CLI is too complex** → Simplify to positional args only, remove optional flags

---

## K — Key Metrics

### Definition of Done ✅

- [ ] All 4 commands (`merge`, `split`, `compress`, `convert`) work end-to-end
- [ ] Single binary — no external dependencies required at runtime
- [ ] Works on Windows, macOS, and Linux
- [ ] All file processing is 100% local (zero network calls)
- [ ] README with clear usage instructions
- [ ] Code is on GitHub

### Success Numbers

| Metric | Target |
|--------|--------|
| Commands working | 4/4 |
| Binary size | < 20 MB |
| Merge 10 PDFs | < 5 seconds |
| Compress a 50MB PDF | < 10 seconds |
| Zero external API calls | ✓ |

### Future Value

- Open-source contribution to the privacy-first tooling ecosystem
- Foundation for a full web dashboard product
- Portfolio piece demonstrating Go systems programming
- Potential to grow into a comprehensive document processing suite
co
