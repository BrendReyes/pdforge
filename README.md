# pdforge

```text
 ________  ________  ________ ________  ________  ________  _______      
|\   __  \|\   ___ \|\  _____\\   __  \|\   __  \|\   ____\|\  ___ \     
\ \  \|\  \ \  \_|\ \ \  \__/\ \  \|\  \ \  \|\  \ \  \___|\ \   __/|    
 \ \   ____\ \  \ \\ \ \   __\\ \  \\\  \ \   _  _\ \  \  __\ \  \_|/__  
  \ \  \___|\ \  \_\\ \ \  \_| \ \  \\\  \ \  \\  \\ \  \|\  \ \  \_|\ \ 
   \ \__\    \ \_______\ \__\   \ \_______\ \__\\ _\\ \_______\ \_______\
    \|__|     \|_______|\|__|    \|_______|\|__|\|__|\|_______|\|_______|  
                                                                                                                                                                                                                
		⠀⠀⠀⠀⠀⠀⠀⢰⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⡄⠀⠀⠀⠀⠀
		⠀⠹⣿⣿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢠⣄⡀⠀⠀
		⠀⠀⠙⢿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣿⣿⡶⠀
		⠀⠀⠀⠀⠉⠛⠇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠸⠟⠋⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠸⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠇⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣶⣶⣶⣶⣶⣶⣶⣶⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣷⡀⠀⠀⠀⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣄⠀⠀⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⣀⣀⣈⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣁⣀⣀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀
```

**pdforge** is a local, privacy-first PDF toolkit for your terminal. Designed for speed and security, all processing happens entirely on your machine—no uploads, no cloud dependencies, and no compromises on privacy.

## Basic tools

- **Convert**: Transform image sets (JPG, PNG, WEBP, TIFF) into a single PDF document.
- **Merge**: Combine multiple PDF files into one, preserving the order you specify.
- **Split**: Break a PDF into pieces by a boundary or extract specific page ranges/segments.
- **Remove Page**: Delete specific pages or ranges from an existing PDF.
- **Optimize**: Compress and optimize PDF structure to reduce file size.

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
Supports two splitting modes:

**Split by boundary**:
Creates two files: one from 1 to N, and another from N+1 to end.
```bash
pdforge split document.pdf 5
```

**Extract segments**:
Extract specific pages or ranges into their own separate files.
```bash
pdforge split -e document.pdf 1-3,5,10-12
```

### Remove Pages
Remove specific pages or ranges from a PDF.

```bash
pdforge rmpage document.pdf 2,4-6
```

### Optimize PDF
Reduce the file size of a PDF while keeping it usable.

```bash
pdforge optimize large_file.pdf -o optimized.pdf
```

## Global Flags

- `-o, --output`: Specify a custom output filename.
- `-d, --dir`: Specify a custom output directory (defaults to the input file's directory).

## Privacy & Security
This project is purely local and offline, which means privacy is certain. 

## Notes

This project still has limited features but will be implemented in the future:
1. password input: protected files won't work since there is now password feature for now
2. rotate files command
3. simple friendly GUI: this is work in progress 

Thanks to the existing library [pdfcpu](https://github.com/pdfcpu/pdfcpu) for making this simple project easier to make.

This project is not or near perfect, there may be still some couple of bugs undiscovered. 

