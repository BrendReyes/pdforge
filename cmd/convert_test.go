package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunConvert(t *testing.T) {
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
			name: "valid single jpg",
			args: []string{
				"testdata/images/sample.jpg",
			},
			expectErr: false,
		},
		{
			name: "valid single png",
			args: []string{
				"testdata/images/sample.png",
			},
			expectErr: false,
		},
		{
			name: "valid single webp",
			args: []string{
				"testdata/images/sample.webp",
			},
			expectErr: false,
		},
		{
			name: "valid single tiff",
			args: []string{
				"testdata/images/sample.tiff",
			},
			expectErr: false,
		},
		{
			name: "valid multiple images mixed types",
			args: []string{
				"testdata/images/sample.jpg",
				"testdata/images/sample.png",
				"testdata/images/sample.webp",
			},
			expectErr: false,
		},
		{
			name: "valid multiple jpg images",
			args: []string{
				"testdata/images/sample.jpg",
				"testdata/images/sample2.jpg",
			},
			expectErr: false,
		},

		// --- output name ---
		{
			name: "custom valid output name",
			args: []string{
				"testdata/images/sample.jpg",
			},
			output:    "myconvert.pdf",
			expectErr: false,
		},
		{
			name: "output name with spaces",
			args: []string{
				"testdata/images/sample.jpg",
			},
			output:    "my convert.pdf",
			expectErr: false,
		},
		{
			name: "output name with no extension",
			args: []string{
				"testdata/images/sample.jpg",
			},
			output:    "result",
			expectErr: true,
		},
		{
			name: "output name with wrong extension",
			args: []string{
				"testdata/images/sample.jpg",
			},
			output:    "result.docx",
			expectErr: true,
		},
		{
			name: "output name uppercase PDF extension",
			args: []string{
				"testdata/images/sample.jpg",
			},
			output:    "result.PDF",
			expectErr: false,
		},

		// --- invalid file types ---
		{
			name: "unsupported gif type",
			args: []string{
				"testdata/images/sample.gif",
			},
			expectErr: true,
		},
		{
			name: "unsupported bmp type",
			args: []string{
				"testdata/images/sample.bmp",
			},
			expectErr: true,
		},
		{
			name: "pdf file passed as image",
			args: []string{
				"testdata/pdfs/sample1.pdf",
			},
			expectErr: true,
		},
		{
			name: "mixed valid and invalid types",
			args: []string{
				"testdata/images/sample.jpg",
				"testdata/images/sample.gif",
			},
			expectErr: true,
		},
		{
			name: "uppercase valid extension JPG",
			args: []string{
				"testdata/images/sample1.JPG",
			},
			expectErr: false,
		},
		{
			name: "uppercase valid extension PNG",
			args: []string{
				"testdata/images/sample1.PNG",
			},
			expectErr: false,
		},

		// --- file existence ---
		{
			name: "file does not exist",
			args: []string{
				"testdata/images/ghost.jpg",
			},
			expectErr: true,
		},
		{
			name: "empty string in args",
			args: []string{
				"",
			},
			expectErr: true,
		},
		{
			name: "file with spaces in name",
			args: []string{
				"testdata/images/sample image.jpg",
			},
			expectErr: false,
		},
		{
			name: "file with special characters in name",
			args: []string{
				"testdata/images/sample(1).jpg",
			},
			expectErr: false,
		},

		// --- directory ---
		{
			name: "valid custom output directory",
			args: []string{
				"testdata/images/sample.jpg",
			},
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name: "non existing directory",
			args: []string{
				"testdata/images/sample.jpg",
			},
			dir:       "testdata/fakedir",
			expectErr: true,
		},
		{
			name: "dir flag points to a file not directory",
			args: []string{
				"testdata/images/sample.jpg",
			},
			dir:       "testdata/images/sample.jpg",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := convertCmd
			cmd.ResetFlags()

			currentDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}

			dir := currentDir
			if tt.dir != "" {
				dir = tt.dir
			}

			cmd.Flags().StringP("output", "o", "", "output")
			cmd.Flags().StringP("dir", "d", dir, "directory")

			if tt.output != "" {
				if err := cmd.Flags().Set("output", tt.output); err != nil {
					t.Fatalf("failed to set output flag: %v", err)
				}
			}

			err = runConvert(cmd, tt.args)

			// cleanup generated output files
			t.Cleanup(func() {
				// clean currentDir
				entries, _ := os.ReadDir(currentDir)
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), "convert_") ||
						entry.Name() == tt.output {
						os.Remove(filepath.Join(currentDir, entry.Name()))
					}
				}
				// clean custom dir if used
				if tt.dir != "" {
					entries, _ := os.ReadDir(tt.dir)
					for _, entry := range entries {
						if strings.HasPrefix(entry.Name(), "convert_") ||
							entry.Name() == tt.output {
							os.Remove(filepath.Join(tt.dir, entry.Name()))
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