![alt text goes here](https://github.com/BrendReyes/pdforge/actions/workflows/ci.yml/badge.svg)


![alt text](assets/pdforge.gif)

**pdforge** is a local, privacy-first PDF toolkit for your terminal. Designed for speed and security, all processing happens entirely on your machine—no uploads, no cloud dependencies, and no compromises on privacy.

## Basic tools

- **Convert**: Transform image sets (JPG, PNG, WEBP, TIFF) into a single PDF document.
- **Merge**: Combine multiple PDF files into one, preserving the order you specify.
- **Split**: Break a PDF into pieces by a boundary or extract specific page ranges/segments.
- **Remove Page**: Delete specific pages or ranges from an existing PDF.
- **Rotate**: Rotate every page—or only selected pages—by a multiple of 90 degrees.
- **Optimize**: Compress and optimize PDF structure to reduce file size.

Password-protected PDFs are supported across the PDF commands via the `-P, --password` flag (the output stays protected).

## Installation

### Using Go
If you have Go installed on your system, you can install `pdforge` directly:

```bash
go install github.com/brendreyes/pdforge@latest
```
### Global Access (Adding to PATH)
To use `pdforge` from any directory, ensure your Go bin directory is in your system's PATH.

#### Linux and macOS
Add this line to your shell profile (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.bash_profile`):
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
Then, restart your terminal or run `source ~/.zshrc` (replace with your shell's config file).

#### Windows
1. Open **Start Search**, type in "env", and select "Edit the system environment variables".
2. Click **Environment Variables**.
3. Under **User variables**, find `Path` and click **Edit**.
4. Click **New** and add the output of running `go env GOPATH` followed by `\bin` (usually `%USERPROFILE%\go\bin`).
5. Click **OK** and restart your terminal.

## How to use

### Convert
Combine one or more images into a single PDF file. Supported formats: `JPG`, `PNG`, `WEBP`, `TIFF or TIF`.

```bash
pdforge convert page1.jpg page2.png -o document.pdf
```

### Merge
Merge multiple PDF files into a single document.

```bash
pdforge merge report_part1.pdf report_part2.pdf -o full_report.pdf
```

### Split or Extract Pages
Supports two splitting modes. The selector can be passed positionally or via `-p, --page`.

**Split by boundary**:
Creates two files: one from 1 to N, and another from N+1 to end.
```bash
pdforge split document.pdf 5
```

**Extract segments**:
Extract specific pages or ranges into their own separate files with `-e, --extract`.
```bash
pdforge split -e document.pdf 1-3,5,10-12
```

In extract mode you can keep only odd or even pages within the selected segments, and print per-output page details:
```bash
pdforge split -e document.pdf 1-10 --odd
pdforge split -e document.pdf 1-10 --even -v
```

### Remove Pages
Remove specific pages or ranges from a PDF.

```bash
pdforge rmpage document.pdf 2,4-6
```

### Rotate Pages
Rotate pages clockwise by a non-zero multiple of 90 degrees. Use a negative value to rotate counter-clockwise, and `-p, --page` to rotate only selected pages (defaults to all pages).

```bash
pdforge rotate document.pdf 90
pdforge rotate document.pdf -90 -o rotated.pdf
pdforge rotate document.pdf 180 --page 1,3-5
```

### Optimize PDF
Reduce the file size of a PDF while keeping it usable.

```bash
pdforge optimize large_file.pdf -o optimized.pdf
```

### Password-Protected PDFs
Pass the password with `-P, --password` to read encrypted PDFs. Works with `merge`, `split`, `rmpage`, `rotate`, and `optimize`. The output remains protected with the same password.

```bash
pdforge rotate secret.pdf 90 -P mypassword
pdforge merge a.pdf b.pdf -P mypassword -o combined.pdf
```

## Common Flags

- `-o, --output`: Specify a custom output filename (or prefix when a command produces multiple files).
- `-d, --dir`: Specify a custom output directory (defaults to the input file's directory).
- `-P, --password`: Password for protected PDFs (available on `merge`, `split`, `rmpage`, `rotate`, and `optimize`).

## Privacy & Security
This project is purely local and offline, which means privacy is certain. 

## Current Limitations

A few rough edges and missing features to be aware of:

- **One password per command**: the `-P, --password` flag applies a single password to all input files. You can't merge two protected PDFs that use *different* passwords in the same command.
- **No encrypt/decrypt command**: there's no way to add, change, or strip a password (yet...maybe implement it later? idk).
- **Password is passed as a flag**: it's visible in your shell history and process list, and there's no interactive/hidden prompt.
- **Convert is image-only**: `convert` builds PDFs from images and does not support password-protected output, plus there's no way to change the size of image, image aligning, etc., just image straight to pdf. 
- **Optimize is structure-focused**: it optimizes PDF structure and text streams, so image-heavy PDFs see limited size savings (reduce image resolution beforehand for best results). It also handles one file at a time.

## Notes

This project still has limited features but more will be implemented in the future:
1. simple friendly GUI: this is work in progress 

Thanks to the existing library [pdfcpu](https://github.com/pdfcpu/pdfcpu) for making this simple project easier to make.
Though, the pdfcpu may already have similar cli tool like this, but we still continue for school project, and learning purposes :D 

This project is not or near perfect, there may be still some couple of bugs undiscovered. 

