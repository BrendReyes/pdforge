package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMerge(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		output    string
		dir       string
		expectErr bool
	}{
		{
			name: "valid merge two pdfs",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			expectErr: false,
		},
		{
			name: "valid merge four pdfs",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
				"testdata/pdfs/sample3.pdf",
				"testdata/pdfs/sample4.pdf",
			},
			expectErr: false,
		},
		{
			name: "invalid file type in args",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/images/sample.jpg",
			},
			expectErr: true,
		},
		{
			name: "file does not exist",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/doesnotexist.pdf",
			},
			expectErr: true,
		},
		{
			name: "uppercase PDF extension",
			args: []string{
				"testdata/pdfs/sample5.PDF",
				"testdata/pdfs/sample2.pdf",
			},
			expectErr: false,
		},
		{
			name: "duplicate input files",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample1.pdf",
			},
			expectErr: false,
		},
		{
			name: "invalid output extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			output: "test.txt",
			expectErr: true,
		},
		{
			name: "valid custom output directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name: "non existing directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			dir:       "testdata/fakedir",
			expectErr: true,
		},
		{
			name: "dir flag points to a file not directory",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			dir:       "testdata/pdfs/sample1.pdf",
			expectErr: true,
		},
		{
			name: "custom valid output name",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			output:    "myresult.pdf",
			expectErr: false,
		},
		{
			name: "output name with spaces",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			output:    "my result.pdf",
			expectErr: false,
		},
		{
			name: "output name with no extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			output:    "result",
			expectErr: true,
		},
		{
			name: "output name with wrong extension",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"testdata/pdfs/sample2.pdf",
			},
			output:    "result.docx",
			expectErr: true,
		},
		{
			name: "empty string in args",
			args: []string{
				"testdata/pdfs/sample1.pdf",
				"",
			},
			expectErr: true,
		},
		{
			name: "file with spaces in name",
			args: []string{
				"testdata/pdfs/sample 1.pdf",
				"testdata/pdfs/sample 2.pdf",
			},
			expectErr: false,
		},
		{
			name: "file with special characters in name",
			args: []string{
				"testdata/pdfs/sample(1).pdf",
				"testdata/pdfs/sample(2).pdf",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := mergeCmd
			cmd.ResetFlags()

			currentDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}

			// use tt.dir if provided, otherwise default to currentDir
			dir := currentDir
			if tt.dir != "" {
				dir = tt.dir
			}

			cmd.Flags().StringP("output", "o", "", "output")
			cmd.Flags().StringP("dir", "d", dir, "directory")

			// set custom output if provided
			if tt.output != "" {
				if err := cmd.Flags().Set("output", tt.output); err != nil {
					t.Fatalf("failed to set output flag: %v", err)
				}
			}

			err = runMerge(cmd, tt.args)

			// cleanup generated output files
			t.Cleanup(func() {
				entries, _ := os.ReadDir(currentDir)
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), "merged_") ||
						entry.Name() == tt.output {
						os.Remove(filepath.Join(currentDir, entry.Name()))
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