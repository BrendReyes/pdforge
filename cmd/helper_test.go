package cmd

import (
    "os"
    "testing"
)

func TestResolveOutputPath(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() // create files to simulate conflicts
        input    string
        expected string
        cleanup  func() // remove files after test
    }{
        {
            name:     "no conflict, returns original path",
            setup:    func() {},
            input:    "test_merged.pdf",
            expected: "test_merged.pdf",
            cleanup:  func() {},
        },
        {
			name: "single conflict, returns (1)",
			setup: func() {
				os.Create("test_conflict.pdf")
		},
			input:    "test_conflict.pdf",
			expected: "test_conflict (1).pdf",
			cleanup: func() {
				os.Remove("test_conflict.pdf")
				os.Remove("test_conflict (1).pdf")
		},
	},
		{
			name: "multiple conflicts, returns next available index",
			setup: func() {
				os.Create("multi.pdf")
				os.Create("multi (1).pdf")
				os.Create("multi (2).pdf")
		},
			input:    "multi.pdf",
			expected: "multi (3).pdf",
			cleanup: func() {
				os.Remove("multi.pdf")
				os.Remove("multi (1).pdf")
				os.Remove("multi (2).pdf")
				os.Remove("multi (3).pdf")
		},
	},
	{
			name: "file with multiple dots in name",
			setup: func() {
				os.Create("my.report.v1.pdf")
		},
			input:    "my.report.v1.pdf",
			expected: "my.report.v1 (1).pdf",
			cleanup: func() {
				os.Remove("my.report.v1.pdf")
				os.Remove("my.report.v1 (1).pdf")
		},
	},
	{
			name: "file without extension",
			setup: func() {
				os.Create("README")
		},
			input:    "README",
			expected: "README (1)",
			cleanup: func() {
				os.Remove("README")
				os.Remove("README (1)")
		},
	},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            defer tt.cleanup()

            result := resolveOutputPath(tt.input)
            if result != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result)
            }
        })
    }
}