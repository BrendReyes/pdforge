package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSplit(t *testing.T) {
	// ensure testdata/output exists before tests run
	os.MkdirAll("testdata/output", 0o755)

	tests := []struct {
		name      string
		args      []string
		output    string
		dir       string
		page      string
		extract   bool
		odd       bool
		even      bool
		verbose   bool
		expectErr bool
	}{
		// --- happy path: basic split ---
		{
			name:      "valid split at page boundary positional",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			expectErr: false,
		},
		{
			name:      "valid split using --page flag",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "3",
			expectErr: false,
		},
		{
			name:      "valid split with custom output prefix",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			output:    "mysplit",
			expectErr: false,
		},
		{
			name:      "valid split with custom output directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name:      "valid split with custom output prefix and directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			output:    "mysplit",
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name:      "valid split with verbose flag",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			verbose:   true,
			expectErr: false,
		},

		// --- happy path: extract mode ---
		{
			name:      "valid extract single page",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1",
			extract:   true,
			expectErr: false,
		},
		{
			name:      "valid extract page range",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-3",
			extract:   true,
			expectErr: false,
		},
		{
			name:      "valid extract multiple segments",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1,3,5",
			extract:   true,
			expectErr: false,
		},
		{
			name:      "valid extract mixed segments and ranges",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1,3-5",
			extract:   true,
			expectErr: false,
		},
		{
			name:      "valid extract odd pages",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-5",
			extract:   true,
			odd:       true,
			expectErr: false,
		},
		{
			name:      "valid extract even pages",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-5",
			extract:   true,
			even:      true,
			expectErr: false,
		},
		{
			name:      "valid extract with custom output directory",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-3",
			extract:   true,
			dir:       "testdata/output",
			expectErr: false,
		},

		// --- invalid input file ---
		{
			name:      "input file does not exist",
			args:      []string{"testdata/pdfs/doesnotexist.pdf", "3"},
			expectErr: true,
		},
		{
			name:      "input file is not a pdf",
			args:      []string{"testdata/images/sample.jpg", "3"},
			expectErr: true,
		},
		{
			name:      "input file has no extension",
			args:      []string{"testdata/pdfs/sample", "3"},
			expectErr: true,
		},
		{
			name:      "empty string as input",
			args:      []string{"", "3"},
			expectErr: true,
		},
		{
			name:      "input file with wrong extension",
			args:      []string{"testdata/pdfs/sample.docx", "3"},
			expectErr: true,
		},

		// --- invalid selector ---
		{
			name:      "missing selector no positional no flag",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			expectErr: true,
		},
		{
			name:      "both positional and --page flag provided",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			page:      "3",
			expectErr: true,
		},
		{
			name:      "selector is zero",
			args:      []string{"testdata/pdfs/sample1.pdf", "0"},
			expectErr: true,
		},
		{
			name:      "selector is negative",
			args:      []string{"testdata/pdfs/sample1.pdf", "-1"},
			expectErr: true,
		},
		{
			name:      "selector exceeds page count",
			args:      []string{"testdata/pdfs/sample1.pdf", "9999"},
			expectErr: true,
		},
		{
			name:      "selector equals page count",
			args:      []string{"testdata/pdfs/sample1.pdf", "45"},
			expectErr: true,
		},
		{
			name:      "selector is not a number",
			args:      []string{"testdata/pdfs/sample1.pdf", "abc"},
			expectErr: true,
		},
		{
			name:      "selector is float",
			args:      []string{"testdata/pdfs/sample1.pdf", "3.5"},
			expectErr: true,
		},
		{
			name:      "wrong argument order selector before pdf",
			args:      []string{"3", "testdata/pdfs/sample1.pdf"},
			expectErr: true,
		},

		// --- invalid extract selectors ---
		{
			name:      "extract with invalid range start greater than end",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "5-2",
			extract:   true,
			expectErr: true,
		},
		{
			name:      "extract with page out of bounds",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "9999",
			extract:   true,
			expectErr: true,
		},
		{
			name:      "extract with empty segment in selector",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1,,3",
			extract:   true,
			expectErr: true,
		},
		{
			name:      "extract with non numeric selector",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "abc",
			extract:   true,
			expectErr: true,
		},
		{
			name:      "extract with invalid range format",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-2-3",
			extract:   true,
			expectErr: true,
		},

		// --- invalid flag combinations ---
		{
			name:      "odd and even flags together",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-5",
			extract:   true,
			odd:       true,
			even:      true,
			expectErr: true,
		},
		{
			name:      "odd without extract",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			odd:       true,
			expectErr: true,
		},
		{
			name:      "even without extract",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			even:      true,
			expectErr: true,
		},

		// --- invalid output ---
		{
			name:      "output with wrong extension",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1-3",
			extract:   true,
			output:    "result.txt",
			expectErr: true,
		},
		{
			name:      "non existing output directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			dir:       "testdata/fakedir",
			expectErr: true,
		},
		{
			name:      "dir flag points to a file not directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			dir:       "testdata/pdfs/sample1.pdf",
			expectErr: true,
		},
		{
			name:      "output directory with file name",
			args:      []string{"testdata/pdfs/sample1.pdf", "3"},
			output:       "testdata/pdfs/output/split_success.pdf",
			expectErr: false,
		},

		// --- edge cases ---
		{
			name:      "split at page 1",
			args:      []string{"testdata/pdfs/sample1.pdf", "1"},
			expectErr: false,
		},
		{
			name:      "pdf with spaces in name",
			args:      []string{"testdata/pdfs/sample 1.pdf", "3"},
			expectErr: false,
		},
		{
			name:      "pdf with parentheses in name",
			args:      []string{"testdata/pdfs/sample(1).pdf", "3"},
			expectErr: false,
		},
		{
			name:      "extract single page as output file",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "1",
			extract:   true,
			output:    "page1.pdf",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset package-level flags before each test
			splitExtract = tt.extract
			splitOdd = tt.odd
			splitEven = tt.even
			splitVerbose = tt.verbose
			splitPage = tt.page
			splitOutput = tt.output
			splitDir = tt.dir

			cmd := splitCmd
			cmd.ResetFlags()

			// Mock stdin to auto-answer 'N' to the promptYesNo directory creation
			cmd.SetIn(strings.NewReader("N\n"))

			currentDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}

			// re-register flags after reset
			cmd.Flags().BoolVarP(&splitExtract, "extract", "e", tt.extract, "")
			cmd.Flags().BoolVar(&splitOdd, "odd", tt.odd, "")
			cmd.Flags().BoolVarP(&splitEven, "even", "n", tt.even, "")
			cmd.Flags().BoolVarP(&splitVerbose, "verbose", "v", tt.verbose, "")
			cmd.Flags().StringVarP(&splitPage, "page", "p", tt.page, "")
			cmd.Flags().StringVarP(&splitOutput, "output", "o", tt.output, "")
			cmd.Flags().StringVarP(&splitDir, "dir", "d", tt.dir, "")

			err = runSplit(cmd, tt.args)

			// cleanup generated split files
			t.Cleanup(func() {
				// clean currentDir
				entries, _ := os.ReadDir(currentDir)
				for _, entry := range entries {
					if (strings.HasPrefix(entry.Name(), "split_") ||
						strings.HasPrefix(entry.Name(), "mysplit_") ||
						strings.HasSuffix(entry.Name(), ".pdf")) &&
						strings.HasSuffix(entry.Name(), ".pdf") {
						os.Remove(filepath.Join(currentDir, entry.Name()))
					}
				}
				// clean custom dir if used
				if tt.dir != "" {
					entries, _ := os.ReadDir(tt.dir)
					for _, entry := range entries {
						if strings.HasSuffix(entry.Name(), ".pdf") {
							os.Remove(filepath.Join(tt.dir, entry.Name()))
						}
					}
				}
				// clean input file directory
				if len(tt.args) > 0 && tt.args[0] != "" {
					inputDir := filepath.Dir(tt.args[0])
					if inputDir != "." && inputDir != currentDir {
						entries, _ := os.ReadDir(inputDir)
						for _, entry := range entries {
							if (strings.HasPrefix(entry.Name(), "split_") ||
								strings.HasPrefix(entry.Name(), "mysplit_")) &&
								strings.HasSuffix(entry.Name(), ".pdf") {
								os.Remove(filepath.Join(inputDir, entry.Name()))
							}
						}
					}
				}
			})

			if tt.expectErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}