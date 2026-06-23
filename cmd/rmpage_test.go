package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRmpage(t *testing.T) {
	// ensure testdata/output exists before tests run
	os.MkdirAll("testdata/output", 0o755)

	tests := []struct {
		name      string
		args      []string
		page      string
		output    string
		dir       string
		expectErr bool
	}{
		// --- happy path ---
		{
			name:      "valid remove single page positional",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			expectErr: false,
		},
		{
			name:      "valid remove single page via --page flag",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			page:      "8",
			expectErr: false,
		},
		{
			name:      "valid remove range positional",
			args:      []string{"testdata/pdfs/sample1.pdf", "1-3"},
			expectErr: false,
		},
		{
			name:      "valid remove combination positional",
			args:      []string{"testdata/pdfs/sample1.pdf", "1,6-11,17"},
			expectErr: false,
		},
		{
			name:      "valid remove with custom output name",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			output:    "myremoved.pdf",
			expectErr: false,
		},
		{
			name:      "valid remove with custom output directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name:      "uppercase PDF extension input",
			args:      []string{"testdata/pdfs/sample5.PDF", "1"},
			expectErr: false,
		},
		{
			name:      "pdf with spaces in name",
			args:      []string{"testdata/pdfs/sample 1.pdf", "1"},
			expectErr: false,
		},
		{
			name:      "pdf with parentheses in name",
			args:      []string{"testdata/pdfs/sample(1).pdf", "1"},
			expectErr: false,
		},
		{
			name:      "output name with spaces",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			output:    "my removed.pdf",
			expectErr: false,
		},
		{
			name:      "output uppercase PDF extension",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			output:    "removed_upper.PDF",
			expectErr: false,
		},

		// --- selector handling ---
		{
			name:      "missing selector no positional no flag",
			args:      []string{"testdata/pdfs/sample1.pdf"},
			expectErr: true,
		},
		{
			name:      "both positional and --page flag provided",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			page:      "8",
			expectErr: true,
		},
		{
			name:      "invalid page specification",
			args:      []string{"testdata/pdfs/sample1.pdf", "abc"},
			expectErr: true,
		},

		// --- invalid input file ---
		{
			name:      "input file is not a pdf",
			args:      []string{"testdata/images/sample.jpg", "1"},
			expectErr: true,
		},
		{
			name:      "input file has no extension",
			args:      []string{"testdata/pdfs/sample", "1"},
			expectErr: true,
		},
		{
			name:      "input file does not exist",
			args:      []string{"testdata/pdfs/doesnotexist.pdf", "1"},
			expectErr: true,
		},
		{
			name:      "empty string as input",
			args:      []string{"", "1"},
			expectErr: true,
		},

		// --- invalid output ---
		{
			name:      "output name with no extension",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			output:    "result",
			expectErr: true,
		},
		{
			name:      "output name with wrong extension",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			output:    "result.docx",
			expectErr: true,
		},

		// --- directory ---
		{
			name:      "non existing directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			dir:       "testdata/fakedir",
			expectErr: true,
		},
		{
			name:      "dir flag points to a file not directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "8"},
			dir:       "testdata/pdfs/sample1.pdf",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rmpageCmd
			cmd.ResetFlags()

			// Mock stdin to auto-answer 'N' to the promptYesNo directory creation
			cmd.SetIn(strings.NewReader("N\n"))

			cmd.Flags().StringP("page", "p", "", "page")
			cmd.Flags().StringP("output", "o", "", "output")
			cmd.Flags().StringP("dir", "d", "", "directory")

			if tt.page != "" {
				if err := cmd.Flags().Set("page", tt.page); err != nil {
					t.Fatalf("failed to set page flag: %v", err)
				}
			}
			if tt.dir != "" {
				if err := cmd.Flags().Set("dir", tt.dir); err != nil {
					t.Fatalf("failed to set dir flag: %v", err)
				}
			}
			if tt.output != "" {
				if err := cmd.Flags().Set("output", tt.output); err != nil {
					t.Fatalf("failed to set output flag: %v", err)
				}
			}

			err := cmd.RunE(cmd, tt.args)

			// cleanup generated output files
			t.Cleanup(func() {
				outDir := tt.dir
				if outDir == "" && len(tt.args) > 0 && tt.args[0] != "" {
					outDir = filepath.Dir(tt.args[0])
				}
				if outDir == "" {
					outDir = "."
				}

				entries, _ := os.ReadDir(outDir)
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), "removed_") || entry.Name() == filepath.Base(tt.output) {
						os.Remove(filepath.Join(outDir, entry.Name()))
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
