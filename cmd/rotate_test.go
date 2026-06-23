package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRotate(t *testing.T) {
	os.MkdirAll("testdata/output", 0o755)

	tests := []struct {
		name      string
		args      []string
		page      string
		output    string
		dir       string
		password  string
		expectErr bool
	}{
		// --- happy path ---
		{
			name:      "rotate all pages clockwise 90",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			expectErr: false,
		},
		{
			name:      "rotate all pages 180",
			args:      []string{"testdata/pdfs/sample1.pdf", "180"},
			expectErr: false,
		},
		{
			name:      "rotate counter-clockwise negative 90",
			args:      []string{"testdata/pdfs/sample1.pdf", "-90"},
			expectErr: false,
		},
		{
			name:      "rotate selected pages via --page",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			page:      "1,3-5",
			expectErr: false,
		},
		{
			name:      "rotate with custom output name",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			output:    "myrotated.pdf",
			expectErr: false,
		},
		{
			name:      "rotate with custom output directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			dir:       "testdata/output",
			expectErr: false,
		},
		{
			name:      "uppercase PDF extension input",
			args:      []string{"testdata/pdfs/sample5.PDF", "90"},
			expectErr: false,
		},
		{
			name:      "pdf with spaces in name",
			args:      []string{"testdata/pdfs/sample 1.pdf", "90"},
			expectErr: false,
		},

		// --- invalid rotation ---
		{
			name:      "rotation not a multiple of 90",
			args:      []string{"testdata/pdfs/sample1.pdf", "45"},
			expectErr: true,
		},
		{
			name:      "rotation is zero",
			args:      []string{"testdata/pdfs/sample1.pdf", "0"},
			expectErr: true,
		},
		{
			name:      "rotation is not a number",
			args:      []string{"testdata/pdfs/sample1.pdf", "abc"},
			expectErr: true,
		},

		// --- invalid input ---
		{
			name:      "input is not a pdf",
			args:      []string{"testdata/images/sample.jpg", "90"},
			expectErr: true,
		},
		{
			name:      "input does not exist",
			args:      []string{"testdata/pdfs/doesnotexist.pdf", "90"},
			expectErr: true,
		},
		{
			name:      "input has no extension",
			args:      []string{"testdata/pdfs/sample", "90"},
			expectErr: true,
		},

		// --- invalid page selection ---
		{
			name:      "invalid page specification",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			page:      "abc",
			expectErr: true,
		},

		// --- invalid output ---
		{
			name:      "output with no extension",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			output:    "result",
			expectErr: true,
		},
		{
			name:      "output with wrong extension",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			output:    "result.docx",
			expectErr: true,
		},

		// --- directory ---
		{
			name:      "non existing directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			dir:       "testdata/fakedir",
			expectErr: true,
		},
		{
			name:      "dir flag points to a file not directory",
			args:      []string{"testdata/pdfs/sample1.pdf", "90"},
			dir:       "testdata/pdfs/sample1.pdf",
			expectErr: true,
		},

		// --- password protected ---
		{
			name:      "protected pdf with correct password",
			args:      []string{"testdata/pdfs/sample_protected.pdf", "90"},
			password:  "samplefiles",
			expectErr: false,
		},
		{
			name:      "protected pdf without password",
			args:      []string{"testdata/pdfs/sample_protected.pdf", "90"},
			expectErr: true,
		},
		{
			name:      "protected pdf with wrong password",
			args:      []string{"testdata/pdfs/sample_protected.pdf", "90"},
			password:  "wrongpw",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rotateCmd
			cmd.ResetFlags()

			// Mock stdin to auto-answer 'N' to the promptYesNo directory creation
			cmd.SetIn(strings.NewReader("N\n"))

			cmd.Flags().StringP("page", "p", "", "page")
			cmd.Flags().StringP("output", "o", "", "output")
			cmd.Flags().StringP("dir", "d", "", "directory")
			cmd.Flags().StringP("password", "P", "", "password")

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

			if tt.password != "" {
				if err := cmd.Flags().Set("password", tt.password); err != nil {
					t.Fatalf("failed to set password flag: %v", err)
				}
			}

			err := runRotate(cmd, tt.args)

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
					if strings.HasPrefix(entry.Name(), "rotated_") || entry.Name() == filepath.Base(tt.output) {
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
