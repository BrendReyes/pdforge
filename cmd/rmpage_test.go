package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePageSpecification(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
		wantErr  bool
		errMsg   string
	}{
		// Basic valid cases
		{
			name:     "single page",
			input:    "8",
			expected: []int{8},
		},
		{
			name:     "range of pages",
			input:    "1-18",
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18},
		},
		{
			name:     "combination of single and range",
			input:    "1,6-11,17",
			expected: []int{1, 6, 7, 8, 9, 10, 11, 17},
		},
		{
			name:     "reverse order input",
			input:    "17,1,6-11",
			expected: []int{1, 6, 7, 8, 9, 10, 11, 17},
		},
		{
			name:     "all duplicates",
			input:    "5,5,5",
			expected: []int{5},
		},

		// Whitespace handling
		{
			name:     "spaces around commas",
			input:    "1 , 5 , 10",
			expected: []int{1, 5, 10},
		},
		{
			name:     "spaces around range",
			input:    "1 - 5",
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "tabs and spaces mixed",
			input:    "1,  2  -  4  , 7",
			expected: []int{1, 2, 3, 4, 7},
		},

		// Single element ranges
		{
			name:     "single element range",
			input:    "5-5",
			expected: []int{5},
		},
		{
			name:     "single element range with spaces",
			input:    "10 - 10",
			expected: []int{10},
		},

		// Large numbers
		{
			name:     "large page numbers",
			input:    "1000-1005",
			expected: []int{1000, 1001, 1002, 1003, 1004, 1005},
		},
		{
			name:     "mix of small and large numbers",
			input:    "1,5,1000,1002-1004",
			expected: []int{1, 5, 1000, 1002, 1003, 1004},
		},

		// Complex combinations
		{
			name:     "complex multi-part specification",
			input:    "1,3-5,7,10-15,20",
			expected: []int{1, 3, 4, 5, 7, 10, 11, 12, 13, 14, 15, 20},
		},
		{
			name:     "overlapping ranges",
			input:    "1-5,3-8",
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:     "completely overlapping ranges",
			input:    "1-10,1-10",
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},

		// Error cases
		{
			name:    "multiple dashes in range",
			input:   "1-5-10",
			wantErr: true,
		},
		{
			name:    "non-numeric input",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "inverted range",
			input:   "10-5",
			wantErr: true,
		},
		{
			name:    "negative numbers",
			input:   "-5",
			wantErr: true,
		},
		{
			name:    "negative in range",
			input:   "-5-10",
			wantErr: true,
		},
		{
			name:    "zero page number",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "zero in range",
			input:   "0-5",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only comma",
			input:   ",",
			wantErr: true,
		},
		{
			name:    "only dashes",
			input:   "-",
			wantErr: true,
		},
		{
			name:    "leading comma",
			input:   ",1,2",
			wantErr: true,
		},
		{
			name:    "trailing comma",
			input:   "1,2,",
			wantErr: true,
		},
		{
			name:    "consecutive commas",
			input:   "1,,2",
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid numbers",
			input:   "1,abc,5",
			wantErr: true,
		},
		{
			name:    "floating point",
			input:   "1.5,2.3",
			wantErr: true,
		},
		{
			name:    "hexadecimal",
			input:   "0xFF",
			wantErr: true,
		},
		{
			name:    "very large range that inverts",
			input:   "9999-1",
			wantErr: true,
		},
		{
			name:    "text mixed with numbers",
			input:   "page5",
			wantErr: true,
		},
		{
			name:    "special characters",
			input:   "1,2@5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePageSpecification(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePageSpecification(%q)\n  error = %v\n  wantErr = %v", tt.input, err, tt.wantErr)
				return
			}

			if err != nil {
				return // Expected error case
			}

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("parsePageSpecification(%q)\n  got length %d\n  want length %d\n  got %v\n  want %v",
					tt.input, len(result), len(tt.expected), result, tt.expected)
				return
			}

			// Check values
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parsePageSpecification(%q)\n  got %v\n  want %v",
						tt.input, result, tt.expected)
					return
				}
			}
		})
	}
}

func TestResolveOutputPathForRmPage(t *testing.T) {
	t.Run("no conflict returns original", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "remove.pdf")

		got := resolveOutputPath(path)
		if got != path {
			t.Fatalf("resolveOutputPath() = %s, want %s", got, path)
		}
	})

	t.Run("single conflict returns (1)", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "remove.pdf")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("failed creating fixture file: %v", err)
		}

		got := resolveOutputPath(path)
		want := filepath.Join(dir, "remove (1).pdf")
		if got != want {
			t.Fatalf("resolveOutputPath() = %s, want %s", got, want)
		}
	})

	t.Run("multiple conflicts returns next index", func(t *testing.T) {
		dir := t.TempDir()
		fixtures := []string{
			filepath.Join(dir, "report.pdf"),
			filepath.Join(dir, "report (1).pdf"),
			filepath.Join(dir, "report (2).pdf"),
		}

		for _, f := range fixtures {
			if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
				t.Fatalf("failed creating fixture file %s: %v", f, err)
			}
		}

		got := resolveOutputPath(filepath.Join(dir, "report.pdf"))
		want := filepath.Join(dir, "report (3).pdf")
		if got != want {
			t.Fatalf("resolveOutputPath() = %s, want %s", got, want)
		}
	})
}
