package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOptimize(t *testing.T) {
	os.MkdirAll("testdata/output", 0o755)

	tests := []struct {
		name      string
		args      []string
		output    string
		dir       string
		expectErr bool
	}{
		// --- happy path ---
		{
			name: "valid single pdf",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			expectErr: false,
		},
		{
			name: "valid pdf with custom output name",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "myoptimized.pdf",
			expectErr: false,
		},
		{
			name: "valid pdf with custom output directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name: "valid pdf with custom output name and directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "myoptimized.pdf",
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name: "valid pdf with spaces in name",
			args: []string{
				"testdata/pdfs/sample 1.pdf",
			},
			expectErr: false,
		},
		{
			name: "valid pdf with special characters in name",
			args: []string{
				"testdata/pdfs/sample(1).pdf",
			},
			expectErr: false,
		},
		{
			name: "output name with spaces",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "my optimized.pdf",
			expectErr: false,
		},

		// --- invalid input file type ---
		{
			name: "image file passed instead of pdf",
			args: []string{
				"testdata/images/sample.jpg",
			},
			expectErr: true,
		},
		{
			name: "png file passed instead of pdf",
			args: []string{
				"testdata/images/sample.png",
			},
			expectErr: true,
		},
		{
			name: "no extension file",
			args: []string{
				"testdata/pdfs/sample",
			},
			expectErr: true,
		},
		{
			name: "uppercase PDF extension input",
			args: []string{
				"testdata/pdfs/sample5.PDF",
			},
			expectErr: false,
		},
		{
			name: "docx file passed instead of pdf",
			args: []string{
				"testdata/pdfs/sample.docx",
			},
			expectErr: true,
		},

		// --- file existence ---
		{
			name: "file does not exist",
			args: []string{
				"testdata/pdfs/doesnotexist.pdf",
			},
			expectErr: true,
		},
		{
			name: "empty string as input",
			args: []string{
				"",
			},
			expectErr: true,
		},

		// --- invalid output ---
		{
			name: "output name with no extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "result",
			expectErr: true,
		},
		{
			name: "output name with wrong extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "result.docx",
			expectErr: true,
		},
		{
			name: "output name with txt extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "result.txt",
			expectErr: true,
		},
		{
			name: "output uppercase PDF extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "result.PDF",
			expectErr: false,
		},

		// --- directory ---
		{
			name: "non existing directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			dir:       "testdata/fakedir",
			expectErr: true,
		},
		{
			name: "dir flag points to a file not directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			dir:       "testdata/pdfs/sample1.pdf",
			expectErr: true,
		},

		// -- output flag with directory --
		{
			name: "output with directory but without name",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "testdata/output",
			expectErr: true,
		},
		{
			name: "output with directory with name",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			output:    "testdata/output/optimize_success.pdf",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := optimizeCmd
			cmd.ResetFlags()

			// Mock stdin to auto-answer 'N' to the promptYesNo directory creation
			cmd.SetIn(strings.NewReader("N\n"))

			cmd.Flags().StringP("output", "o", "", "output")
			cmd.Flags().StringP("dir", "d", "", "directory")

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

			err := runOptimize(cmd, tt.args)

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
					if (strings.HasPrefix(entry.Name(), "optimized_") ||
						entry.Name() == tt.output) &&
						strings.HasSuffix(entry.Name(), ".pdf") {
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